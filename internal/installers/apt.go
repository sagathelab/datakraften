package installers

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
)

var APTBasePackages = []string{
	"build-essential",
	"curl",
	"wget",
	"git",
	"ca-certificates",
	"gnupg",
	"lsb-release",
	"unzip",
	"tar",
	"jq",
	"procps",
	"locales",
}

func APTPackageInstalled(pkg string) bool {
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
}

func APTUpdate() error {
	fmt.Println("    Updating APT cache...")
	r := exec.Run("sudo", "apt-get", "update", "-qq")
	if r.Code != 0 {
		return fmt.Errorf("apt update failed: %s", r.Stderr)
	}
	return nil
}

func APTInstall(pkg string) error {
	if APTPackageInstalled(pkg) {
		return nil
	}
	r := exec.Run("sudo", "apt-get", "install", "-y", "-qq", pkg)
	if r.Code != 0 {
		return fmt.Errorf("apt install %s failed: %s", pkg, r.Stderr)
	}
	return nil
}

func APTEnsurePackages(pkgs []string) (int, error) {
	var missing []string
	for _, pkg := range pkgs {
		if !APTPackageInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}
	if len(missing) == 0 {
		return 0, nil
	}

	if err := APTUpdate(); err != nil {
		return 0, err
	}
	for _, pkg := range missing {
		fmt.Printf("    Installing %s...\n", pkg)
		if err := APTInstall(pkg); err != nil {
			return 0, err
		}
	}
	return len(missing), nil
}
