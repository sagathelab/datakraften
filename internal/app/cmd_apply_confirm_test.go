package app

import (
	"io"
	"strings"
	"testing"
)

func TestConfirmApply(t *testing.T) {
	for _, tc := range []struct {
		input     string
		yes, fail bool
	}{
		{"\n", true, false}, {"yes\n", true, false}, {"no\n", false, false}, {"", false, true}, {"yes", true, false}, {"maybe\n", false, true},
	} {
		yes, err := confirmApply(strings.NewReader(tc.input), io.Discard)
		if yes != tc.yes || (err != nil) != tc.fail {
			t.Errorf("input %q: yes=%v err=%v", tc.input, yes, err)
		}
	}
}
