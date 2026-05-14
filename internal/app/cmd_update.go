package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update dk to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("cannot determine executable path: %w", err)
			}

			fmt.Println("  Checking for updates...")

			resp, err := http.Get("https://api.github.com/repos/sagathelab/datakraften/releases/latest")
			if err != nil {
				return fmt.Errorf("cannot fetch latest release: %w", err)
			}
			defer resp.Body.Close()

			var release githubRelease
			if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
				return fmt.Errorf("cannot parse release info: %w", err)
			}

			fmt.Printf("  Latest version: %s\n", release.TagName)
			fmt.Printf("  Current version: %s\n", version)
			fmt.Println()

			if version == release.TagName || version == "dev" {
				fmt.Println("  ✓ Already up to date.")
				return nil
			}

			binaryName := fmt.Sprintf("dk-%s-%s", runtime.GOOS, runtime.GOARCH)
			checksumName := binaryName + ".sha256"

			var binaryURL, checksumURL string
			for _, a := range release.Assets {
				if a.Name == binaryName {
					binaryURL = a.BrowserDownloadURL
				}
				if a.Name == checksumName {
					checksumURL = a.BrowserDownloadURL
				}
			}

			if binaryURL == "" {
				return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
			}

			fmt.Printf("  Downloading %s...\n", binaryName)

			binResp, err := http.Get(binaryURL)
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}
			defer binResp.Body.Close()

			tmpPath := execPath + ".new"
			f, err := os.Create(tmpPath)
			if err != nil {
				return fmt.Errorf("cannot create temp file: %w", err)
			}

			hash := sha256.New()
			written, err := io.Copy(f, io.TeeReader(binResp.Body, hash))
			if err != nil {
				f.Close()
				os.Remove(tmpPath)
				return fmt.Errorf("download incomplete: %w", err)
			}
			f.Close()

			if checksumURL != "" {
				chkResp, err := http.Get(checksumURL)
				if err == nil {
					chkBytes, _ := io.ReadAll(chkResp.Body)
					chkResp.Body.Close()
					parts := strings.Fields(string(chkBytes))
					if len(parts) > 0 {
						expected := strings.TrimSpace(parts[0])
						got := fmt.Sprintf("%x", hash.Sum(nil))
						if got != expected {
							os.Remove(tmpPath)
							return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, got)
						}
					}
				}
			}

			if err := os.Rename(tmpPath, execPath); err != nil {
				if err := os.Remove(execPath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("cannot replace binary: %w", err)
				}
				if err := os.Rename(tmpPath, execPath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("cannot replace binary: %w", err)
				}
			}

			if err := os.Chmod(execPath, 0755); err != nil {
				return fmt.Errorf("cannot set permissions: %w", err)
			}

			fmt.Printf("  ✓ Updated to %s (%d bytes)\n", release.TagName, written)
			fmt.Println()
			fmt.Println("  Run 'dk doctor' to verify your setup.")

			return nil
		},
	}
}
