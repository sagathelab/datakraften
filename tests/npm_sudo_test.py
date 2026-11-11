import os
import fcntl
import pathlib
import pty
import select
import signal
import subprocess
import struct
import sys
import tempfile
import termios
import time
import unittest


def terminal_screen(data, width=120, height=50):
    rows = [[(' ', None) for _ in range(width)] for _ in range(height)]
    x = y = 0
    color = None
    wrap = False
    text = data.decode('utf-8', errors='replace')
    i = 0
    while i < len(text):
        c = text[i]
        i += 1
        if c == '\x1b' and i < len(text) and text[i] == ']':
            i += 1
            while i < len(text):
                if text[i] == '\a':
                    i += 1
                    break
                if text[i:i + 2] == '\x1b\\':
                    i += 2
                    break
                i += 1
            continue
        if c == '\x1b' and i < len(text) and text[i] == '[':
            i += 1
            start = i
            while i < len(text) and not ('@' <= text[i] <= '~'):
                i += 1
            if i >= len(text):
                break
            sequence, action = text[start:i], text[i]
            i += 1
            values = [int(v) if v.isdigit() else 0 for v in sequence.split(';')]
            n = values[0] or 1
            if action == 'm':
                for value in values:
                    if value == 0:
                        color = None
                    elif value in (31, 32, 37, 90):
                        color = value
            elif action == 'A':
                y = max(0, y - n)
                wrap = False
            elif action == 'B':
                y = min(height - 1, y + n)
                wrap = False
            elif action == 'C':
                x = min(width - 1, x + n)
                wrap = False
            elif action == 'G':
                x = min(width - 1, n - 1)
                wrap = False
            elif action == 'K':
                if sequence == '2':
                    rows[y] = [(' ', None) for _ in range(width)]
                else:
                    rows[y][x:] = [(' ', None) for _ in range(width - x)]
                wrap = False
            continue
        if c == '\r':
            x, wrap = 0, False
        elif c == '\n':
            if y == height - 1:
                rows.pop(0)
                rows.append([(' ', None) for _ in range(width)])
            else:
                y += 1
            wrap = False
        elif c >= ' ':
            if wrap:
                x = 0
                if y == height - 1:
                    rows.pop(0)
                    rows.append([(' ', None) for _ in range(width)])
                else:
                    y += 1
                wrap = False
            rows[y][x] = (c, color)
            if x == width - 1:
                wrap = True
            else:
                x += 1
    return rows


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
            while b'[y/N]: ' not in output:
                ready, _, _ = select.select([fd], [], [], 10)
                self.assertTrue(ready, output.decode(errors='replace'))
                output += os.read(fd, 4096)
            os.write(fd, b'y')
            while not output.endswith(b'y'):
                ready, _, _ = select.select([fd], [], [], 10)
                self.assertTrue(ready, output.decode(errors='replace'))
                output += os.read(fd, 4096)
            os.write(fd, b'\n')
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
            self.assertEqual(log.read_text().splitlines(), [str(npm), 'update', '-g', '--prefix', '/protected/prefix', '@openai/codex'])
            os.close(fd)

    def test_interactive_password_keeps_original_heading_and_updates_color(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        cases = ((0, 32, 'normal'), (0, 32, 'npm_spinner'), (0, 32, 'osc'), (0, 32, 'systemd'), (42, 31, 'normal'), (0, None, 'scroll'), (0, None, 'resize'), (130, None, 'interrupt'), (0, None, 'suspend'))
        for exit_code, expected_color, mode in cases:
            with self.subTest(exit_code=exit_code, mode=mode), tempfile.TemporaryDirectory() as tmp:
                root = pathlib.Path(tmp)
                home, bin_dir = root / 'home', root / 'bin'
                config = home / '.config' / 'datakraften'
                config.mkdir(parents=True)
                bin_dir.mkdir()
                (config / 'config.yaml').write_text('ai_tools:\n  codex:\n    enabled: true\n')
                if mode == 'systemd':
                    (config / 'config.yaml').write_text('ai_tools:\n  codex:\n    enabled: true\n  claude:\n    enabled: true\n')
                npm = bin_dir / 'npm'
                npm.write_text('''#!/bin/sh
case "$1" in
  list) echo '{"dependencies":{"@openai/codex":{},"@anthropic-ai/claude-code":{}}}';;
  config) echo /protected/prefix;;
  update) echo 'npm error code EACCES' >&2; exit 1;;
esac
''')
                sudo = bin_dir / 'sudo'
                sudo.write_text(f'''#!{sys.executable}
import fcntl, struct, sys, termios, time
if not sys.stdin.isatty() or not sys.stdout.isatty():
    sys.exit(90)
if {mode!r} == "systemd":
    sys.stdout.write("\\x1b]3008;start=0123456789abcdef0123456789abcdef;type=elevate;user=root;hostname=fedora;machineid=0123456789abcdef0123456789abcdef;bootid=0123456789abcdef0123456789abcdef;pid=12345;pidfdid=67890;comm=sudo;cwd=/tmp\\x1b\\\\")
    sys.stdout.flush()
original = termios.tcgetattr(0)
no_echo = termios.tcgetattr(0)
no_echo[3] &= ~termios.ECHO
termios.tcsetattr(0, termios.TCSANOW, no_echo)
try:
    sys.stdout.write("[sudo] Password: ")
    sys.stdout.flush()
    password = sys.stdin.readline()
finally:
    termios.tcsetattr(0, termios.TCSANOW, original)
if {mode!r} == "npm_spinner":
    for part in (b"\\x1b[", b"1G\\x1b[0K\\xe2", b"\\xa0\\x99", b"\\x1b[1G\\x1b[0K"):
        sys.stdout.buffer.write(part)
        sys.stdout.buffer.flush()
        time.sleep(0.03)
if {mode!r} == "osc":
    for part in (b"\\x1b]", b"0;npm update", b"\\x07", b"\\x1b]9;4;1;50\\x1b", b"\\\\"):
        sys.stdout.buffer.write(part)
        sys.stdout.buffer.flush()
        time.sleep(0.03)
sys.stdout.write("\\nchanged packages\\n")
if {mode!r} == "systemd":
    sys.stdout.write("\\x1b]3008;end=1234567890abcdef;exit=success;status=0\\x1b\\\\")
if {mode!r} == "resize":
    time.sleep(0.15)
    size = fcntl.ioctl(0, termios.TIOCGWINSZ, b"\\0" * 8)
    rows, cols, _, _ = struct.unpack('HHHH', size)
    sys.stdout.write("size %dx%d\\n" % (rows, cols))
if {mode!r} == "scroll":
    for i in range(70):
        sys.stdout.write("sudo line %d\\n" % i)
sys.stdout.flush()
sys.exit({exit_code} if password == "secret-test-password\\n" else 91)
''')
                npm.chmod(0o755)
                sudo.chmod(0o755)
                env = dict(os.environ, HOME=str(home), PATH=str(bin_dir), TERM='xterm')
                env.pop('NO_COLOR', None)
                pid, fd = pty.fork()
                if pid == 0:
                    rows, cols = (47, 212) if mode == 'systemd' else (50, 120)
                    fcntl.ioctl(1, termios.TIOCSWINSZ, struct.pack('HHHH', rows, cols, 0, 0))
                    os.execve(binary, [binary, 'update'] if mode == 'systemd' else [binary, 'update', 'npm'], env)
                output = b''
                sent_confirm = sent_password = False
                reaped = False
                try:
                    while True:
                        ready, _, _ = select.select([fd], [], [], 15)
                        if not ready:
                            os.kill(pid, 9)
                            self.fail('interactive sudo timed out: ' + output.decode(errors='replace'))
                        try:
                            chunk = os.read(fd, 65536)
                        except OSError:
                            break
                        if not chunk:
                            break
                        output += chunk
                        if not sent_confirm and b'[y/N]' in output:
                            screen = terminal_screen(output, width=212, height=47) if mode == 'systemd' else terminal_screen(output)
                            self.assertTrue(any('Updating npm' in ''.join(c for c, _ in row) for row in screen))
                            os.write(fd, b'y\n')
                            sent_confirm = True
                        if not sent_password and b'[sudo] Password:' in output:
                            screen = terminal_screen(output, width=212, height=47) if mode == 'systemd' else terminal_screen(output)
                            self.assertTrue(any('Updating npm' in ''.join(c for c, _ in row) for row in screen))
                            if mode == 'resize':
                                fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack('HHHH', 40, 80, 0, 0))
                            if mode == 'suspend':
                                os.write(fd, b'\x1a')
                                deadline = time.monotonic() + 10
                                while True:
                                    stopped_pid, stopped_status = os.waitpid(pid, os.WUNTRACED | os.WNOHANG)
                                    if stopped_pid and os.WIFSTOPPED(stopped_status):
                                        break
                                    if stopped_pid or time.monotonic() >= deadline:
                                        os.kill(pid, signal.SIGKILL)
                                        self.fail('process did not suspend at Ctrl-Z')
                                    time.sleep(0.01)
                                flags = termios.tcgetattr(fd)[3]
                                self.assertTrue(flags & termios.ECHO, f'terminal echo not restored during suspend: flags={flags:#x}, stop={os.WSTOPSIG(stopped_status)}, output={output!r}')
                                self.assertTrue(flags & termios.ICANON, f'terminal mode not restored during suspend: flags={flags:#x}, stop={os.WSTOPSIG(stopped_status)}')
                                os.kill(pid, signal.SIGCONT)
                                time.sleep(0.1)
                            os.write(fd, b'\x03' if mode == 'interrupt' else b'secret-test-password\n')
                            sent_password = True
                    _, status = os.waitpid(pid, 0)
                    reaped = True
                    if mode == 'interrupt':
                        flags = termios.tcgetattr(fd)[3]
                        self.assertTrue(flags & termios.ECHO, 'terminal echo was not restored')
                        self.assertTrue(flags & termios.ICANON, 'terminal canonical mode was not restored')
                finally:
                    if not reaped:
                        try:
                            os.kill(pid, signal.SIGKILL)
                            os.waitpid(pid, 0)
                        except (ProcessLookupError, ChildProcessError):
                            pass
                    os.close(fd)
                self.assertTrue(sent_confirm and sent_password, output.decode(errors='replace'))
                self.assertEqual(os.waitstatus_to_exitcode(status) == 0, exit_code == 0, output.decode(errors='replace'))
                self.assertNotIn(b'secret-test-password', output)
                if mode == 'interrupt':
                    continue
                self.assertIn(b'changed packages', output)
                if mode == 'scroll':
                    self.assertIn(b'sudo line 69', output)
                if mode == 'resize':
                    self.assertIn(b'size 40x80', output)
                if expected_color is None:
                    continue
                screen = terminal_screen(output, width=212, height=47) if mode == 'systemd' else terminal_screen(output)
                headings = [row for row in screen if ''.join(c for c, _ in row).strip() == '● Updating npm']
                self.assertEqual(len(headings), 1, output.decode(errors='replace'))
                dot = next(color for c, color in headings[0] if c == '●')
                self.assertEqual(dot, expected_color, output.decode(errors='replace'))

    def test_apply_system_sudo_preserves_progress_heading(self):
        binary = os.environ.get('DK_TEST_BINARY')
        if not binary:
            self.skipTest('set DK_TEST_BINARY to the built dk executable')
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            home, bin_dir = root / 'home', root / 'bin'
            config = home / '.config' / 'datakraften'
            config.mkdir(parents=True)
            bin_dir.mkdir()
            (config / 'config.yaml').write_text('source: default\n')
            scripts = {
                'apt-get': '#!/bin/sh\nexit 0\n',
                'dpkg': '#!/bin/sh\n[ "$2" = curl ] && exit 1\nprintf "ii %s installed\\n" "$2"\n',
                'brew': '#!/bin/sh\n[ "$1" = --prefix ] && echo /mock/brew\nexit 0\n',
                'fish': '#!/bin/sh\nexit 0\n',
                'sudo': f'''#!{sys.executable}
import sys, termios
if not sys.stdin.isatty() or not sys.stdout.isatty():
    sys.exit(90)
original = termios.tcgetattr(0)
no_echo = termios.tcgetattr(0)
no_echo[3] &= ~termios.ECHO
termios.tcsetattr(0, termios.TCSANOW, no_echo)
try:
    sys.stdout.write("[sudo] Password: ")
    sys.stdout.flush()
    password = sys.stdin.readline()
finally:
    termios.tcsetattr(0, termios.TCSANOW, original)
sys.stdout.write("\\nsystem command complete\\n")
sys.stdout.flush()
sys.exit(0 if password == "secret-test-password\\n" else 91)
''',
            }
            for name, script in scripts.items():
                executable = bin_dir / name
                executable.write_text(script)
                executable.chmod(0o755)
            env = dict(os.environ, HOME=str(home), PATH=str(bin_dir), TERM='xterm')
            env.pop('NO_COLOR', None)
            pid, fd = pty.fork()
            if pid == 0:
                fcntl.ioctl(1, termios.TIOCSWINSZ, struct.pack('HHHH', 80, 120, 0, 0))
                os.execve(binary, [binary, 'apply', '--yes'], env)
            output = b''
            prompts_answered = 0
            try:
                while True:
                    ready, _, _ = select.select([fd], [], [], 15)
                    if not ready:
                        os.kill(pid, 9)
                        self.fail('system sudo timed out: ' + output.decode(errors='replace'))
                    try:
                        chunk = os.read(fd, 65536)
                    except OSError:
                        break
                    if not chunk:
                        break
                    output += chunk
                    count = output.count(b'[sudo] Password:')
                    if count > prompts_answered:
                        os.write(fd, b'secret-test-password\n')
                        prompts_answered = count
                _, status = os.waitpid(pid, 0)
            finally:
                os.close(fd)
            self.assertEqual(os.waitstatus_to_exitcode(status), 0, output.decode(errors='replace'))
            self.assertEqual(prompts_answered, 2, output.decode(errors='replace'))
            self.assertNotIn(b'secret-test-password', output)
            screen = terminal_screen(output, height=80)
            headings = [row for row in screen if ''.join(c for c, _ in row).strip() == '● Checking system packages']
            self.assertEqual(len(headings), 1, output.decode(errors='replace'))
            dot = next(color for c, color in headings[0] if c == '●')
            self.assertEqual(dot, 32, output.decode(errors='replace'))


if __name__ == '__main__':
    unittest.main()
