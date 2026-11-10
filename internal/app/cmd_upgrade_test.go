package app

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallUpgrade(t *testing.T) {
	binary := "new binary"
	sum := sha256.Sum256([]byte(binary))
	tests := []struct {
		name      string
		status    int
		checksum  string
		wantError bool
	}{
		{"valid", http.StatusOK, fmt.Sprintf("%x  dk\n", sum), false},
		{"missing", http.StatusNotFound, "not found", true},
		{"empty", http.StatusOK, "", true},
		{"malformed", http.StatusOK, "not-a-checksum", true},
		{"mismatch", http.StatusOK, strings.Repeat("0", 64), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.checksum)
			}))
			defer server.Close()
			dir := t.TempDir()
			path := filepath.Join(dir, "dk")
			if err := os.WriteFile(path, []byte("old binary"), 0755); err != nil {
				t.Fatal(err)
			}
			written, err := installUpgrade(server.Client(), "", server.URL, path, strings.NewReader(binary))
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %v", err, tt.wantError)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := binary
			if tt.wantError {
				want = "old binary"
			} else if written != int64(len(binary)) {
				t.Fatalf("written = %d", written)
			}
			if string(got) != want {
				t.Fatalf("binary = %q, want %q", got, want)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0755 {
				t.Fatalf("permissions = %o", info.Mode().Perm())
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("temporary files left behind: %v", entries)
			}
		})
	}
}

func TestInstallUpgradeRenameFailurePreservesDestination(t *testing.T) {
	binary := "new binary"
	sum := sha256.Sum256([]byte(binary))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%x", sum)
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "dk")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "keep"), []byte("safe"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := installUpgrade(server.Client(), "", server.URL, path, strings.NewReader(binary)); err == nil {
		t.Fatal("expected rename error")
	}
	if _, err := os.Stat(filepath.Join(path, "keep")); err != nil {
		t.Fatalf("destination damaged: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}
