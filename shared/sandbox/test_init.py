"""Run with: python3 -m unittest discover -s shared/sandbox -p 'test_*.py'."""

from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
INIT = (ROOT / "shared/sandbox/init.sh").read_text().replace(
    "/mnt/dotfiles", str(ROOT)
)


class SandboxInitTest(unittest.TestCase):
    def run_init(self, home):
        return subprocess.run(
            ["bash", "--noprofile", "--norc", "-euc", INIT + '\ncommand -v pi'],
            env={"HOME": str(home), "PATH": "/usr/bin:/bin"},
            text=True,
            capture_output=True,
            check=True,
        )

    def make_runtime(self, home):
        runtime = home / ".pi/agent/install/pi"
        runtime.parent.mkdir(parents=True)
        runtime.write_text("#!/bin/sh\nexit 0\n")
        runtime.chmod(0o755)
        launcher = home / ".local/bin/pi"
        launcher.parent.mkdir(parents=True)
        launcher.symlink_to(runtime)
        return runtime

    def test_preserves_runtime_state_and_yazi(self):
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            runtime = self.make_runtime(home)
            settings = home / ".pi/agent/settings.json"
            settings.write_text('{"local": true}\n')
            auth = home / ".pi/agent/auth.json"
            auth.write_text("secret")
            yazi = home / ".config/yazi"
            yazi.mkdir(parents=True)
            config = yazi / "yazi.toml"
            config.write_text("original")
            yazi.chmod(0o555)
            try:
                for _ in range(2):
                    result = self.run_init(home)
                    self.assertEqual(result.stderr, "")
                self.assertTrue(runtime.is_file())
                self.assertTrue(settings.is_symlink())
                self.assertEqual(
                    settings.resolve(), ROOT / "shared/.pi/agent/settings.json"
                )
                self.assertEqual(auth.read_text(), "secret")
                self.assertEqual(config.read_text(), "original")
                for name in (
                    "AGENTS.md", "APPEND_SYSTEM.md", "settings.json", "keybindings.json"
                ):
                    config_link = home / ".pi/agent" / name
                    self.assertTrue(config_link.is_symlink())
                    self.assertEqual(
                        config_link.resolve(), ROOT / "shared/.pi/agent" / name
                    )
            finally:
                yazi.chmod(0o755)

    def test_repairs_old_agent_link_without_writing_to_repo(self):
        with tempfile.TemporaryDirectory() as directory:
            home = Path(directory)
            self.make_runtime(home)
            # Simulate the old hook's link without deleting a real runtime.
            agent = home / ".pi/agent"
            agent.rename(home / ".pi/saved-agent")
            agent.symlink_to(ROOT / "shared/.pi/agent")
            launcher = home / ".local/bin/pi"
            launcher.unlink()
            launcher.symlink_to(home / ".pi/saved-agent/install/pi")
            self.run_init(home)
            self.assertFalse(agent.is_symlink())
            self.assertTrue((agent / "settings.json").is_file())
            self.assertTrue((agent / "settings.json").is_symlink())


if __name__ == "__main__":
    unittest.main()
