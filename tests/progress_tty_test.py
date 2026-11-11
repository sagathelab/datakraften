import fcntl
import os
import pathlib
import pty
import select
import struct
import tempfile
import termios
import unittest


class ProgressTTYTest(unittest.TestCase):
    def run_preview(self, config_text, no_color=False):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built executable')
        with tempfile.TemporaryDirectory() as home:
            config = pathlib.Path(home) / '.config/datakraften/config.yaml'
            config.parent.mkdir(parents=True)
            config.write_text(config_text)
            pid, fd = pty.fork()
            if pid == 0:
                os.environ.update(HOME=home, TERM='xterm')
                if no_color:
                    os.environ['NO_COLOR'] = '1'
                else:
                    os.environ.pop('NO_COLOR', None)
                os.execv(binary, [binary, 'apply', '--dry-run'])
            try:
                fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack('HHHH', 40, 100, 0, 0))
                output = b''
                while True:
                    ready, _, _ = select.select([fd], [], [], 15)
                    if not ready:
                        os.kill(pid, 9)
                        self.fail('progress command timed out')
                    try:
                        chunk = os.read(fd, 65536)
                    except OSError:
                        break
                    if not chunk:
                        break
                    output += chunk
                _, status = os.waitpid(pid, 0)
                return os.waitstatus_to_exitcode(status), output
            finally:
                os.close(fd)

    def test_apply_preview_updates_visible_heading(self):
        status, output = self.run_preview('source: default\n')
        self.assertEqual(status, 0)
        self.assertIn(b'\x1b[90m', output)
        self.assertIn(b'\x1b[32m', output)
        self.assertGreaterEqual(output.count(b'\x1b[2K'), 8)
        self.assertIn(b'Preview complete', output)

    def test_failure_turns_heading_red(self):
        status, output = self.run_preview('source: team\n')
        self.assertNotEqual(status, 0)
        self.assertIn(b'\x1b[90m', output)
        self.assertIn(b'\x1b[31m', output)
        self.assertIn(b'\x1b[2K', output)
        self.assertIn(b'Team source is active but no url is set', output)
        self.assertNotIn(b'Usage:', output)
        self.assertNotIn(b'Error: preview failed:', output)

    def test_no_color_uses_plain_text(self):
        status, output = self.run_preview('source: default\n', no_color=True)
        self.assertEqual(status, 0)
        self.assertNotIn(b'\x1b', output)
        self.assertIn(b'[...] Checking system packages', output)


if __name__ == '__main__':
    unittest.main()
