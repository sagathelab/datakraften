package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/installers"
	"go.yaml.in/yaml/v3"
)

const maxImportSize = 1 << 20

var importKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func fetchInitTemplate(address string) ([]byte, string, error) {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return nil, "", fmt.Errorf("template URL must be an HTTPS URL without embedded credentials")
	}
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != "https" || req.URL.User != nil {
			return fmt.Errorf("invalid template redirect")
		}
		return nil
	}}
	resp, err := client.Get(address)
	if err != nil {
		return nil, "", fmt.Errorf("cannot download template; check the address, network and access")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("template server returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImportSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("cannot read template response")
	}
	if len(data) > maxImportSize {
		return nil, "", fmt.Errorf("template exceeds the 1 MiB limit")
	}
	return data, u.Hostname(), nil
}

func importConfig(defaults, template []byte) ([]byte, []string, error) {
	base := map[string]interface{}{}
	partial := map[string]interface{}{}
	if err := yaml.Unmarshal(defaults, &base); err != nil {
		return nil, nil, fmt.Errorf("invalid built-in defaults: %w", err)
	}
	if len(strings.TrimSpace(string(template))) == 0 {
		return nil, nil, fmt.Errorf("template is empty")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(template))
	if err := decoder.Decode(&partial); err != nil || partial == nil {
		return nil, nil, fmt.Errorf("template is not a valid YAML mapping")
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, nil, fmt.Errorf("template must contain exactly one YAML document")
	}
	if err := validateImport(partial); err != nil {
		return nil, nil, err
	}
	delete(partial, "source")
	paths := []string{}
	mergeImport(base, partial, "", &paths)
	base["source"] = "custom"
	if err := validateImport(base); err != nil {
		return nil, nil, fmt.Errorf("invalid merged configuration: %w", err)
	}
	out, err := yaml.Marshal(base)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot encode configuration: %w", err)
	}
	sort.Strings(paths)
	return out, paths, nil
}

func mergeImport(dst, src map[string]interface{}, prefix string, paths *[]string) {
	for k, v := range src {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		if child, ok := v.(map[string]interface{}); ok {
			if existing, ok := dst[k].(map[string]interface{}); ok {
				mergeImport(existing, child, path, paths)
				continue
			}
		}
		dst[k] = v
		*paths = append(*paths, path)
	}
}

func validateImport(data map[string]interface{}) error {
	fields := map[string]map[string]string{
		"system": {"package_manager": "string"}, "tooling": {"package_manager": "string"},
		"shell": {"default": "string", "prompt": "string", "history": "string", "fuzzy_finder": "string"},
	}
	runtimeFields := map[string]string{"enabled": "bool", "manager": "string", "version": "string"}
	for name, value := range data {
		if !importKeyPattern.MatchString(name) {
			return fmt.Errorf("template contains an invalid field name")
		}
		switch name {
		case "version":
			version, integer := value.(int)
			textVersion, text := value.(string)
			if !(integer && version == 1) && !(text && textVersion == "1") {
				return fmt.Errorf("version must be 1")
			}
		case "source":
			if s, ok := value.(string); !ok || (s != "default" && s != "custom") {
				return fmt.Errorf("source must be default or custom")
			}
		case "system", "tooling", "shell":
			if err := validateImportFields(name, value, fields[name]); err != nil {
				return err
			}
		case "runtimes":
			m, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("runtimes must be a mapping")
			}
			for key, item := range m {
				if key != "node" && key != "python" && key != "go" && key != "dotnet" {
					return fmt.Errorf("unknown field runtimes.%s", key)
				}
				if err := validateImportFields("runtimes."+key, item, runtimeFields); err != nil {
					return err
				}
			}
		case "ai_tools", "ai_apps":
			m, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%s must be a mapping", name)
			}
			for key, item := range m {
				if !importKeyPattern.MatchString(key) {
					return fmt.Errorf("%s contains an invalid field name", name)
				}
				if enabled, ok := item.(bool); ok {
					m[key] = map[string]interface{}{"enabled": enabled}
					continue
				}
				if err := validateImportFields(name+"."+key, item, runtimeFields); err != nil {
					return err
				}
			}
		case "tools", "editors":
			m, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%s must be a mapping", name)
			}
			for key, item := range m {
				if !importKeyPattern.MatchString(key) {
					return fmt.Errorf("%s contains an invalid field name", name)
				}
				_, b := item.(bool)
				_, s := item.(string)
				if !b && !(name == "editors" && s) {
					return fmt.Errorf("%s.%s has an invalid type", name, key)
				}
			}
		default:
			return fmt.Errorf("unknown field %s", name)
		}
	}
	return nil
}

