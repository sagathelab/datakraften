import os
import pathlib
import pty
import select
import tempfile
import unittest


class NpmSudoTest(unittest.TestCase):
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
            env = dict(os.environ, HOME=str(home), PATH=str(bin_dir), TEST_SUDO_LOG=str(log))
            pid, fd = pty.fork()
            if pid == 0:
                os.execve(binary, [binary, 'update', 'npm'], env)
            output = b''
            while b'[y/N]' not in output:
                ready, _, _ = select.select([fd], [], [], 10)
                self.assertTrue(ready, output.decode(errors='replace'))
                output += os.read(fd, 4096)
            os.write(fd, b'y\n')
            _, status = os.waitpid(pid, 0)
            self.assertEqual(os.waitstatus_to_exitcode(status), 0)
            self.assertEqual(log.read_text().splitlines(), [str(npm), 'update', '-g', '--prefix', '/protected/prefix', '@openai/codex'])
            os.close(fd)


if __name__ == '__main__':
    unittest.main()
