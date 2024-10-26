package pip

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Environment variables backup structure
type EnvBackup struct {
	OldPath        string
	OldPythonHome  string
	OldPS1         string
	VirtualEnvPath string
}

var backup EnvBackup

func deactivate(nondestructive bool) error {
	// Restore PATH
	if backup.OldPath != "" {
		os.Setenv("PATH", backup.OldPath)
	}

	// Restore PYTHONHOME
	if backup.OldPythonHome != "" {
		os.Setenv("PYTHONHOME", backup.OldPythonHome)
	} else {
		os.Unsetenv("PYTHONHOME")
	}

	// Clear virtual environment
	os.Unsetenv("VIRTUAL_ENV")

	// Reset hash in shell if using bash or zsh
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") || strings.Contains(shell, "zsh") {
		cmd := exec.Command("hash", "-r")
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Clear backup if not in nondestructive mode
	if !nondestructive {
		backup = EnvBackup{}
	}

	return nil
}

func activate(envDir string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	envPath := filepath.Join(wd, envDir)
	// Deactivate any existing environment first
	deactivate(true)

	// Backup current environment
	backup = EnvBackup{
		OldPath:        os.Getenv("PATH"),
		OldPythonHome:  os.Getenv("PYTHONHOME"),
		OldPS1:         os.Getenv("PS1"),
		VirtualEnvPath: envPath,
	}

	// Set VIRTUAL_ENV
	os.Setenv("VIRTUAL_ENV", envPath)

	// Update PATH
	binPath := filepath.Join(envPath, "bin")
	newPath := binPath + string(os.PathListSeparator) + os.Getenv("PATH")
	os.Setenv("PATH", newPath)

	// Unset PYTHONHOME
	os.Unsetenv("PYTHONHOME")

	// Reset hash in shell if using bash or zsh
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") || strings.Contains(shell, "zsh") {
		cmd := exec.Command("hash", "-r")
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	return nil
}

func GetFirstValidPythonPath() (string, error) {
	cmd := exec.Command("which", "-a", "python", "python3")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute 'which' command: %v", err)
	}

	paths := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, path := range paths {
		if isValidPythonPath(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("no valid Python path found")
}

func isValidPythonPath(path string) bool {
	cmd := exec.Command(path, "--version")
	err := cmd.Run()
	return err == nil
}
