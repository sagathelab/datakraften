package app

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sagathelab/datakraften/internal/config"
	"go.yaml.in/yaml/v3"
)

const maxImportSize = 1 << 20

func fetchInitTemplate(address string) ([]byte, string, error) {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return nil, "", fmt.Errorf("template URL must be an HTTPS URL without embedded credentials")
	}
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != "https" || req.URL.Hostname() == "" || req.URL.User != nil {
			return fmt.Errorf("invalid template redirect")
		}
		req.Header.Del("Referer")
		return nil
	}}
	resp, err := client.Get(address)
	if err != nil {
		return nil, "", fmt.Errorf("cannot download template; check the HTTPS raw YAML address, network and access")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("template server returned HTTP %d; check the raw YAML address and access", resp.StatusCode)
	}
	if contentType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); contentType == "text/html" || contentType == "application/xhtml+xml" {
		return nil, "", fmt.Errorf("template server returned HTML; use an HTTPS raw YAML address")
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
	base, err := config.Decode(defaults)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid built-in defaults: %w", err)
	}
	partial, err := config.Decode(template)
	if err != nil {
		return nil, nil, err
	}
	return mergeConfig(base, partial)
}

func mergeConfig(base, partial map[string]any) ([]byte, []string, error) {
	delete(partial, "source")
	var paths []string
	mergeImport(base, partial, "", &paths)
	base["source"] = "custom"
	if _, ok := base["version"]; !ok {
		base["version"] = 1
	}
	if err := config.Validate(base); err != nil {
		return nil, nil, fmt.Errorf("invalid merged configuration: %w", err)
	}
	out, err := yaml.Marshal(base)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot encode configuration: %w", err)
	}
	sort.Strings(paths)
	return out, paths, nil
}

func mergeImport(dst, src map[string]any, prefix string, paths *[]string) {
	for key, value := range src {
		path := prefix + key
		if child, ok := value.(map[string]any); ok {
			if existing, ok := dst[key].(map[string]any); ok {
				mergeImport(existing, child, path+".", paths)
				continue
			}
			if len(child) > 0 {
				nested := make(map[string]any)
				mergeImport(nested, child, path+".", paths)
				dst[key] = nested
				continue
			}
		}
		dst[key] = value
		*paths = append(*paths, path)
	}
}

func initFromURL(configPath, address string, existing, yes bool) error {
	data, host, err := fetchInitTemplate(address)
	if err != nil {
		return err
	}
	defaults := []byte(builtInConfig(defaultConfigYAML))
	result, paths, err := importConfig(defaults, data)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	fmt.Printf("  Template host: %s\n", host)
	if len(paths) == 0 {
		fmt.Println("  Template uses all defaults.")
	} else {
		if err := previewImport(defaults, result, paths); err != nil {
			return err
		}
	}
	fmt.Println("  The complete result will be stored locally for review and editing.")
	saved, err := saveInitConfig(configPath, result, existing, yes)
	if err != nil || !saved {
		return err
	}
	fmt.Printf("  Local configuration saved: %s\n", configPath)
	fmt.Println("  Edit the local file as needed, then run 'dk apply'. The template is not fetched again.")
	return nil
}

func previewImport(defaults, result []byte, paths []string) error {
	before, err := config.Decode(defaults)
	if err != nil {
		return fmt.Errorf("cannot preview defaults: %w", err)
	}
	after, err := config.Decode(result)
	if err != nil {
		return fmt.Errorf("cannot preview configuration: %w", err)
	}
	fmt.Println("  Template overrides (text values hidden):")
	for _, path := range paths {
		fmt.Printf("    %s: %s -> %s\n", path, previewValue(before, path), previewValue(after, path))
	}
	warnUnusedConfig(paths)
	return nil
}

func warnUnusedConfig(paths []string) {
	var ignored []string
	for _, path := range paths {
		if path != "version" && !importSettingApplied(path) {
			ignored = append(ignored, path)
		}
	}
	if len(ignored) > 0 {
		fmt.Printf("  [WARN] These settings will be saved, but current dk apply does not use them: %s\n", strings.Join(ignored, ", "))
	}
}

func importSettingApplied(path string) bool {
	switch path {
	case "runtimes.node.enabled", "runtimes.python.enabled", "runtimes.go.enabled", "runtimes.dotnet.enabled", "editors.zed",
		"ai_tools.codex.enabled", "ai_tools.claude.enabled", "ai_tools.gemini.enabled", "ai_tools.opencode.enabled", "ai_tools.copilot.enabled",
		"ai_apps.codex.enabled", "ai_apps.claude.enabled", "ai_apps.copilot.enabled":
		return true
	default:
		return false
	}
}

func previewValue(root map[string]any, path string) string {
	var value any = root
	for _, key := range strings.Split(path, ".") {
		m, ok := value.(map[string]any)
		if !ok {
			return "(not set)"
		}
		value, ok = m[key]
		if !ok {
			return "(not set)"
		}
	}
	switch v := value.(type) {
	case bool, int:
		return fmt.Sprint(v)
	case string:
		if v == "" {
			return "(empty text)"
		}
		return "(text hidden)"
	case map[string]any:
		return fmt.Sprintf("(%d fields)", len(v))
	default:
		return "(unknown)"
	}
}

func writeInitAtomic(path string, data []byte, replace bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
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
	if replace {
		err = os.Rename(temp, path)
	} else {
		err = os.Link(temp, path)
	}
	if err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	return nil
}
