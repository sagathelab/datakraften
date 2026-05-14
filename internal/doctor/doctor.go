package doctor

import "fmt"

type Check struct {
	ID       string
	Title    string
	Category string
	Severity string // "critical", "warning", "info"
	Status   string // "pass", "fail", "skip"
	Message  string
	Fix      string
}

type Report struct {
	Checks []Check
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
