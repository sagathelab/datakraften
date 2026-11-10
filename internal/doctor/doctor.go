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
	critical, warnings, info, passed := 0, 0, 0, 0

	for _, c := range r.Checks {
		if c.Status == "pass" {
			passed++
		} else if c.Status == "fail" {
			switch c.Severity {
			case "critical":
				critical++
			case "warning":
				warnings++
			default:
				info++
			}
		}
	}
	fmt.Println("  Summary")
	fmt.Printf("    %d passed  ·  %d critical  ·  %d warnings  ·  %d informational\n", passed, critical, warnings, info)
}

func (r *Report) PrintJSON() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}
