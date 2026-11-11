import os
import fcntl
import pathlib
import pty
import select
import subprocess
import struct
import tempfile
import termios
import unittest


class NpmSudoTest(unittest.TestCase):
    def test_apply_claude_permission_retry(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        for answer in ('y', 'n', None):
            with self.subTest(answer=answer), tempfile.TemporaryDirectory() as tmp:
                root = pathlib.Path(tmp)
                home, bin_dir = root / 'home', root / 'bin'
                config = home / '.config' / 'datakraften'
                config.mkdir(parents=True)
                bin_dir.mkdir()
                (config / 'config.yaml').write_text('source: default\nai_tools:\n  claude:\n    enabled: true\n')
                scripts = {
                    'brew': '#!/bin/sh\n[ "$1" = "--prefix" ] && echo /mock/brew\nexit 0\n',
                    'fish': '#!/bin/sh\nexit 0\n',
                    'npm': '#!/bin/sh\nif [ "$1" = config ]; then echo /protected/prefix; exit 0; fi\necho "npm error code EACCES" >&2\necho "stack trace should not leak" >&2\nexit 1\n',
                    'sudo': '#!/bin/sh\nprintf "%s\\n" "$@" > "$TEST_SUDO_LOG"\n',
                }
                for name, script in scripts.items():
                    executable = bin_dir / name
                    executable.write_text(script)
                    executable.chmod(0o755)
                log = root / 'sudo.log'
                env = dict(os.environ, HOME=str(home), PATH=str(bin_dir), TEST_SUDO_LOG=str(log), TERM='dumb')
                if answer is None:
                    result = subprocess.run([binary, 'apply', '--yes'], env=env, input=b'', capture_output=True, timeout=10)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn(b'no interactive terminal', result.stdout)
                    self.assertNotIn(b'stack trace should not leak', result.stdout + result.stderr)
                    self.assertFalse(log.exists())
                    continue
                pid, fd = pty.fork()
                if pid == 0:
                    os.execve(binary, [binary, 'apply', '--yes'], env)
                output = b''
                while b'[y/N]' not in output:
                    ready, _, _ = select.select([fd], [], [], 10)
                    self.assertTrue(ready, output.decode(errors='replace'))
                    output += os.read(fd, 65536)
                os.write(fd, (answer + '\n').encode())
                while True:
                    try:
                        chunk = os.read(fd, 65536)
                    except OSError:
                        break
                    if not chunk:
                        break
                    output += chunk
                _, status = os.waitpid(pid, 0)
                os.close(fd)
                self.assertNotIn(b'stack trace should not leak', output)
                self.assertNotIn(b'Usage:', output)
                if answer == 'y':
                    self.assertEqual(os.waitstatus_to_exitcode(status), 0, output.decode(errors='replace'))
                    self.assertEqual(log.read_text().splitlines(), [str(bin_dir / 'npm'), 'install', '-g', '--prefix', '/protected/prefix', '@anthropic-ai/claude-code'])
                else:
                    self.assertNotEqual(os.waitstatus_to_exitcode(status), 0)
                    self.assertFalse(log.exists())
                    self.assertEqual(output.count(b'sudo retry declined'), 1, output.decode(errors='replace'))
                    self.assertNotIn(b'Error: apply failed:', output)

    def test_permission_retry_uses_selected_packages_and_prefix(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            home, bin_dir = root / 'home', root / 'bin'
            config = home / '.config' / 'datakraften'
            config.mkdir(parents=True)
            bin_dir.mkdir()
            (config / 'config.yaml').write_text('ai_tools:\n  codex:\n    enabled: true\n')
            npm = bin_dir / 'npm'
            npm.write_text('''#!/bin/sh
case "$1" in
  list) echo '{"dependencies":{"@openai/codex":{},"other":{}}}';;
  config) echo /protected/prefix;;
  update) echo 'npm error code EACCES' >&2; exit 1;;
esac
''')
            sudo = bin_dir / 'sudo'
            sudo.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > "$TEST_SUDO_LOG"\n')
            npm.chmod(0o755)
            sudo.chmod(0o755)
            log = root / 'sudo.log'
            env = dict(os.environ, HOME=str(home), PATH=str(bin_dir), TEST_SUDO_LOG=str(log), TERM='xterm')
            env.pop('NO_COLOR', None)
            pid, fd = pty.fork()
            if pid == 0:
                fcntl.ioctl(1, termios.TIOCSWINSZ, struct.pack('HHHH', 40, 100, 0, 0))
                os.execve(binary, [binary, 'update', 'npm'], env)
            output = b''
            while b'[y/N]' not in output:
                ready, _, _ = select.select([fd], [], [], 10)
                self.assertTrue(ready, output.decode(errors='replace'))
                output += os.read(fd, 4096)
            os.write(fd, b'y\n')
            while True:
                try:
                    chunk = os.read(fd, 4096)
                except OSError:
                    break
                if not chunk:
                    break
                output += chunk
            _, status = os.waitpid(pid, 0)
            self.assertEqual(os.waitstatus_to_exitcode(status), 0)
            self.assertIn(b'\x1b[2K', output.split(b'[y/N]')[0])
            self.assertIn(b'\x1b[32m\xe2\x97\x8f\x1b[0m Updating npm', output)
            self.assertEqual(log.read_text().splitlines(), [str(npm), 'update', '-g', '--prefix', '/protected/prefix', '@openai/codex'])
            os.close(fd)


if __name__ == '__main__':
    unittest.main()
