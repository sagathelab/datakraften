package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func newUpgradeCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dk CLI to the latest version",
		Long:  `Download and install the latest Datakraften CLI binary from GitHub Releases. Verifies SHA256 checksum before replacing the current binary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("cannot determine executable path: %w", err)
			}

			if !jsonOutput {
				displayTitle(os.Stdout, "upgrade", "Checking for a new dk release")
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
				fmt.Println("  Release")
				fmt.Printf("    Current: %s\n", version)
				fmt.Printf("    Latest:  %s\n", release.TagName)
				fmt.Println()
			}

			if version == release.TagName || version == "dev" {
				if jsonOutput {
					printUpdateJSON(version, release.TagName, false, "", 0, nil)
				} else {
					if version == "dev" && version != release.TagName {
						fmt.Println("  [SKIP] Development build; upgrade skipped.")
					} else {
						fmt.Println("  [OK] Already up to date.")
					}
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

			if checksumURL == "" {
				return fmt.Errorf("no checksum found for %s", binaryName)
			}
			if dryRun {
				if jsonOutput {
					printUpdateJSON(version, release.TagName, false, binaryName, 0, nil)
				} else {
					fmt.Printf("  [SKIP] Dry run: would upgrade to %s (%s).\n", release.TagName, binaryName)
				}
				return nil
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

			written, err := installUpgrade(client, token, checksumURL, execPath, binResp.Body)
			if err != nil {
				return err
			}

			if jsonOutput {
				printUpdateJSON(version, release.TagName, true, binaryName, written, nil)
			} else {
				fmt.Printf("  [OK] Updated to %s (%d bytes)\n", release.TagName, written)
				fmt.Println()
				fmt.Println("  Next steps")
				fmt.Println("    Run 'dk doctor' to verify your setup.")
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be upgraded without making changes")
	return cmd
}

func installUpgrade(client *http.Client, token, checksumURL, execPath string, binary io.Reader) (int64, error) {
	req, err := newGitHubRequest(http.MethodGet, checksumURL, token)
	if err != nil {
		return 0, fmt.Errorf("cannot create checksum request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot download checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("cannot download checksum: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(data) > 4096 {
		return 0, fmt.Errorf("cannot read checksum")
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("empty checksum")
	}
	expected, err := hex.DecodeString(fields[0])
	if err != nil || len(expected) != sha256.Size {
		return 0, fmt.Errorf("invalid SHA256 checksum")
	}
	f, err := os.CreateTemp(filepath.Dir(execPath), ".dk-upgrade-*")
	if err != nil {
		return 0, fmt.Errorf("cannot create temp file: %w", err)
	}
	defer os.Remove(f.Name())
	hash := sha256.New()
	written, copyErr := io.Copy(f, io.TeeReader(binary, hash))
	if copyErr == nil {
		copyErr = f.Chmod(0755)
	}
	closeErr := f.Close()
	if copyErr != nil {
		return 0, fmt.Errorf("cannot write upgrade: %w", copyErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("cannot close upgrade: %w", closeErr)
	}
	if !strings.EqualFold(fields[0], hex.EncodeToString(hash.Sum(nil))) {
		return 0, fmt.Errorf("checksum mismatch")
	}
	if err := os.Rename(f.Name(), execPath); err != nil {
		return 0, fmt.Errorf("cannot replace binary: %w", err)
	}
	return written, nil
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
