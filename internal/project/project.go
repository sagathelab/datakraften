package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const maxConfigSize = 1 << 20

type Project struct {
	Root   string
	Path   string
	Digest string
	Config Config
}

type Config struct {
	Version int      `yaml:"version"`
	Project Settings `yaml:"project"`
}

type Settings struct {
	Tools map[string]string `yaml:"tools"`
	Env   map[string]string `yaml:"env"`
	Tasks map[string]string `yaml:"tasks"`
}

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var itemName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

func Discover(start string) (*Project, error) {
	start, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve project directory: %w", err)
	}
	start, err = filepath.EvalSymlinks(start)
	if err != nil {
		return nil, fmt.Errorf("resolve project directory: %w", err)
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, fmt.Errorf("inspect project directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project directory is not a directory: %s", start)
	}
	gitRoot, err := findGitRoot(start)
	if err != nil {
		return nil, err
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, "datakraften.yaml")
		info, err := os.Lstat(path)
		if err == nil {
			if !info.Mode().IsRegular() {
				return nil, fmt.Errorf("project config must be a regular file: %s", path)
			}
			return load(path)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("inspect project config: %w", err)
		}
		if gitRoot == "" || dir == gitRoot {
			return nil, nil
		}
	}
}

func findGitRoot(start string) (string, error) {
	for dir := start; ; dir = filepath.Dir(dir) {
		info, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return dir, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect git boundary: %w", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
	}
}

func readConfig(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("project config is not a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(f, maxConfigSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxConfigSize {
		return nil, errors.New("project config exceeds 1 MiB")
	}
	return data, nil
}

func load(path string) (*Project, error) {
	data, err := readConfig(path)
	if err != nil {
		return nil, fmt.Errorf("read project config: %w", err)
	}
	var node yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("parse project config: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("parse project config: %w", err)
		}
		return nil, errors.New("project config must contain one YAML document")
	}
	if err := validateNode(&node); err != nil {
		return nil, err
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("project config must be a YAML mapping")
	}
	root := node.Content[0]
	hasProject := false
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "project" {
			if root.Content[i+1].Kind != yaml.MappingNode {
				return nil, errors.New("project must be a mapping")
			}
			hasProject = true
		}
	}
	if !hasProject {
		return nil, errors.New("project config requires a project mapping")
	}
	var cfg Config
	decoder = yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse project config: %w", err)
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported project config version %d; expected 1", cfg.Version)
	}
	for name, version := range cfg.Project.Tools {
		if !itemName.MatchString(name) || strings.TrimSpace(version) == "" || strings.ContainsAny(version, "\x00\r\n") {
			return nil, fmt.Errorf("invalid tool or version for %q", name)
		}
		if name != "node" {
			return nil, fmt.Errorf("unsupported project tool %q; only node is currently supported", name)
		}
	}
	for name, value := range cfg.Project.Env {
		if !envName.MatchString(name) || strings.ContainsRune(value, '\x00') {
			return nil, fmt.Errorf("invalid environment variable %q", name)
		}
	}
	for name, command := range cfg.Project.Tasks {
		if !itemName.MatchString(name) || strings.TrimSpace(command) == "" || strings.ContainsRune(command, '\x00') {
			return nil, fmt.Errorf("invalid task %q", name)
		}
	}
	sum := sha256.Sum256(data)
	return &Project{Root: filepath.Dir(path), Path: path, Digest: hex.EncodeToString(sum[:]), Config: cfg}, nil
}

func validateNode(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Tag == "!!null" {
		return errors.New("project config does not support aliases or null values")
	}
	for _, child := range node.Content {
		if err := validateNode(child); err != nil {
			return err
		}
	}
	return nil
}
