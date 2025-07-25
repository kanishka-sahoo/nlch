// Package update provides auto-update functionality for nlch
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	RepoOwner = "kanishka-sahoo"
	RepoName  = "nlch"
	UpdateURL = "https://api.github.com/repos/" + RepoOwner + "/" + RepoName + "/releases/latest"
)

// Build version can be set during compilation
var BuildVersion = "0.3.0"

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// GetCurrentVersion returns the current version of nlch
func GetCurrentVersion() string {
	return BuildVersion
}

// CheckForUpdates checks if a newer version is available
func CheckForUpdates() (*Release, bool, error) {
	resp, err := http.Get(UpdateURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release info: %v", err)
	}

	currentVersion := "v" + GetCurrentVersion()
	hasUpdate := release.TagName != currentVersion

	return &release, hasUpdate, nil
}

// GetPlatformAssetName returns the asset name for the current platform
func GetPlatformAssetName() string {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Convert Go OS names to our naming convention
	switch osName {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	case "windows":
		osName = "windows"
	}

	binaryName := "nlch-" + osName + "-" + archName
	if osName == "windows" {
		binaryName += ".exe"
	}

	return binaryName
}

// DownloadUpdate downloads the latest version
func DownloadUpdate(release *Release) (string, error) {
	assetName := GetPlatformAssetName()

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no asset found for platform: %s", assetName)
	}

	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, assetName)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write update: %v", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile, 0755); err != nil {
			return "", fmt.Errorf("failed to make executable: %v", err)
		}
	}

	return tempFile, nil
}

// InstallUpdate replaces the current binary with the updated one
func InstallUpdate(updatePath string) error {
	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	// On Windows, we need to use a different approach
	if runtime.GOOS == "windows" {
		return installUpdateWindows(updatePath, currentExe)
	}

	return installUpdateUnix(updatePath, currentExe)
}

// installUpdateUnix handles update installation on Unix systems
func installUpdateUnix(updatePath, currentExe string) error {
	// Create backup
	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Replace current executable
	if err := copyFile(updatePath, currentExe); err != nil {
		// Restore backup on failure
		copyFile(backupPath, currentExe)
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	// Remove backup and temp file
	os.Remove(backupPath)
	os.Remove(updatePath)

	return nil
}

// installUpdateWindows handles update installation on Windows
func installUpdateWindows(updatePath, currentExe string) error {
	// On Windows, we can't replace a running executable directly
	// We'll create a batch script to do it after the process exits
	batchScript := `@echo off
timeout /t 2 /nobreak >nul
move /y "%s" "%s"
start "" "%s"
del "%%0"
`
	batchPath := filepath.Join(os.TempDir(), "nlch_update.bat")
	batchContent := fmt.Sprintf(batchScript, updatePath, currentExe, currentExe)

	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		return fmt.Errorf("failed to create update script: %v", err)
	}

	// Start the batch script and exit
	cmd := exec.Command("cmd", "/c", batchPath)
	cmd.Start()

	fmt.Println("Update will complete after nlch exits. Please restart nlch.")
	os.Exit(0)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// AutoUpdate performs an automatic update check and update
func AutoUpdate(force bool) error {
	fmt.Println("Checking for updates...")

	release, hasUpdate, err := CheckForUpdates()
	if err != nil {
		return fmt.Errorf("update check failed: %v", err)
	}

	if !hasUpdate && !force {
		fmt.Println("nlch is already up to date.")
		return nil
	}

	if hasUpdate {
		fmt.Printf("New version available: %s (current: v%s)\n", release.TagName, GetCurrentVersion())
	}

	if force || hasUpdate {
		fmt.Println("Downloading update...")
		updatePath, err := DownloadUpdate(release)
		if err != nil {
			return fmt.Errorf("download failed: %v", err)
		}

		fmt.Println("Installing update...")
		if err := InstallUpdate(updatePath); err != nil {
			return fmt.Errorf("installation failed: %v", err)
		}

		fmt.Printf("Successfully updated to %s!\n", release.TagName)
	}

	return nil
}

// ShouldCheckForUpdates returns true if we should check for updates
// This implements a simple time-based check (once per day)
func ShouldCheckForUpdates() bool {
	configDir, err := getConfigDir()
	if err != nil {
		return false
	}

	lastCheckFile := filepath.Join(configDir, ".last_update_check")

	info, err := os.Stat(lastCheckFile)
	if err != nil {
		// File doesn't exist, should check
		return true
	}

	// Check if last check was more than 24 hours ago
	return time.Since(info.ModTime()) > 24*time.Hour
}

// UpdateLastCheckTime updates the timestamp of the last update check
func UpdateLastCheckTime() {
	configDir, err := getConfigDir()
	if err != nil {
		return
	}

	lastCheckFile := filepath.Join(configDir, ".last_update_check")

	// Create config directory if it doesn't exist
	os.MkdirAll(configDir, 0755)

	// Touch the file to update timestamp
	file, err := os.Create(lastCheckFile)
	if err != nil {
		return
	}
	file.Close()
}

// getConfigDir returns the configuration directory for nlch
func getConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "nlch"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "nlch"), nil
}

// NotifyUpdateAvailable shows a subtle notification about available updates
func NotifyUpdateAvailable() {
	if !ShouldCheckForUpdates() {
		return
	}

	go func() {
		defer UpdateLastCheckTime()

		_, hasUpdate, err := CheckForUpdates()
		if err != nil {
			return // Silently fail for background checks
		}

		if hasUpdate {
			fmt.Fprintf(os.Stderr, "\n💡 A new version of nlch is available! Run 'nlch --update' to update.\n\n")
		}
	}()
}
