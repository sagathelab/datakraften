package projectnode

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrBackendNotFound = errors.New("Node version manager is not available")
	ErrNotInstalled    = errors.New("requested Node version is not installed")
	ErrInvalidVersion  = errors.New("invalid Node version")
)

type Installation struct {
	Requested      string
	Version        string
	NodePath       string
	BinDir         string
	Backend        string
	ManagerPath    string
	StorePath      string
	ManagedStore   string
	DownloadSource string
}

type backend struct {
	name         string
	managerPath  string
	userStores   []string
	managedStore string
	mirrorEnv    string
}

type version struct {
	parts [3]uint64
	count int
}

func parseVersion(s string, exact bool) (version, error) {
	var v version
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) > 3 || (exact && len(parts) != 3) {
		return v, ErrInvalidVersion
	}
	for i, part := range parts {
		if part == "" || len(part) > 1 && part[0] == '0' {
			return v, ErrInvalidVersion
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return v, ErrInvalidVersion
			}
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return v, ErrInvalidVersion
		}
		v.parts[i] = n
	}
	v.count = len(parts)
	return v, nil
}

func dataHome() (string, error) {
	if p := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(p) {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil || !filepath.IsAbs(home) {
		return "", fmt.Errorf("cannot determine home directory")
	}
	return filepath.Join(home, ".local", "share"), nil
}

func fnmStore(data string) (string, error) {
	if dir := os.Getenv("FNM_DIR"); dir != "" {
		if !filepath.IsAbs(dir) {
			return "", fmt.Errorf("FNM_DIR must be an absolute path")
		}
		return dir, nil
	}
	modern := filepath.Join(data, "fnm")
	dirs := []string{modern}
	if home, err := os.UserHomeDir(); err == nil && filepath.IsAbs(home) {
		dirs = append(dirs, filepath.Join(home, ".fnm"))
		if runtime.GOOS == "darwin" {
			dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "fnm"))
		}
	}
	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("inspect fnm store: %w", err)
		}
		return dir, nil
	}
	return modern, nil
}

func nvmStores() ([]string, error) {
	if dir := os.Getenv("NVM_DIR"); dir != "" {
		if !filepath.IsAbs(dir) {
			return nil, fmt.Errorf("NVM_DIR must be an absolute path")
		}
		return []string{dir}, nil
	}
	var dirs []string
	if config := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(config) {
		dirs = append(dirs, filepath.Join(config, "nvm"))
	}
	if home, err := os.UserHomeDir(); err == nil && filepath.IsAbs(home) {
		dirs = append(dirs, filepath.Join(home, ".nvm"))
	}
	return dirs, nil
}

func nvmScript(stores []string) (string, error) {
	var paths []string
	for _, dir := range stores {
		paths = append(paths, filepath.Join(dir, "nvm.sh"))
	}
	switch runtime.GOOS {
	case "linux":
		paths = append(paths, "/usr/share/nvm/nvm.sh", "/home/linuxbrew/.linuxbrew/opt/nvm/nvm.sh")
	case "darwin":
		paths = append(paths, "/opt/homebrew/opt/nvm/nvm.sh", "/usr/local/opt/nvm/nvm.sh")
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect nvm script: %w", err)
		}
		if err == nil && info.Mode().IsRegular() {
			return path, nil
		}
	}
	return "", fmt.Errorf("%w: nvm.sh not found; install nvm or set NVM_DIR to its directory", ErrBackendNotFound)
}

func findBackend(manager string) (backend, error) {
	data, err := dataHome()
	if err != nil {
		return backend{}, err
	}
	b := backend{name: manager, managedStore: filepath.Join(data, "datakraften", manager)}
	switch manager {
	case "fnm":
		b.managerPath, err = exec.LookPath("fnm")
		if err != nil || !filepath.IsAbs(b.managerPath) {
			return backend{}, fmt.Errorf("%w: install fnm and make it available in an absolute directory in PATH", ErrBackendNotFound)
		}
		store, err := fnmStore(data)
		if err != nil {
			return backend{}, err
		}
		b.userStores, b.mirrorEnv = []string{store}, "FNM_NODE_DIST_MIRROR"
	case "nvm":
		b.userStores, err = nvmStores()
		if err != nil {
			return backend{}, err
		}
		b.managerPath, err = nvmScript(b.userStores)
		if err != nil {
			return backend{}, err
		}
		b.mirrorEnv = "NVM_NODEJS_ORG_MIRROR"
	default:
		return backend{}, fmt.Errorf("unsupported Node version manager %q; choose fnm or nvm", manager)
	}
	return b, nil
}

func Resolve(request string) (Installation, error) {
	return ResolveFor("fnm", request)
}

func ResolveFor(manager, request string) (Installation, error) {
	wanted, err := parseVersion(request, false)
	if err != nil {
		return Installation{}, fmt.Errorf("%w: %q; use a major, major.minor or major.minor.patch version", ErrInvalidVersion, request)
	}
	b, err := findBackend(manager)
	if err != nil {
		return Installation{}, err
	}
	return b.resolve(request, wanted, b.stores())
}

