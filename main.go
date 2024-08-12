package main

import (
	"bufio"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Get the current Windows user folder
var user = os.Getenv("USERNAME")
var backupPath = fmt.Sprintf("C:\\Users\\%s\\OneDrive\\Backup", user)

// Detect WSL distros and get the default one
func getDefaultWSL() (string, error) {
	cmd := exec.Command("wsl", "-l", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return parts[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("default WSL distribution not found")
}

var defaultWSL, _ = getDefaultWSL()

// Define the source and destination map with specific files
var paths = map[string][]string{
	"win": {
		// folders or recursive group
		".aws\\config",
		".aws\\credentials",
		".azure\\azureProfile.json",
		".azure\\service_principal_entries.json",
		".gnupg",
		".ssh",
		"AppData\\Local\\Packages\\Microsoft.WindowsTerminal_*\\LocalState\\settings.json",
		"Development",

		// direct files group
		".gitconfig",
		".gitignore",
		".oh-my-posh.json",
		".wakatime.cfg",
	},
	"wsl": {
		// folders or recursive group
		".docker\\config.json",
		".histdb",
		".john",
		".kube\\config",
		".kube\\config-files",
		".sqlmap",
		"Development",

		// direct files group
		".autobump.yaml",
		".freterc",
		".gitconfig",
		".gitignore",
		".npmrc",
		".npmrc.vizir",
		".p10k.zsh",
		".zshrc",
		"pyvenv.cfg", // TODO: do I really need to backup this file?
	},
}

var excludedFolders = []string{
	".idea",
	".terraform",
	".terragrunt-cache",
	".venv",
	".vs",
	"bin",
	"desktop.ini",
	"dist",
	"node_modules",
	"site-package",
	"vendor",
}

func copyFile(sourcePath, destinationPath string) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		logger.Error(err)
		return
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		logger.Error(err)
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		logger.Error(err)
	}
}

func copyFiles(source, destination string, items []string, operation string) {
	for _, item := range items {
		resolvedItem := resolveWildcardPath(filepath.Join(source, item))
		relativePath := strings.TrimPrefix(resolvedItem, source)
		sourcePath := filepath.Join(source, relativePath)
		destinationPath := filepath.Join(destination, relativePath)
		copyItemRecursively(sourcePath, destinationPath, operation)
	}
}

func copyItemRecursively(sourcePath, destinationPath, operation string) {
	// Remove null characters from the source path (Illegal characters in path)
	cleanSourcePath := strings.ReplaceAll(sourcePath, "\x00", "")
	info, err := os.Stat(cleanSourcePath)
	if err != nil {
		logger.Error(err)
		return
	}

	if info.IsDir() {
		if !shouldProcessDirectory(cleanSourcePath, excludedFolders) {
			logger.Warnf("Skipping folder %s", cleanSourcePath)
			return
		}

		// Ensure the destination folder exists
		if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
			_ = os.MkdirAll(destinationPath, os.ModePerm)
		}
		entries, err := os.ReadDir(cleanSourcePath)
		if err != nil {
			logger.Error(err)
			return
		}
		for _, entry := range entries {
			copyItemRecursively(filepath.Join(cleanSourcePath, entry.Name()), filepath.Join(destinationPath, entry.Name()), operation)
		}

		logger.Infof("%s folder %s to %s", operation, cleanSourcePath, destinationPath)
	} else {
		// Ensure the parent folder exists in the destination
		parentFolder := filepath.Dir(destinationPath)
		if _, err := os.Stat(parentFolder); os.IsNotExist(err) {
			_ = os.MkdirAll(parentFolder, os.ModePerm)
		}
		copyFile(cleanSourcePath, destinationPath)

		logger.Infof("%s file %s to %s", operation, cleanSourcePath, destinationPath)
	}
}

func shouldProcessDirectory(directory string, excludedFolders []string) bool {
	for _, excluded := range excludedFolders {
		if filepath.Base(directory) == excluded {
			return false
		}
	}
	return true
}

func resolveWildcardPath(path string) string {
	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return path
	}
	return matches[0]
}

// Perform the requested operation
func main() {
	logger.SetLevel(logger.DebugLevel)

	var rootCmd = &cobra.Command{
		Use: "backup-restore",
	}

	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup files",
		Run: func(cmd *cobra.Command, args []string) {
			// Backup WIN files from root user folder in Windows
			copyFiles(fmt.Sprintf("C:\\Users\\%s", user), filepath.Join(backupPath, "win"), paths["win"], "Backed up")

			// Backup WSL files from default WSL path
			copyFiles(fmt.Sprintf("\\\\wsl.localhost\\%s\\home\\%s", defaultWSL, user), filepath.Join(backupPath, "wsl"), paths["wsl"], "Backed up")
		},
	}

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore files",
		Run: func(cmd *cobra.Command, args []string) {
			// Restore WIN files to root user folder in Windows
			copyFiles(filepath.Join(backupPath, "win"), fmt.Sprintf("C:\\Users\\%s", user), paths["win"], "Restored")

			// Restore WSL files to default WSL path
			copyFiles(filepath.Join(backupPath, "wsl"), fmt.Sprintf("\\\\wsl.localhost\\%s\\home\\%s", defaultWSL, user), paths["wsl"], "Restored")
		},
	}

	rootCmd.AddCommand(backupCmd, restoreCmd)
	_ = rootCmd.Execute()
}

// TODO: download the dotfiles from the github.com/user/dotfiles repository, where user is the current user
// inject 1Password credentials in the ones that are not public
