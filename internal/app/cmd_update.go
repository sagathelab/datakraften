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

			if !jsonOutput {
				fmt.Println("  Checking for updates...")
			}

			client := &http.Client{Timeout: 30 * time.Second}
			token := githubToken()

			apiURL := "https://api.github.com/repos/sagathelab/datakraften/releases/latest"
			if verbose {
				fmt.Fprintf(os.Stderr, "  [verbose] GET %s\n", apiURL)
				if token != "" {
					fmt.Fprintf(os.Stderr, "  [verbose] using GitHub token authentication\n")
				}
			}

			req, err := newGitHubRequest(http.MethodGet, apiURL, token)
			if err != nil {
				return fmt.Errorf("cannot create release request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("cannot fetch latest release: %w", err)
			}
			defer resp.Body.Close()

			if verbose {
				fmt.Fprintf(os.Stderr, "  [verbose] response: %s\n", resp.Status)
			}

			if resp.StatusCode != http.StatusOK {
				message := parseGitHubAPIMessage(resp.Body)
				if message == "" {
					message = resp.Status
				}
				if resp.StatusCode == http.StatusNotFound {
					if hasReleases, _ := repoHasReleases(client, token); !hasReleases {
						return fmt.Errorf("no releases found in repository — publish a release on GitHub first")
					}
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

			if !jsonOutput {
				fmt.Printf("  Latest version: %s\n", release.TagName)
				fmt.Printf("  Current version: %s\n", version)
				fmt.Println()
			}

			if version == release.TagName || version == "dev" {
				if jsonOutput {
					printUpdateJSON(version, release.TagName, false, "", 0, nil)
				} else {
					fmt.Println("  ✓ Already up to date.")
				}
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

			if verbose {
				fmt.Fprintf(os.Stderr, "  [verbose] binary: %s\n", binaryURL)
				if checksumURL != "" {
					fmt.Fprintf(os.Stderr, "  [verbose] checksum: %s\n", checksumURL)
				}
			}

			if !jsonOutput {
				fmt.Printf("  Downloading %s...\n", binaryName)
			}

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

			var checksumErr error
			hash := sha256.New()
			written, err := io.Copy(f, io.TeeReader(binResp.Body, hash))
			if err != nil {
				f.Close()
				os.Remove(tmpPath)
				return fmt.Errorf("download incomplete: %w", err)
			}
			f.Close()

			if checksumURL != "" {
				if verbose {
					fmt.Fprintf(os.Stderr, "  [verbose] verifying checksum...\n")
				}
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
									checksumErr = fmt.Errorf("checksum mismatch: expected %s, got %s", expected, got)
								} else if verbose {
									fmt.Fprintf(os.Stderr, "  [verbose] checksum verified ✓\n")
								}
							}
						}
					}
				}
			}

			if checksumErr != nil {
				return checksumErr
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

			if jsonOutput {
				printUpdateJSON(version, release.TagName, true, binaryName, written, nil)
			} else {
				fmt.Printf("  ✓ Updated to %s (%d bytes)\n", release.TagName, written)
				fmt.Println()
				fmt.Println("  Run 'dk doctor' to verify your setup.")
			}

			return nil
		},
	}
}

func repoHasReleases(client *http.Client, token string) (bool, error) {
	req, err := newGitHubRequest(http.MethodGet, "https://api.github.com/repos/sagathelab/datakraften/releases?per_page=1", token)
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	var releases []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return false, err
	}
	return len(releases) > 0, nil
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

type updateResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Updated        bool   `json:"updated"`
	Binary         string `json:"binary,omitempty"`
	BytesWritten   int64  `json:"bytes_written,omitempty"`
}

func printUpdateJSON(currentVer, latestVer string, updated bool, binary string, bytesWritten int64, err error) {
	result := updateResult{
		CurrentVersion: currentVer,
		LatestVersion:  latestVer,
		Updated:        updated,
		Binary:         binary,
		BytesWritten:   bytesWritten,
	}
	if err != nil {
		result.Binary = ""
		result.BytesWritten = 0
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
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
