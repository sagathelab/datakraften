package system

import (
	"os"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
)

func DetectWSL() (bool, int) {
	r := exec.Run("uname", "-r")
	if strings.Contains(strings.ToLower(r.Stdout), "microsoft") ||
		strings.Contains(strings.ToLower(r.Stdout), "wsl") {
		r2 := exec.Run("cat", "/proc/sys/fs/binfmt_misc/WSLInterop")
		if r2.Code == 0 {
			return true, 2
		}
		return true, 1
	}
	return false, 0
}

func Distro() (string, string) {
	r := exec.Run("cat", "/etc/os-release")
	if r.Code != 0 {
		return "", ""
	}

	var name, version string
	for _, line := range strings.Split(r.Stdout, "\n") {
		if strings.HasPrefix(line, "ID=") {
			name = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	return name, version
}

func HasSystemd() bool {
	r := exec.Run("systemctl", "--version")
	return r.Code == 0
}

func CurrentShell() string {
	return os.Getenv("SHELL")
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/root"
	}
	return home
}

func UserName() string {
	return os.Getenv("USER")
}

func IsRoot() bool {
	return os.Geteuid() == 0
}
