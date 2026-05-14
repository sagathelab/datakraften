package profiles

type Profile struct {
	Name        string
	Description string
}

var allProfiles = []Profile{
	{Name: "minimal", Description: "Core system tools only"},
	{Name: "default", Description: "General developer setup"},
	{Name: "ai", Description: "AI-native development environment"},
	{Name: "dotnet", Description: ".NET developer setup"},
	{Name: "frontend", Description: "Frontend developer setup"},
	{Name: "platform", Description: "Platform engineer setup"},
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
