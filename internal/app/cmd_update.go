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
	"time"

	"github.com/spf13/cobra"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type githubAPIError struct {
	Message string `json:"message"`
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

			client := &http.Client{Timeout: 30 * time.Second}
			token := githubToken()

			req, err := newGitHubRequest(http.MethodGet, "https://api.github.com/repos/sagathelab/datakraften/releases/latest", token)
			if err != nil {
				return fmt.Errorf("cannot create release request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("cannot fetch latest release: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				message := parseGitHubAPIMessage(resp.Body)
				if message == "" {
					message = resp.Status
				}
				if resp.StatusCode == http.StatusNotFound {
					return fmt.Errorf("cannot fetch latest release (%s): %s. If this is a private repository, set GH_TOKEN or GITHUB_TOKEN", resp.Status, message)
				}
				return fmt.Errorf("cannot fetch latest release (%s): %s", resp.Status, message)
			}

			var release githubRelease
			if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
				return fmt.Errorf("cannot parse release info: %w", err)
			}
			if release.TagName == "" {
				return fmt.Errorf("latest release response did not include a tag name")
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

			binReq, err := newGitHubRequest(http.MethodGet, binaryURL, token)
			if err != nil {
				return fmt.Errorf("cannot create binary download request: %w", err)
			}
			binResp, err := client.Do(binReq)
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}
			defer binResp.Body.Close()
			if binResp.StatusCode != http.StatusOK {
				message := parseGitHubAPIMessage(binResp.Body)
				if message == "" {
					message = binResp.Status
				}
				return fmt.Errorf("download failed (%s): %s", binResp.Status, message)
			}

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
				chkReq, err := newGitHubRequest(http.MethodGet, checksumURL, token)
				if err == nil {
					chkResp, err := client.Do(chkReq)
					if err == nil {
						defer chkResp.Body.Close()
						if chkResp.StatusCode == http.StatusOK {
							chkBytes, _ := io.ReadAll(chkResp.Body)
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

func githubToken() string {
	if token := strings.TrimSpace(os.Getenv("GH_TOKEN")); token != "" {
		return token
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

func newGitHubRequest(method, url, token string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "datakraften-dk")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func parseGitHubAPIMessage(r io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return ""
	}

	var apiErr githubAPIError
	if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Message) != "" {
		return strings.TrimSpace(apiErr.Message)
	}

	return strings.TrimSpace(string(body))
}
