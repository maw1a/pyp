package pip

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// PipManager represents a wrapper for pip package manager
type PipManager struct {
	pythonPath string
	pipPath    string
}

// NewPipManager creates a new instance of PipManager
func NewPipManager(pythonPath string) (*PipManager, error) {
	if pythonPath == "" {
		pythonPath = "python" // Default to system Python
	}

	// Check if Python is installed
	cmd := exec.Command(pythonPath, "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Python not found: %v", err)
	}

	// Get pip path
	cmd = exec.Command(pythonPath, "-m", "pip", "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pip not found: %v", err)
	}

	return &PipManager{
		pythonPath: pythonPath,
		pipPath:    "pip",
	}, nil
}

// Init initialize a Python project
func (pm *PipManager) Init(cmd *cobra.Command) error {
	// Check if environment & config is ready
	envReady, err := IsEnvReady()
	configReady, err := IsConfigReady()

	if configReady {
		if !envReady {
			if err := createEnv(pm.pythonPath); err != nil {
				return err
			}
		}
		fmt.Println("Project already initialized")
		return nil
	} else if !configReady && err == nil {
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		license, _ := cmd.Flags().GetString("license")

		if !cmd.Flags().Changed("name") {
			prompt := &survey.Input{Message: "What is your project name?", Default: "my_project"}
			if err := survey.AskOne(prompt, &name); err != nil {
				return err
			}
		}

		if !cmd.Flags().Changed("version") {
			prompt := &survey.Input{Message: "What is your project version?", Default: "0.1.0"}
			if err := survey.AskOne(prompt, &version); err != nil {
				return err
			}
		}

		if !cmd.Flags().Changed("license") {
			licenses := []string{"ISC", "MIT", "Apache-2.0"}
			prompt := &survey.Select{Message: "Choose a license:", Options: licenses, Default: "ISC"}
			if err := survey.AskOne(prompt, &license); err != nil {
				return err
			}
		}

		args := CreateConfigArgs{Name: name, Version: version, License: license}

		if err := createEnv(pm.pythonPath); err != nil {
			return err
		}

		if err := createConfig(pm, args); err != nil {
			return err
		}

		fmt.Println("Environment created successfully")

		return nil
	} else {
		return fmt.Errorf("failed to check environment: %v", err)
	}
}

// Install installs a Python package
func (pm *PipManager) Install(args ...string) error {
	if ready, err := IsEnvReady(); err != nil {
		return err
	} else if !ready {
		if err := createEnv(pm.pythonPath); err != nil {
			return err
		}
	}

	activate(VENV_DIR)
	defer deactivate(true)

	var (
		cfg  PypConfig
		pkgs []string
	)

	if len(args) == 0 {
		configFile, err := os.ReadFile(PYPCONFIG_TOML)
		if err != nil {
			return fmt.Errorf("failed to read pypconfig.toml: %v", err)
		}
		if _, err := toml.Decode(string(configFile), &cfg); err != nil {
			return fmt.Errorf("failed to decode pypconfig.toml: %v", err)
		}

		pkgs = cfg.Project.Dependencies
	} else {
		pkgs = args
	}

	var cmd *exec.Cmd
	cmd = exec.Command(pm.pipPath, append([]string{"install"}, pkgs...)...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package: %v, stderr: %s", err, stderr.String())
	}

	return updateConfigDeps(pm)
}

// Uninstall removes a Python package
func (pm *PipManager) Uninstall(packageName string) error {
	activate(VENV_DIR)
	defer deactivate(true)

	cmd := exec.Command(pm.pipPath, "uninstall", "-y", packageName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall package: %v, stderr: %s", err, stderr.String())
	}

	return updateConfigDeps(pm)
}

// ListInstalled returns a list of installed packages
func (pm *PipManager) ListInstalled() ([]string, error) {
	activate(VENV_DIR)
	defer deactivate(true)

	cmd := exec.Command(pm.pipPath, "list", "--format=freeze")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list packages: %v, stderr: %s", err, stderr.String())
	}

	packages := strings.Split(stdout.String(), "\n")
	var result []string
	for _, pkg := range packages {
		if pkg != "" {
			result = append(result, pkg)
		}
	}
	return result, nil
}

// RunScript runs a pyp script pypconfig.toml
func (pm *PipManager) RunScript(script string) error {
	activate(VENV_DIR)
	defer deactivate(true)

	var cfg PypConfig

	configFile, err := os.ReadFile(PYPCONFIG_TOML)
	if err != nil {
		return fmt.Errorf("failed to read pypconfig.toml: %v", err)
	}
	if _, err := toml.Decode(string(configFile), &cfg); err != nil {
		return fmt.Errorf("failed to decode pypconfig.toml: %v", err)
	}

	cmnd := cfg.Scripts[script]
	if cmnd == "" {
		return fmt.Errorf("script '%s' not found in pypconfig.toml", script)
	}

	parts := strings.Split(cmnd, " ")
	nm := parts[0]
	arg := parts[1:]

	cmd := exec.Command(nm, arg...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", cmnd)
	} else {
		cmd = exec.Command("bash", "-c", cmnd)
	}

	output, err := cmd.CombinedOutput()

	fmt.Print(string(output))

	return nil
}

// IsInstalled checks if a package is installed
func (pm *PipManager) IsInstalled(packageName string) (bool, error) {
	activate(VENV_DIR)
	defer deactivate(true)

	installed, err := pm.ListInstalled()
	if err != nil {
		return false, err
	}

	for _, pkg := range installed {
		if strings.EqualFold(strings.Split(pkg, "==")[0], packageName) {
			return true, nil
		}
	}
	return false, nil
}

// Update upgrades a Python package to the latest version
func (pm *PipManager) Update(packageName string) error {
	activate(VENV_DIR)
	defer deactivate(true)

	cmd := exec.Command(pm.pipPath, "install", "--upgrade", packageName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade package: %v, stderr: %s", err, stderr.String())
	}

	return updateConfigDeps(pm)
}