func (b backend) stores() []string {
	return append(append([]string(nil), b.userStores...), b.managedStore)
}

func (b backend) resolve(request string, wanted version, dirs []string) (Installation, error) {
	best := Installation{
		Requested: request, Backend: b.name, ManagerPath: b.managerPath, ManagedStore: b.managedStore,
		DownloadSource: "https://nodejs.org/dist",
	}
	if os.Getenv(b.mirrorEnv) != "" {
		best.DownloadSource = "mirror configured via " + b.mirrorEnv
	}
	var bestVersion version
	for _, root := range dirs {
		base := filepath.Join(root, "node-versions")
		if b.name == "nvm" {
			base = filepath.Join(root, "versions", "node")
		}
		entries, err := os.ReadDir(base)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Installation{}, fmt.Errorf("read %s installations: %w", b.name, err)
		}
		for _, entry := range entries {
			v, err := parseVersion(entry.Name(), true)
			if err != nil {
				continue
			}
			if compareParts(v, wanted, wanted.count) != 0 || (best.Version != "" && compareParts(v, bestVersion, 3) <= 0) {
				continue
			}
			bin := filepath.Join(base, entry.Name(), "installation", "bin")
			if b.name == "nvm" {
				bin = filepath.Join(base, entry.Name(), "bin")
			}
			node := filepath.Join(bin, "node")
			if info, err := os.Stat(node); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
				continue
			}
			bestVersion = v
			best.Version = strings.TrimPrefix(entry.Name(), "v")
			best.NodePath, best.BinDir, best.StorePath = node, bin, root
		}
	}
	if best.Version == "" {
		return best, fmt.Errorf("%w: %s in %s; run 'dk project setup' to install it", ErrNotInstalled, request, b.name)
	}
	return best, nil
}

func compareParts(a, b version, count int) int {
	for i := 0; i < count; i++ {
		switch {
		case a.parts[i] < b.parts[i]:
			return -1
		case a.parts[i] > b.parts[i]:
			return 1
		}
	}
	return 0
}

func Install(request string, stdout, stderr io.Writer) error {
	return InstallFor("fnm", request, stdout, stderr)
}

func InstallFor(manager, request string, stdout, stderr io.Writer) error {
	wanted, err := parseVersion(request, false)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidVersion, request)
	}
	b, err := findBackend(manager)
	if err != nil {
		return err
	}
	if _, err := b.resolve(request, wanted, b.stores()); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotInstalled) {
		return err
	}
	if err := b.prepareStore(); err != nil {
		return fmt.Errorf("prepare project %s store: %w", manager, err)
	}
	var cmd *exec.Cmd
	var dirEnv string
	var omit []string
	switch manager {
	case "fnm":
		cmd = exec.Command(b.managerPath, "install", request)
		dirEnv, omit = "FNM_DIR", []string{"FNM_MULTISHELL_PATH", "FNM_COREPACK_ENABLED"}
	case "nvm":
		// Bash -p also skips BASH_ENV and imported functions in this non-interactive shell.
		cmd = exec.Command("/bin/bash", "--noprofile", "--norc", "-p", "-c", `. "$1" --no-use && nvm install --no-progress "$2"`, "dk-nvm", b.managerPath, request)
		dirEnv, omit = "NVM_DIR", []string{"BASH_ENV", "ENV", "PREFIX", "NPM_CONFIG_PREFIX", "NVM_INSTALL_THIRD_PARTY_HOOK"}
	default:
		return fmt.Errorf("unsupported Node version manager %q", manager)
	}
	cmd.Dir = b.managedStore
	cmd.Env = append(withoutEnv(cmd.Environ(), append(omit, dirEnv)...), dirEnv+"="+b.managedStore)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s install %s failed: %w", manager, request, err)
	}
	if _, err := b.resolve(request, wanted, []string{b.managedStore}); err != nil {
		return fmt.Errorf("%s install did not produce an executable Node.js %s in %s: %w", manager, request, b.managedStore, err)
	}
	return nil
}

func (b backend) prepareStore() error {
	managed, err := os.Stat(b.managedStore)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, dir := range b.userStores {
		user, err := os.Stat(dir)
		if filepath.Clean(dir) == b.managedStore || managed != nil && err == nil && os.SameFile(user, managed) {
			return fmt.Errorf("managed store %q is also a user %s store; configure separate directories", b.managedStore, b.name)
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.MkdirAll(b.managedStore, 0700)
}

func Environment(inst Installation, base []string) []string {
	path := ""
	for _, entry := range base {
		if strings.HasPrefix(entry, "PATH=") {
			path = strings.TrimPrefix(entry, "PATH=")
		}
	}
	if path != "" {
		path = inst.BinDir + string(os.PathListSeparator) + path
	} else {
		path = inst.BinDir
	}
	return append(withoutEnv(base, "PATH"), "PATH="+path)
}

func withoutEnv(env []string, keys ...string) []string {
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if !slices.ContainsFunc(keys, func(k string) bool {
			return k == key || k == "NPM_CONFIG_PREFIX" && strings.EqualFold(k, key)
		}) {
			out = append(out, entry)
		}
	}
	return out
}
