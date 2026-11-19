import json
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


def project_env(home, **overrides):
    return dict(HOME=str(home), PATH='/usr/bin:/bin', XDG_CONFIG_HOME=str(home / 'config'),
                XDG_DATA_HOME=str(home / 'data'), **overrides)


class ProjectTest(unittest.TestCase):
    def test_nvm_project_uses_configured_manager_without_changing_default(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp).resolve()
            home = root / 'home'
            nvm = home / 'my nvm'
            (nvm / 'alias').mkdir(parents=True)
            (nvm / 'alias' / 'default').write_text('original\n')
            managed = home / 'data' / 'datakraften' / 'nvm'
            startup = home / 'bash-startup'
            startup.write_text('touch "$HOME/bash-startup-ran"\n')
            (nvm / 'nvm.sh').write_text('''
printf 'sourced\\n' >> "$HOME/nvm-sourced"
[ "$1" = --no-use ] || return 2
nvm() {
  [ "$1" = install ] && [ "$2" = --no-progress ] && [ "$3" = 24 ] || return 3
  [ "$NVM_DIR" = "$EXPECTED_NVM_DIR" ] && [ "$PWD" = "$EXPECTED_NVM_DIR" ] || return 4
  [ -z "$PREFIX$npm_config_prefix$NVM_INSTALL_THIRD_PARTY_HOOK" ] || return 5
  mkdir -p "$NVM_DIR/alias"
  printf '%s\\n' '24' > "$NVM_DIR/alias/default"
  mkdir -p "$NVM_DIR/versions/node/v24.1.0/bin"
  printf '#!/bin/sh\\necho v24.1.0\\n' > "$NVM_DIR/versions/node/v24.1.0/bin/node"
  chmod 755 "$NVM_DIR/versions/node/v24.1.0/bin/node"
}
''')
            repo = root / 'repo'
            (repo / '.git').mkdir(parents=True)
            (repo / 'datakraften.yaml').write_text('''version: 1
project:
  tools:
    node:
      version: "24"
      manager: nvm
  tasks:
    version: node --version
''')
            env = project_env(home, NVM_DIR=str(nvm), EXPECTED_NVM_DIR=str(managed),
                              BASH_ENV=str(startup), PREFIX=str(home / 'global-node'),
                              npm_config_prefix=str(home / 'global-node'),
                              NVM_INSTALL_THIRD_PARTY_HOOK='unexpected',
                              NVM_NODEJS_ORG_MIRROR='https://invalid.example/private-token')
            mock_fnm = root / 'fnm'
            mock_fnm.write_text('#!/bin/sh\ntouch "$HOME/fnm-ran"\nexit 9\n')
            mock_fnm.chmod(0o755)
            env['PATH'] = str(root) + ':' + env['PATH']

            def dk(*args):
                return subprocess.run([binary, *args], cwd=repo, env=env, stdin=subprocess.DEVNULL,
                                      capture_output=True, text=True, timeout=15)

            status = dk('project', 'status', '--json')
            self.assertEqual(status.returncode, 0, status.stderr)
            data = json.loads(status.stdout)
            self.assertEqual(data['node_manager'], 'nvm')
            self.assertEqual(data['node_manager_path'], str(nvm / 'nvm.sh'))
            self.assertFalse(data['trusted'])
            self.assertIn('not installed', data['node_status'])

            preview = dk('project', 'setup', '--dry-run', '--yes')
            self.assertEqual(preview.returncode, 0, preview.stderr)
            self.assertIn(str(managed), preview.stdout)
            self.assertIn('NVM_NODEJS_ORG_MIRROR', preview.stdout)
            self.assertNotIn('private-token', preview.stdout + preview.stderr)
            preview = dk('run', 'version', '--dry-run')
            self.assertEqual(preview.returncode, 0, preview.stderr)
            self.assertIn('24 (nvm)', preview.stderr)

            denied = dk('project', 'setup', '--yes')
            self.assertNotEqual(denied.returncode, 0)
            self.assertIn('not trusted', denied.stderr)
            self.assertFalse((home / 'nvm-sourced').exists())
            self.assertFalse((home / 'config').exists())
            self.assertFalse(managed.exists())

            self.assertEqual(dk('project', 'trust', '--yes').returncode, 0)
            denied = dk('project', 'setup')
            self.assertNotEqual(denied.returncode, 0)
            self.assertIn('use --yes', denied.stderr)
            self.assertFalse(managed.exists())
            setup = dk('project', 'setup', '--yes')
            self.assertEqual(setup.returncode, 0, setup.stderr)
            self.assertEqual(dk('project', 'setup', '--yes').returncode, 0)
            run = dk('run', 'version')
            self.assertEqual(run.returncode, 0, run.stderr)
            self.assertEqual(run.stdout.strip(), 'v24.1.0')
            self.assertFalse((home / 'bash-startup-ran').exists())
            self.assertFalse((home / 'fnm-ran').exists())
            self.assertEqual((home / 'nvm-sourced').read_text(), 'sourced\n')
            self.assertEqual((nvm / 'alias' / 'default').read_text(), 'original\n')
            self.assertFalse((nvm / 'versions').exists())

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
            env = project_env(home)
            preview = subprocess.run([binary, 'run', '--dry-run', 'test'], cwd=repo, env=env, stdin=subprocess.DEVNULL,
                                      text=True, capture_output=True, timeout=10)
            self.assertEqual(preview.returncode, 0, preview.stderr)
            self.assertFalse((repo / 'executed').exists())
            run = subprocess.run([binary, 'run', 'test'], cwd=repo, env=env, stdin=subprocess.DEVNULL,
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
                env = project_env(home)
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
            env = project_env(home)
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
            root = pathlib.Path(tmp).resolve()
            home = root / 'home'
            bin_dir = root / 'bin'
            home.mkdir()
            bin_dir.mkdir()
            fallback_node = bin_dir / 'node'
            fallback_node.write_text('#!/bin/sh\nprintf "global|%s\\n" "$PROJECT_VALUE"\n')
            fallback_node.chmod(0o755)
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
            env = project_env(home, PROJECT_VALUE='inherited')
            env['PATH'] = str(bin_dir) + ':' + env['PATH']
            repos = []
            for version in ('22', '24'):
                repo = root / ('repo-' + version)
                (repo / '.git').mkdir(parents=True)
                (repo / 'subdir').mkdir()
                tool = '"22"' if version == '22' else '{version: "24", manager: fnm}'
                (repo / 'datakraften.yaml').write_text(f'''version: 1
project:
  tools:
    node: {tool}
  env:
    PROJECT_VALUE: "project-{version}"
  tasks:
    test: node
    fail: exit 7
''')
                (repo / '.nvmrc').write_text('18\n')
                (repo / '.node-version').write_text('20\n')
                (repo / 'package.json').write_text('{"engines":{"node":"26"}}\n')
                (repo / 'mise.toml').write_text('[tools]\nnode = "30"\n')
                repos.append(repo)

            def dk(repo, *args):
                return subprocess.run([binary, *args], cwd=repo / 'subdir', env=env,
                                      stdin=subprocess.DEVNULL, text=True, capture_output=True, timeout=15)

            for repo in repos:
                denied = dk(repo, 'run', 'test')
                self.assertNotEqual(denied.returncode, 0)
                self.assertIn('not installed', denied.stderr)
                version = repo.name.removeprefix('repo-')
                self.assertFalse((home / 'data' / 'datakraften' / 'fnm' / 'node-versions' / ('v' + version + '.0.0')).exists())
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

            outside = subprocess.run(['node'], cwd=root, env=env, stdin=subprocess.DEVNULL,
                                     text=True, capture_output=True, timeout=10)
            self.assertEqual(outside.stdout.strip(), 'global|inherited')
            unknown = subprocess.run([binary, 'project', 'status'], cwd=root, env=env, stdin=subprocess.DEVNULL,
                                     text=True, capture_output=True, timeout=10)
            self.assertNotEqual(unknown.returncode, 0)
            self.assertIn('no datakraften.yaml found', unknown.stderr)

            config = repos[0] / 'datakraften.yaml'
            config.write_text(config.read_text().replace('project-22', 'changed'))
            untrusted = dk(repos[0], 'run', 'test')
            self.assertNotEqual(untrusted.returncode, 0)
            self.assertIn('not trusted', untrusted.stderr)
            still_trusted = dk(repos[1], 'run', 'test')
            self.assertEqual(still_trusted.returncode, 0, still_trusted.stderr)


if __name__ == '__main__':
    unittest.main()
