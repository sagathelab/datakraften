package editors

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
)

type EditorStatus struct {
	Name    string
	Command string
	Installed bool
	Path    string
	WindowsSide bool
	Message string
}

func DetectVSCode() EditorStatus {
	path := exec.CommandPath("code")
	installed := path != ""
	return EditorStatus{
		Name:      "VS Code",
		Command:   "code",
		Installed: installed,
		Path:      path,
		WindowsSide: installed && strings.HasPrefix(path, "/mnt/"),
		Message:  editorMessage("VS Code", installed, path),
	}
}

func DetectZed() EditorStatus {
	path := exec.CommandPath("zed")
	installed := path != ""
	return EditorStatus{
		Name:      "Zed",
		Command:   "zed",
		Installed: installed,
		Path:      path,
		WindowsSide: installed && strings.HasPrefix(path, "/mnt/"),
		Message:  editorMessage("Zed", installed, path),
	}
}

func DetectCursor() EditorStatus {
	path := exec.CommandPath("cursor")
	installed := path != ""
	return EditorStatus{
		Name:      "Cursor",
		Command:   "cursor",
		Installed: installed,
		Path:      path,
		WindowsSide: installed && strings.HasPrefix(path, "/mnt/"),
		Message:  editorMessage("Cursor", installed, path),
	}
}

func DetectAll() []EditorStatus {
	return []EditorStatus{
		DetectVSCode(),
		DetectZed(),
		DetectCursor(),
	}
}

func editorMessage(name string, installed bool, path string) string {
	if !installed {
		return fmt.Sprintf("%s CLI not found", name)
	}
	if strings.HasPrefix(path, "/mnt/") {
		return fmt.Sprintf("%s found at %s (Windows-side, may need Linux CLI install)", name, path)
	}
	return ""
}
