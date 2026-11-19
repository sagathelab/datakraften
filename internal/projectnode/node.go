package projectnode

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrBackendNotFound = errors.New("fnm is not available in PATH")
	ErrNotInstalled    = errors.New("requested Node version is not installed")
	ErrInvalidVersion  = errors.New("invalid Node version")
)

type Installation struct {
	Requested string
	Version   string
	NodePath  string
	BinDir    string
	Backend   string
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
	if err != nil || home == "" {
		return "", fmt.Errorf("cannot determine home directory")
	}
	return filepath.Join(home, ".local", "share"), nil
}

func stores() ([]string, error) {
	data, err := dataHome()
	if err != nil {
		return nil, err
	}
	managed := filepath.Join(data, "datakraften", "fnm")
	if dir := os.Getenv("FNM_DIR"); dir != "" {
		if !filepath.IsAbs(dir) {
			return nil, fmt.Errorf("FNM_DIR must be an absolute path")
		}
		return []string{dir, managed}, nil
	}
	modern := filepath.Join(data, "fnm")
	if _, err := os.Stat(modern); err == nil || !os.IsNotExist(err) {
		if err != nil {
			return nil, err
		}
		return []string{modern, managed}, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	legacy := filepath.Join(home, ".fnm")
	if _, err := os.Stat(legacy); err == nil || !os.IsNotExist(err) {
		if err != nil {
			return nil, err
		}
		return []string{legacy, managed}, nil
	}
	if runtime.GOOS == "darwin" {
		legacy = filepath.Join(home, "Library", "Application Support", "fnm")
		if _, err := os.Stat(legacy); err == nil || !os.IsNotExist(err) {
			if err != nil {
				return nil, err
			}
			return []string{legacy, managed}, nil
		}
	}
	return []string{modern, managed}, nil
}

func Resolve(request string) (Installation, error) {
	wanted, err := parseVersion(request, false)
	if err != nil {
		return Installation{}, fmt.Errorf("%w: %q; use a major, major.minor or major.minor.patch version", ErrInvalidVersion, request)
	}
	if _, err := exec.LookPath("fnm"); err != nil {
		return Installation{}, ErrBackendNotFound
	}
	dirs, err := stores()
	if err != nil {
		return Installation{}, err
	}
	type candidate struct {
		version version
		info    Installation
	}
	var candidates []candidate
	for _, root := range dirs {
		base := filepath.Join(root, "node-versions")
		entries, err := os.ReadDir(base)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Installation{}, fmt.Errorf("read fnm installations: %w", err)
		}
		for _, entry := range entries {
			v, err := parseVersion(entry.Name(), true)
			if err != nil {
				continue
			}
			matches := true
			for i := 0; i < wanted.count; i++ {
				if v.parts[i] != wanted.parts[i] {
					matches = false
				}
			}
			if !matches {
				continue
			}
			bin := filepath.Join(base, entry.Name(), "installation", "bin")
			node := filepath.Join(bin, "node")
			if info, err := os.Stat(node); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
				continue
			}
			candidates = append(candidates, candidate{v, Installation{Requested: request, Version: strings.TrimPrefix(entry.Name(), "v"), NodePath: node, BinDir: bin, Backend: "fnm"}})
		}
	}
	if len(candidates) == 0 {
		return Installation{}, fmt.Errorf("%w: %s; run dk project setup to install it", ErrNotInstalled, request)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		for n := 0; n < 3; n++ {
			if candidates[i].version.parts[n] != candidates[j].version.parts[n] {
				return candidates[i].version.parts[n] > candidates[j].version.parts[n]
			}
		}
		return false
	})
	return candidates[0].info, nil
}

func Install(request string, stdout, stderr io.Writer) error {
	if _, err := parseVersion(request, false); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidVersion, request)
	}
	fnm, err := exec.LookPath("fnm")
	if err != nil {
		return ErrBackendNotFound
	}
	if _, err := Resolve(request); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotInstalled) {
		return err
	}
	data, err := dataHome()
	if err != nil {
		return err
	}
	cmd := exec.Command(fnm, "install", request)
	cmd.Env = replaceEnv(os.Environ(), "FNM_DIR", filepath.Join(data, "datakraften", "fnm"))
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fnm install %s failed: %w", request, err)
	}
	if _, err := Resolve(request); err != nil {
		return fmt.Errorf("fnm did not install requested Node version: %w", err)
	}
	return nil
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
	return replaceEnv(base, "PATH", path)
}

func replaceEnv(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, key+"=") {
			out = append(out, entry)
		}
	}
	return append(out, key+"="+value)
}
