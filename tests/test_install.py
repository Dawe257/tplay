"""Offline installer integration tests: python3 -m unittest discover -s tests -v."""

import hashlib
import io
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "install.sh"


class InstallTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.home = self.root / "home with spaces"
        self.home.mkdir()
        self.downloads = self.root / "downloads"
        self.downloads.mkdir()
        self.tmp = self.root / "tmp"
        self.tmp.mkdir()
        self.env = dict(os.environ, HOME=str(self.home), TMPDIR=str(self.tmp),
                        PATH=f"{self.bin}:/usr/bin:/bin", FIXTURES=str(self.downloads),
                        TEST_OS="Linux", TEST_ARCH="x86_64")
        self.stub("uname", '#!/bin/sh\ncase "$1" in -s) echo "$TEST_OS";; -m) echo "$TEST_ARCH";; esac\n')
        self.stub("curl", '''#!/bin/sh
set -eu
out=
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o|--output) out=$2; shift 2;;
        -w|--write-out) shift 2;;
        https://*) url=$1; shift;;
        *) shift;;
    esac
done
case "$url" in
    https://github.com/dzhenetl/tplay/releases/latest)
        printf 'https://github.com/dzhenetl/tplay/releases/tag/v1.2.3';;
    https://github.com/dzhenetl/tplay/releases/download/v1.2.3/*)
        cp "$FIXTURES/${url##*/}" "$out";;
    *) exit 22;;
esac
''')
        self.archive()

    def stub(self, name, body):
        path = self.bin / name
        path.write_text(body)
        path.chmod(0o755)

    def archive(self, target="linux_amd64", corrupt=False):
        name = f"tplay_1.2.3_{target}.tar.gz"
        path = self.downloads / name
        with tarfile.open(path, "w:gz") as archive:
            payload = b"#!/bin/sh\necho installed\n"
            info = tarfile.TarInfo("tplay")
            info.size = len(payload)
            archive.addfile(info, io.BytesIO(payload))
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        (self.downloads / "checksums.txt").write_text(f"{digest}  {name}\n")
        if corrupt:
            path.write_bytes(b"corrupted download")

    def run_install(self):
        # Pass the script over stdin, exactly as curl | sh does.
        return subprocess.run(["/bin/sh"], input=SCRIPT.read_text(),
                              text=True, capture_output=True, env=self.env)

    def test_installs_verified_binary(self):
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        installed = self.home / ".local/bin/tplay"
        self.assertTrue(os.access(installed, os.X_OK))
        self.assertEqual(subprocess.check_output([installed], text=True), "installed\n")
        self.assertEqual(list(self.tmp.iterdir()), [])
        self.assertIn("PATH", result.stderr + result.stdout)

    def test_macos_arm64(self):
        self.env.update(TEST_OS="Darwin", TEST_ARCH="arm64")
        self.archive("darwin_arm64")
        self.assertEqual(self.run_install().returncode, 0)

    def test_linux_aarch64(self):
        self.env["TEST_ARCH"] = "aarch64"
        self.archive("linux_arm64")
        self.assertEqual(self.run_install().returncode, 0)

    def test_bad_checksum_preserves_existing_install(self):
        installed = self.home / ".local/bin/tplay"
        installed.parent.mkdir(parents=True)
        installed.write_text("old binary")
        self.archive(corrupt=True)
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("checksum", result.stderr.lower())
        self.assertEqual(installed.read_text(), "old binary")
        self.assertEqual(list(self.tmp.iterdir()), [])

    def test_missing_checksum_rejected(self):
        (self.downloads / "checksums.txt").write_text("0" * 64 + "  other.tar.gz\n")
        self.assertNotEqual(self.run_install().returncode, 0)
        self.assertFalse((self.home / ".local/bin/tplay").exists())

    def test_download_failure(self):
        (self.downloads / "checksums.txt").unlink()
        self.assertNotEqual(self.run_install().returncode, 0)
        self.assertEqual(list(self.tmp.iterdir()), [])

    def test_unsupported_platform(self):
        self.env["TEST_OS"] = "FreeBSD"
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unsupported", result.stderr)

    def test_unsupported_architecture(self):
        self.env["TEST_ARCH"] = "riscv64"
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unsupported architecture", result.stderr)

    def test_shasum_fallback_and_missing_ffmpeg_warning(self):
        # A controlled PATH exercises the macOS checksum tool and dependency warning.
        for tool in ("sh", "tar", "gzip", "awk", "mktemp", "chmod", "mv", "mkdir",
                     "rm", "cat", "cp", "shasum"):
            (self.bin / tool).symlink_to(shutil.which(tool))
        self.env["PATH"] = str(self.bin)
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("Warning: ffmpeg is missing", result.stderr)
        self.assertIn("Warning: ffprobe is missing", result.stderr)

    def test_invalid_archive_with_valid_checksum(self):
        path = self.downloads / "tplay_1.2.3_linux_amd64.tar.gz"
        path.write_bytes(b"invalid tar")
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        (self.downloads / "checksums.txt").write_text(f"{digest}  {path.name}\n")
        result = self.run_install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("extract", result.stderr)
        self.assertFalse((self.home / ".local/bin/tplay").exists())
        self.assertEqual(list(self.tmp.iterdir()), [])


if __name__ == "__main__":
    unittest.main()
