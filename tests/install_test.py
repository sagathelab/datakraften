import hashlib
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
INSTALL = ROOT / 'install'


class InstallTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.home = self.root / 'home'
        self.home.mkdir()
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.mock('id', 'echo 1000')
        self.mock('uname', 'case "$1" in -s) echo "${TEST_OS}";; -m) echo "${TEST_ARCH}";; *) echo "${TEST_OS}";; esac')
        self.mock('curl', '''
case "$*" in
  *api.github.com*) printf '%s\\n' "${TEST_RELEASE}" ;;
  *.sha256*) printf '%s  dk\\n' "${TEST_HASH}" ;;
  *) while [ "$#" -gt 0 ]; do
       if [ "$1" = -o ]; then shift; printf '%s' binary > "$1"; exit; fi
       shift
     done ;;
esac
''')
        self.env = dict(os.environ, HOME=str(self.home), PATH=f'{self.bin}:/usr/bin:/bin',
                        SHELL='/bin/bash', TEST_OS='Linux', TEST_ARCH='x86_64',
                        TEST_RELEASE='{"tag_name":"v1.2.3"}',
                        TEST_HASH=hashlib.sha256(b'binary').hexdigest(), TMPDIR=str(self.root))

    def mock(self, name, body):
        path = self.bin / name
        path.write_text('#!/bin/bash\n' + body + '\n')
        path.chmod(0o755)

    def run_install(self):
        return subprocess.run(['bash', str(INSTALL)], env=self.env, capture_output=True, text=True)

    def test_release_platforms_and_idempotent_path(self):
        for os_name, arch, asset in [('Linux', 'x86_64', 'dk-linux-amd64'),
                                    ('Linux', 'aarch64', 'dk-linux-arm64'),
                                    ('Darwin', 'x86_64', 'dk-darwin-amd64'),
                                    ('Darwin', 'arm64', 'dk-darwin-arm64')]:
            with self.subTest(os=os_name, arch=arch):
                self.env.update(TEST_OS=os_name, TEST_ARCH=arch, SHELL='/bin/zsh')
                result = self.run_install()
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertIn(asset, result.stdout)
                self.assertIn('Checksum verified', result.stdout)
                self.assertEqual((self.home / '.local/bin/dk').read_text(), 'binary')
                self.assertEqual((self.home / '.zshrc').read_text().count('# >>> datakraften >>>'), 1)
                self.assertFalse(list(self.root.glob('datakraften.*')))

    def test_checksum_mismatch_preserves_existing_binary(self):
        target = self.home / '.local/bin/dk'
        target.parent.mkdir(parents=True)
        target.write_text('existing')
        self.env['TEST_HASH'] = '0' * 64
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(target.read_text(), 'existing')
        self.assertFalse(list(self.root.glob('datakraften.*')))

    def test_source_fallback_sets_path(self):
        self.env['TEST_RELEASE'] = '{}'
        self.mock('git', 'if [ "$1" = clone ]; then mkdir -p "$5/.git"; fi')
        self.mock('go', 'mkdir -p bin; printf source > bin/dk')
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual((self.home / '.local/bin/dk').read_text(), 'source')
        self.assertIn('# >>> datakraften >>>', (self.home / '.profile').read_text())

    def test_install_copies_match(self):
        self.assertEqual(INSTALL.read_bytes(), (ROOT / 'site/public/install').read_bytes())


if __name__ == '__main__':
    unittest.main()
