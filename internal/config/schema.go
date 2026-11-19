package config

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"

	"go.yaml.in/yaml/v3"
)

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func Decode(data []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var doc, extra yaml.Node
	if err := decoder.Decode(&doc); err == io.EOF {
		return nil, fmt.Errorf("configuration is empty")
	} else if err != nil {
		return nil, fmt.Errorf("invalid YAML syntax")
	}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("configuration must contain exactly one valid YAML document")
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode || doc.Content[0].ShortTag() != "!!map" {
		return nil, fmt.Errorf("configuration must be a YAML mapping")
	}
	settings, err := readMapping(doc.Content[0], "")
	if err != nil {
		return nil, err
	}
	if err := Validate(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func readMapping(node *yaml.Node, prefix string) (map[string]any, error) {
	result := make(map[string]any)
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Kind != yaml.ScalarNode || key.ShortTag() != "!!str" || !keyPattern.MatchString(key.Value) {
			return nil, fmt.Errorf("invalid field name at line %d", key.Line)
		}
		path := prefix + key.Value
		if _, ok := result[key.Value]; ok {
			return nil, fmt.Errorf("duplicate field %s", path)
		}
		var item any
		var err error
		switch {
		case value.Kind == yaml.MappingNode && value.ShortTag() == "!!map":
			item, err = readMapping(value, path+".")
		case value.Kind == yaml.ScalarNode && value.ShortTag() == "!!null":
			err = fmt.Errorf("%s cannot be null; omit it or supply a value", path)
		case value.Kind == yaml.ScalarNode && slices.Contains([]string{"!!str", "!!bool", "!!int"}, value.ShortTag()):
			if value.Decode(&item) != nil {
				err = fmt.Errorf("%s has an invalid value", path)
			}
		default:
			err = fmt.Errorf("%s must be a mapping, string, boolean or integer (YAML aliases and custom tags are not supported)", path)
		}
		if err != nil {
			return nil, err
		}
		result[key.Value] = item
	}
	return result, nil
}

func Validate(data map[string]any) error {
	source, hasSource := data["source"]
	profile, hasProfile := data["profile"]
	_, hasURL := data["url"]
	_, hasTeam := data["team"]
	if source == "team" || profile == "team" || hasTeam || hasURL {
		return retiredTeamError()
	}
	if hasSource {
		s, _ := source.(string)
		if err := ValidateSource(s); err != nil {
			return err
		}
	}
	if hasProfile {
		s, _ := profile.(string)
		if ValidateSource(s) != nil {
			return fmt.Errorf("profile must be default or custom; use source instead")
		}
		if hasSource && source != s {
			return fmt.Errorf("profile conflicts with source; use source instead")
		}
		data["source"] = s
		delete(data, "profile")
	}

	for key, value := range data {
		switch key {
		case "source":
		case "version":
			if value != 1 && value != "1" {
				return fmt.Errorf("version must be 1")
			}
		case "system", "tooling":
			if err := validateFields(key, value, "package_manager"); err != nil {
				return err
			}
		case "shell":
			if err := validateFields(key, value, "default", "prompt", "history", "fuzzy_finder"); err != nil {
				return err
			}
		case "runtimes", "ai_tools", "ai_apps", "tools", "editors":
			if err := validateEntries(key, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown field %s", key)
		}
	}
	return nil
}

func validateEntries(section string, value any) error {
	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be a mapping", section)
	}
	for key, item := range m {
		path := section + "." + key
		switch section {
		case "runtimes", "ai_tools", "ai_apps":
			if section == "runtimes" {
				if !slices.Contains([]string{"node", "python", "go", "dotnet"}, key) {
					return fmt.Errorf("unknown field %s", path)
				}
			} else if enabled, ok := item.(bool); ok {
				m[key] = map[string]any{"enabled": enabled}
				continue
			}
			if err := validateFields(path, item, "enabled", "manager", "version"); err != nil {
				return err
			}
		case "tools":
			if _, ok := item.(bool); !ok {
				return fmt.Errorf("%s must be a boolean", path)
			}
		case "editors":
			if _, ok := item.(bool); ok {
				continue
			}
			if text, ok := item.(string); ok {
				if key != "zed" {
					continue
				}
				if text == "true" || text == "false" {
					m[key] = text == "true"
					continue
				}
			}
			if key == "zed" {
				return fmt.Errorf("%s must be true or false", path)
			}
			return fmt.Errorf("%s must be a boolean or string", path)
		}
	}
	return nil
}

func validateFields(path string, value any, fields ...string) error {
	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be a mapping", path)
	}
	for key, item := range m {
		if !slices.Contains(fields, key) {
			return fmt.Errorf("unknown field %s.%s", path, key)
		}
		if key == "enabled" {
			if _, ok := item.(bool); !ok {
				return fmt.Errorf("%s.enabled must be a boolean", path)
			}
		} else if _, ok := item.(string); !ok {
			return fmt.Errorf("%s.%s must be a string", path, key)
		}
	}
	return nil
}