func validateImportFields(path string, value interface{}, fields map[string]string) error {
	m, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s must be a mapping", path)
	}
	for key, item := range m {
		if !importKeyPattern.MatchString(key) {
			return fmt.Errorf("%s contains an invalid field name", path)
		}
		kind, ok := fields[key]
		if !ok {
			return fmt.Errorf("unknown field %s.%s", path, key)
		}
		switch kind {
		case "string":
			if _, ok := item.(string); !ok {
				return fmt.Errorf("%s.%s must be a string", path, key)
			}
		case "bool":
			if _, ok := item.(bool); !ok {
				return fmt.Errorf("%s.%s must be a boolean", path, key)
			}
		}
	}
	return nil
}

func initFromURL(configPath, configDir, address string, existing, migrate, yes bool) error {
	if migrate {
		cfg, err := config.Load()
		if err != nil || cfg.Source != "team" || cfg.URL == "" {
			return fmt.Errorf("existing team configuration with a URL is required for migration")
		}
		address = cfg.URL
	}
	data, host, err := fetchInitTemplate(address)
	if err != nil {
		return err
	}
	defaults := []byte(fmt.Sprintf(defaultConfigYAML, string(installers.DetectPackageManager())))
	result, paths, err := importConfig(defaults, data)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	fmt.Printf("  Template host: %s\n", host)
	if len(paths) == 0 {
		fmt.Println("  Template uses all defaults.")
	} else {
		var before, after map[string]interface{}
		if err := yaml.Unmarshal(defaults, &before); err != nil {
			return fmt.Errorf("cannot preview defaults")
		}
		if err := yaml.Unmarshal(result, &after); err != nil {
			return fmt.Errorf("cannot preview configuration")
		}
		fmt.Println("  Template overrides:")
		for _, path := range paths {
			fmt.Printf("    %s: %s -> %s\n", path, previewValue(before, path), previewValue(after, path))
		}
		var ignored []string
		for _, path := range paths {
			if !importSettingApplied(path) {
				ignored = append(ignored, path)
			}
		}
		if len(ignored) > 0 {
			fmt.Printf("  [WARN] These imported settings are saved, but current dk apply does not use them: %s\n", strings.Join(ignored, ", "))
		}
	}
	fmt.Println("  Text values are hidden. Other settings use Datakraften defaults.")
	fmt.Println("  [WARN] dk apply currently honors runtime and supported AI enabled flags and editors.zed; other settings may not affect apply.")
	fmt.Println("  The complete result will be stored locally for review and editing.")
	if existing && !yes {
		var confirmed bool
		prompt := &survey.Confirm{Message: "Replace the existing local configuration?", Default: false}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return fmt.Errorf("confirmation required: %w", err)
		}
		if !confirmed {
			fmt.Println("  Cancelled; existing configuration preserved.")
			return nil
		}
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	var backup string
	if migrate {
		original, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("read team config: %w", err)
		}
		backup = configPath + ".team.bak"
		f, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return fmt.Errorf("create team backup: %w", err)
		}
		_, writeErr := f.Write(original)
		closeErr := f.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(backup)
			return fmt.Errorf("write team backup failed")
		}
	}
	if err := writeInitAtomic(configPath, result); err != nil {
		if backup != "" {
			_ = os.Remove(backup)
		}
		return err
	}
	fmt.Printf("  Local configuration saved: %s\n", configPath)
	if backup != "" {
		fmt.Printf("  Previous team configuration saved: %s\n", backup)
	}
	fmt.Println("  Edit the local file as needed, then run 'dk apply'. The template is not fetched again.")
	return nil
}

func importSettingApplied(path string) bool {
	switch path {
	case "version", "runtimes.node.enabled", "runtimes.python.enabled", "runtimes.go.enabled", "runtimes.dotnet.enabled", "editors.zed",
		"ai_tools.codex.enabled", "ai_tools.claude.enabled", "ai_tools.gemini.enabled", "ai_tools.opencode.enabled", "ai_tools.copilot.enabled",
		"ai_apps.codex.enabled", "ai_apps.claude.enabled", "ai_apps.copilot.enabled":
		return true
	default:
		return false
	}
}

func previewValue(root map[string]interface{}, path string) string {
	var value interface{} = root
	for _, key := range strings.Split(path, ".") {
		m, ok := value.(map[string]interface{})
		if !ok {
			return "(not set)"
		}
		value, ok = m[key]
		if !ok {
			return "(not set)"
		}
	}
	switch v := value.(type) {
	case bool, int, float64:
		return fmt.Sprint(v)
	case string:
		if v == "" {
			return "(empty text)"
		}
		return "(text hidden)"
	case []interface{}:
		return fmt.Sprintf("(%d entries)", len(v))
	case map[string]interface{}:
		return fmt.Sprintf("(%d fields)", len(v))
	default:
		return "(unknown)"
	}
}

func writeInitAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write configuration: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync configuration: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close configuration: %w", err)
	}
	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	return nil
}
