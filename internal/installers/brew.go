package installers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/system"
)

var DefaultBrewPackages = []string{
	"gh",
	"azure-cli",
	"fnm",
	"uv",
	"atuin",
	"fzf",
	"broot",
	"fd",
	"bottom",
	"starship",
	"powershell",
}

var AIToolPackages = map[string]string{
	"opencode":      "opencode",
	"github-copilot-cli": "github/copilot-cli/copilot",
}

func BrewInstalled() bool {
	return exec.CommandExists("brew")
}

func BrewInstallScript() string {
	return `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
}

func BrewPrefix() string {
	r := exec.Run("brew", "--prefix")
	if r.Code == 0 {
		return strings.TrimSpace(r.Stdout)
	}

	for _, prefix := range []string{"/home/linuxbrew/.linuxbrew", "/usr/local", "/opt/homebrew"} {
		if info, err := os.Stat(filepath.Join(prefix, "bin", "brew")); err == nil && !info.IsDir() {
			return prefix
		}
	}
	return "/home/linuxbrew/.linuxbrew"
}

func BrewEnsureInstalled() (bool, error) {
	if BrewInstalled() {
		return false, nil
	}

	fmt.Println("    Installing Homebrew...")
	home := system.HomeDir()

	r := exec.RunWithInput("\n\n",
		"bash", "-c",
		fmt.Sprintf(`echo | NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`),
	)
	if r.Code != 0 {
		return false, fmt.Errorf("homebrew install failed: %s", r.Stderr)
	}

	brewBin := filepath.Join(BrewPrefix(), "bin")
	profilePath := filepath.Join(home, ".profile")

	initLine := fmt.Sprintf(`eval "$(%s/brew shellenv)"`, brewBin)

	content, _ := os.ReadFile(profilePath)
	if !strings.Contains(string(content), initLine) {
		f, _ := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		fmt.Fprintf(f, "\n# >>> datakraften >>>\n%s\n# <<< datakraften <<<\n", initLine)
	}

	os.Setenv("PATH", brewBin+":"+os.Getenv("PATH"))

	return true, nil
}

func BrewPackageInstalled(pkg string) bool {
	if pkg == "" {
		return false
	}
	r := exec.Run("brew", "list", "--formula", pkg)
	if r.Code == 0 {
		return true
	}
	r2 := exec.Run("brew", "list", "--cask", pkg)
	return r2.Code == 0
}

func BrewInstall(pkg string) error {
	if BrewPackageInstalled(pkg) {
		return nil
	}
	fmt.Printf("    Installing %s via Homebrew...\n", pkg)
	r := exec.Run("brew", "install", pkg)
	if r.Code != 0 {
		return fmt.Errorf("brew install %s failed: %s", pkg, r.Stderr)
	}
	return nil
}

func BrewEnsurePackages(pkgs []string) (int, error) {
	var toInstall []string
	for _, pkg := range pkgs {
		if !BrewPackageInstalled(pkg) {
			toInstall = append(toInstall, pkg)
		}
	}
	if len(toInstall) == 0 {
		return 0, nil
	}

	for _, pkg := range toInstall {
		if err := BrewInstall(pkg); err != nil {
			return 0, err
		}
	}
	return len(toInstall), nil
}
