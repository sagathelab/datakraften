package profiles

type Profile struct {
	Name        string
	Description string
}

var allProfiles = []Profile{
	{Name: "minimal", Description: "Core system packages only"},
	{Name: "default", Description: "Full developer workstation"},
	{Name: "ai", Description: "AI-native development — Codex, Claude Code, Gemini, OpenCode, Copilot, runtimes, and shell"},
	{Name: "custom", Description: "Custom configuration — edit everything locally"},
	{Name: "team", Description: "Team config — shared setup from a remote YAML URL"},
}

func Available() []string {
	names := make([]string, len(allProfiles))
	for i, p := range allProfiles {
		names[i] = p.Name
	}
	return names
}

func Describe(name string) string {
	for _, p := range allProfiles {
		if p.Name == name {
			return p.Description
		}
	}
	return ""
}

func All() []Profile {
	return allProfiles
}
