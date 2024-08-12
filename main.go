package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var user = os.Getenv("USERNAME")
var backupPath = fmt.Sprintf("C:\\Users\\%s\\OneDrive\\Backup", user)

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
		// Development,

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
		// "Development",

		// direct files group
		".autobump.yaml",
		".freterc",
		".gitconfig",
		".gitignore",
		".npmrc",
		".npmrc.vizir",
		".p10k.zsh",
		".zshrc",
		"pyvenv.cfg",
	},
}

var excludedFolders = []string{
	".venv",
	"node_modules",
}

func copyFile(sourcePath, destinationPath string) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		fmt.Println(err)
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
	cleanSourcePath := strings.ReplaceAll(sourcePath, "\x00", "")
	info, err := os.Stat(cleanSourcePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if info.IsDir() {
		if !shouldProcessDirectory(cleanSourcePath, excludedFolders) {
			fmt.Printf("Skipping folder %s\n", cleanSourcePath)
			return
		}

		if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
			os.MkdirAll(destinationPath, os.ModePerm)
		}
		entries, err := os.ReadDir(cleanSourcePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, entry := range entries {
			copyItemRecursively(filepath.Join(cleanSourcePath, entry.Name()), filepath.Join(destinationPath, entry.Name()), operation)
		}
		fmt.Printf("%s folder %s to %s\n", operation, sourcePath, destinationPath)
	} else {
		parentFolder := filepath.Dir(destinationPath)
		if _, err := os.Stat(parentFolder); os.IsNotExist(err) {
			os.MkdirAll(parentFolder, os.ModePerm)
		}
		copyFile(cleanSourcePath, destinationPath)
		fmt.Printf("%s file %s to %s\n", operation, sourcePath, destinationPath)
	}
}

func shouldProcessDirectory(directory string, excludedFolders []string) bool {
	currentDir := directory
	for currentDir != "" {
		for _, excluded := range excludedFolders {
			if filepath.Base(currentDir) == excluded {
				return false
			}
		}
		currentDir = filepath.Dir(currentDir)
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

func main() {
	var rootCmd = &cobra.Command{Use: "backup-restore"}

	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup files",
		Run: func(cmd *cobra.Command, args []string) {
			copyFiles(fmt.Sprintf("C:\\Users\\%s", user), filepath.Join(backupPath, "win"), paths["win"], "Backed up")
			copyFiles(fmt.Sprintf("\\\\wsl.localhost\\%s\\home\\%s", defaultWSL, user), filepath.Join(backupPath, "wsl"), paths["wsl"], "Backed up")
		},
	}

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore files",
		Run: func(cmd *cobra.Command, args []string) {
			copyFiles(filepath.Join(backupPath, "win"), fmt.Sprintf("C:\\Users\\%s", user), paths["win"], "Restored")
			copyFiles(filepath.Join(backupPath, "wsl"), fmt.Sprintf("\\\\wsl.localhost\\%s\\home\\%s", defaultWSL, user), paths["wsl"], "Restored")
		},
	}

	rootCmd.AddCommand(backupCmd, restoreCmd)
	rootCmd.Execute()
}
