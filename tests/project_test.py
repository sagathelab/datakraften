import os
import pathlib
import pty
import select
import shlex
import signal
import subprocess
import sys
import tempfile
import time
import unittest


class ProjectTest(unittest.TestCase):
    def test_dry_run_does_not_execute_or_trust(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            home = root / 'home'
            home.mkdir()
            repo = root / 'repo'
            (repo / '.git').mkdir(parents=True)
            (repo / 'datakraften.yaml').write_text('version: 1\nproject:\n  tasks:\n    test: touch executed\n')
            env = dict(os.environ, HOME=str(home), XDG_CONFIG_HOME=str(home / '.config'))
            preview = subprocess.run([binary, 'run', '--dry-run', 'test'], cwd=repo, env=env,
                                     text=True, capture_output=True, timeout=10)
            self.assertEqual(preview.returncode, 0, preview.stderr)
            self.assertFalse((repo / 'executed').exists())
            run = subprocess.run([binary, 'run', 'test'], cwd=repo, env=env,
                                 text=True, capture_output=True, timeout=10)
            self.assertNotEqual(run.returncode, 0)
            self.assertIn('not trusted', run.stderr)
            self.assertFalse((repo / 'executed').exists())

    def test_direct_signals_reach_task_and_children(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        for sig in (signal.SIGINT, signal.SIGTERM):
            with self.subTest(signal=sig), tempfile.TemporaryDirectory() as tmp:
                root = pathlib.Path(tmp)
                home = root / 'home'
                home.mkdir()
                repo = root / 'repo'
                (repo / '.git').mkdir(parents=True)
                script = root / 'worker.py'
                script.write_text('''import os, pathlib, signal, subprocess, sys, time
root = pathlib.Path(sys.argv[1])
role = sys.argv[2]
def handle(sig, frame):
    (root / (role + ".signal")).write_text(str(sig))
    sys.exit(128 + sig)
signal.signal(signal.SIGINT, handle)
signal.signal(signal.SIGTERM, handle)
if role == "parent":
    subprocess.Popen([sys.executable, __file__, str(root), "child"])
(root / (role + ".pid")).write_text(str(os.getpid()))
while True:
    time.sleep(1)
''')
                task = 'exec ' + ' '.join(shlex.quote(s) for s in (sys.executable, str(script), str(root), 'parent'))
                (repo / 'datakraften.yaml').write_text('version: 1\nproject:\n  tasks:\n    wait: ' + task + '\n')
                env = dict(os.environ, HOME=str(home), XDG_CONFIG_HOME=str(home / '.config'))
                trust = subprocess.run([binary, 'project', 'trust', '--yes'], cwd=repo, env=env,
                                       capture_output=True, text=True, timeout=10)
                self.assertEqual(trust.returncode, 0, trust.stderr)
                process = subprocess.Popen([binary, 'run', 'wait'], cwd=repo, env=env,
                                           stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                try:
                    deadline = time.monotonic() + 10
                    while not (root / 'child.pid').exists():
                        if process.poll() is not None or time.monotonic() > deadline:
                            self.fail('task did not start')
                        time.sleep(0.02)
                    process.send_signal(sig)
                    out, err = process.communicate(timeout=10)
                    self.assertEqual(process.returncode, 128 + sig, (out, err))
                    deadline = time.monotonic() + 5
                    while not (root / 'child.signal').exists() and time.monotonic() < deadline:
                        time.sleep(0.02)
                    self.assertEqual((root / 'parent.signal').read_text(), str(sig))
                    self.assertEqual((root / 'child.signal').read_text(), str(sig))
                finally:
                    if process.poll() is None:
                        process.kill()
                        process.communicate(timeout=5)
                    for name in ('parent.pid', 'child.pid'):
                        if (root / name).exists():
                            try:
                                os.kill(int((root / name).read_text()), signal.SIGKILL)
                            except ProcessLookupError:
                                pass

    def test_project_task_retains_interactive_terminal(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            home = root / 'home'
            home.mkdir()
            repo = root / 'repo'
            (repo / '.git').mkdir(parents=True)
            (repo / 'datakraften.yaml').write_text('''version: 1
project:
  tasks:
    interactive: 'test -t 0 && printf "input: " && read answer && printf "got %s\\n" "$answer"'
''')
            env = dict(os.environ, HOME=str(home), XDG_CONFIG_HOME=str(home / '.config'))
            trust = subprocess.run([binary, 'project', 'trust', '--yes'], cwd=repo, env=env,
                                   capture_output=True, text=True, timeout=10)
            self.assertEqual(trust.returncode, 0, trust.stderr)
            pid, fd = pty.fork()
            if pid == 0:
                os.chdir(repo)
                os.execve(binary, [binary, 'run', 'interactive'], env)
            output = b''
            sent = False
            reaped = False
            try:
                while True:
                    ready, _, _ = select.select([fd], [], [], 10)
                    if not ready:
                        self.fail('interactive task timed out: ' + output.decode(errors='replace'))
                    try:
                        chunk = os.read(fd, 4096)
                    except OSError:
                        break
                    if not chunk:
                        break
                    output += chunk
                    if not sent and b'input: ' in output:
                        os.write(fd, b'terminal-test\n')
                        sent = True
                _, status = os.waitpid(pid, 0)
                reaped = True
                self.assertEqual(os.waitstatus_to_exitcode(status), 0, output.decode(errors='replace'))
                self.assertTrue(sent)
                self.assertIn(b'got terminal-test', output)
            finally:
                if not reaped:
                    try:
                        os.kill(pid, signal.SIGKILL)
                        os.waitpid(pid, 0)
                    except (ProcessLookupError, ChildProcessError):
                        pass
                os.close(fd)

    def test_two_projects_use_independent_node_versions_and_trust(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            home = root / 'home'
            bin_dir = root / 'bin'
            home.mkdir()
            bin_dir.mkdir()
            fnm = bin_dir / 'fnm'
            fnm.write_text(f'''#!{sys.executable}
import os, pathlib, sys
if len(sys.argv) != 3 or sys.argv[1] != 'install':
    sys.exit(1)
version = sys.argv[2]
path = pathlib.Path(os.environ['FNM_DIR']) / 'node-versions' / ('v' + version + '.0.0') / 'installation' / 'bin'
path.mkdir(parents=True, exist_ok=True)
node = path / 'node'
node.write_text('#!/bin/sh\\nprintf "%s|%s|%s|%s\\\\n" "' + version + '" "$PROJECT_VALUE" "$PWD" "$1"\\n')
node.chmod(0o755)
''')
            fnm.chmod(0o755)
            env = dict(os.environ, HOME=str(home), XDG_CONFIG_HOME=str(home / '.config'),
                       XDG_DATA_HOME=str(home / '.local' / 'share'), PATH=str(bin_dir) + ':/usr/bin:/bin')
            env.pop('FNM_DIR', None)
            repos = []
            for version in ('22', '24'):
                repo = root / ('repo-' + version)
                (repo / '.git').mkdir(parents=True)
                (repo / 'subdir').mkdir()
                (repo / 'datakraften.yaml').write_text(f'''version: 1
project:
  tools:
    node: "{version}"
  env:
    PROJECT_VALUE: "project-{version}"
  tasks:
    test: node
    fail: exit 7
''')
                repos.append(repo)

            def dk(repo, *args):
                return subprocess.run([binary, *args], cwd=repo / 'subdir', env=env,
                                      text=True, capture_output=True, timeout=15)

            for repo in repos:
                denied = dk(repo, 'run', 'test')
                self.assertNotEqual(denied.returncode, 0)
                version = repo.name.removeprefix('repo-')
                self.assertFalse((home / '.local' / 'share' / 'datakraften' / 'fnm' / 'node-versions' / ('v' + version + '.0.0')).exists())
                status = dk(repo, 'project', 'status')
                self.assertEqual(status.returncode, 0, status.stderr)
                trust = dk(repo, 'project', 'trust', '--yes')
                self.assertEqual(trust.returncode, 0, trust.stderr)
                setup = dk(repo, 'project', 'setup', '--yes')
                self.assertEqual(setup.returncode, 0, setup.stderr)

            for repo, version in zip(repos, ('22', '24')):
                result = dk(repo, 'run', 'test', '--', 'argument with spaces')
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stdout.strip(), f'{version}|project-{version}|{repo.resolve()}|argument with spaces')
                failed = dk(repo, 'run', 'fail')
                self.assertEqual(failed.returncode, 7, failed.stderr)
                self.assertNotIn('Usage:', failed.stderr)

            config = repos[0] / 'datakraften.yaml'
            config.write_text(config.read_text().replace('project-22', 'changed'))
            untrusted = dk(repos[0], 'run', 'test')
            self.assertNotEqual(untrusted.returncode, 0)
            self.assertIn('not trusted', untrusted.stderr)
            still_trusted = dk(repos[1], 'run', 'test')
            self.assertEqual(still_trusted.returncode, 0, still_trusted.stderr)


if __name__ == '__main__':
    unittest.main()
