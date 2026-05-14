package doctor

import (
	"encoding/json"
	"fmt"
	"os"
)

type Check struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

type Report struct {
	Checks []Check `json:"checks"`
}

func (r *Report) Add(c Check) {
	r.Checks = append(r.Checks, c)
}

func (r *Report) Print() {
	critical := 0
	warnings := 0

	for _, c := range r.Checks {
		if c.Status == "fail" && c.Severity == "critical" {
			critical++
		} else if c.Status == "fail" {
			warnings++
		}
	}

	if critical == 0 {
		fmt.Println("  ✓ No critical issues found")
	} else {
		fmt.Printf("  ✗ %d critical issue(s) found\n", critical)
	}

	if warnings == 0 {
		fmt.Println("  ✓ No warnings")
	} else {
		fmt.Printf("  ⚠ %d warning(s) found\n", warnings)
	}
}

func (r *Report) PrintJSON() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}
