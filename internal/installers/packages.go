package installers

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
)

type PackageManager string

const (
	APT   PackageManager = "apt"
	DNF   PackageManager = "dnf"
	YUM   PackageManager = "yum"
	PACMAN PackageManager = "pacman"
	BREW  PackageManager = "brew"
)

func DetectPackageManager() PackageManager {
	if exec.CommandExists("apt-get") {
		return APT
	}
	if exec.CommandExists("dnf") {
		return DNF
	}
	if exec.CommandExists("yum") {
		return YUM
	}
	if exec.CommandExists("pacman") {
		return PACMAN
	}
	if exec.CommandExists("brew") {
		return BREW
	}
	return APT
}

func (pm PackageManager) InstallCommand(pkgs []string) (string, []string) {
	switch pm {
	case APT:
		return "sudo", append([]string{"apt-get", "install", "-y", "-qq"}, pkgs...)
	case DNF:
		return "sudo", append([]string{"dnf", "install", "-y"}, pkgs...)
	case YUM:
		return "sudo", append([]string{"yum", "install", "-y"}, pkgs...)
	case PACMAN:
		return "sudo", append([]string{"pacman", "-S", "--noconfirm"}, pkgs...)
	case BREW:
		return "brew", append([]string{"install"}, pkgs...)
	default:
		return "sudo", append([]string{"apt-get", "install", "-y", "-qq"}, pkgs...)
	}
}

func (pm PackageManager) UpdateCommand() (string, []string) {
	switch pm {
	case APT:
		return "sudo", []string{"apt-get", "update", "-qq"}
	case DNF:
		return "sudo", []string{"dnf", "makecache"}
	case YUM:
		return "sudo", []string{"yum", "makecache"}
	case PACMAN:
		return "sudo", []string{"pacman", "-Sy"}
	case BREW:
		return "brew", []string{"update"}
	default:
		return "sudo", []string{"apt-get", "update", "-qq"}
	}
}

func (pm PackageManager) CheckInstalled(pkg string) bool {
	switch pm {
	case APT:
		r := exec.Run("dpkg", "-l", pkg)
		if r.Code != 0 {
			return false
		}
		for _, line := range strings.Split(r.Stdout, "\n") {
			if strings.HasPrefix(line, "ii") && strings.Contains(line, " "+pkg+" ") {
				return true
			}
		}
		return false
	case DNF, YUM:
		r := exec.Run("rpm", "-q", pkg)
		return r.Code == 0
	case PACMAN:
		r := exec.Run("pacman", "-Q", pkg)
		return r.Code == 0
	case BREW:
		return BrewPackageInstalled(pkg)
	default:
		return exec.CommandExists(pkg)
	}
}

func NativeBasePackages() []string {
	switch DetectPackageManager() {
	case APT:
		return []string{
			"build-essential", "curl", "wget", "git",
			"ca-certificates", "gnupg", "lsb-release",
			"unzip", "tar", "jq", "procps", "locales",
		}
	case DNF, YUM:
		return []string{
			"@development-tools", "curl", "wget", "git",
			"ca-certificates", "gnupg", "unzip", "tar",
			"jq", "procps-ng", "glibc-langpack-en",
		}
	case PACMAN:
		return []string{
			"base-devel", "curl", "wget", "git",
			"ca-certificates", "gnupg", "unzip", "tar",
			"jq", "procps-ng",
		}
	case BREW:
		return []string{
			"curl", "wget", "git", "gnupg",
			"unzip", "jq",
		}
	default:
		return []string{"curl", "wget", "git", "unzip", "jq"}
	}
}

func NativeUpdate() error {
	pm := DetectPackageManager()
	exe, args := pm.UpdateCommand()
	fmt.Printf("    Updating %s cache...\n", pm)
	r := exec.Run(exe, args...)
	if r.Code != 0 {
		return fmt.Errorf("%s update failed: %s", pm, r.Stderr)
	}
	return nil
}

func NativeInstall(pkg string) error {
	pm := DetectPackageManager()
	if pm.CheckInstalled(pkg) {
		return nil
	}
	fmt.Printf("    Installing %s via %s...\n", pkg, pm)
	exe, args := pm.InstallCommand([]string{pkg})
	r := exec.Run(exe, args...)
	if r.Code != 0 {
		return fmt.Errorf("%s install %s failed: %s", pm, pkg, r.Stderr)
	}
	return nil
}

func NativeEnsurePackages(pkgs []string) ([]string, error) {
	pm := DetectPackageManager()
	var missing []string
	for _, pkg := range pkgs {
		if !pm.CheckInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	if err := NativeUpdate(); err != nil {
		return nil, err
	}
	for _, pkg := range missing {
		if err := NativeInstall(pkg); err != nil {
			return nil, err
		}
	}
	return missing, nil
}

func FishPackageName() string {
	pm := DetectPackageManager()
	switch pm {
	case DNF, YUM:
		return "fish"
	case PACMAN:
		return "fish"
	case APT:
		return "fish"
	default:
		return "fish"
	}
}
