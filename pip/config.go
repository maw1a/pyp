package pip

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

// PypConfig represents fields of pypconfig.toml
type (
	CreateConfigArgs struct {
		Name    string
		Version string
		License string
	}

	PypConfig_Python struct {
		Version string `toml:"version"`
		Pip     string `toml:"pip"`
	}

	PypConfig_Project struct {
		Name        string `toml:"name"`
		Version     string `toml:"version"`
		Description string `toml:"description"`
		Homepage    string `toml:"homepage,omitempty"`
		License     string `toml:"license"`

		Authors     []string `toml:"authors"`
		Maintainers []string `toml:"maintainers"`

		Dependencies   []string `toml:"dependencies"`
		RequiresPython string   `toml:"requires-python,omitempty"`
	}

	PypConfig_Scripts map[string]string

	PypConfig_Pyp struct {
		Version string `toml:"version"`
	}

	PypConfig struct {
		Python  PypConfig_Python  `toml:"python"`
		Project PypConfig_Project `toml:"project"`
		Scripts PypConfig_Scripts `toml:"scripts"`
		Pyp     PypConfig_Pyp     `toml:"pyp"`
	}
)

// ! CONSTANTS
const (
	PYP_VERSION    = "0.1.0"
	VENV_DIR       = ".env"
	PYPCONFIG_TOML = "pypconfig.toml"
)

// Get pyp version
func PypVersion() error {
	fmt.Println("pyp CLI", PYP_VERSION)

	return nil
}

// Check if the .env environment is initialized
func IsEnvReady() (bool, error) {
	_, err := os.Stat(VENV_DIR)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to check environment: %v", err)
	}
}

// Check if pypconfig.toml is created
func IsConfigReady() (bool, error) {
	_, err := os.Stat(PYPCONFIG_TOML)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to find pypconfig.toml: %v", err)
	}
}

// Create .env environment
func createEnv(pythonPath string) error {
	cmd := exec.Command(pythonPath, "-m", "venv", VENV_DIR)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create environment: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// Create pypconfig.toml
func createConfig(pm *PipManager, args CreateConfigArgs) error {
	var (
		config        PypConfig
		stdout        bytes.Buffer
		pythonVersion string
		pipVersion    string
	)

	cmd := exec.Command(pm.pythonPath, "--version")
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return err
	}
	pythonVersion = strings.Split(strings.Trim(stdout.String(), "\n"), " ")[1]

	stdout.Reset()
	cmd = exec.Command(pm.pythonPath, "-m", "pip", "--version")
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return err
	}
	pipVersion = strings.Split(strings.Trim(stdout.String(), "\n"), " ")[1]

	pkgs, err := pm.ListInstalled()
	if err != nil {
		return err
	}

	config = PypConfig{
		Python: PypConfig_Python{
			Version: pythonVersion,
			Pip:     pipVersion,
		},
		Project: PypConfig_Project{
			Name:         args.Name,
			Version:      args.Version,
			Description:  "",
			License:      args.License,
			Dependencies: pkgs,
		},
		Scripts: PypConfig_Scripts{
			"test": "echo \"Error: no test specified\" && exit 1",
		},
		Pyp: PypConfig_Pyp{
			Version: PYP_VERSION,
		},
	}

	file, err := os.Create(PYPCONFIG_TOML)
	if err != nil {
		fmt.Printf("failed to create config file: %s", err)
		return err
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(config); err != nil {
		fmt.Printf("failed to encode config to file: %s", err)
		return err
	}

	fmt.Println("pypconfig.toml created successfully")

	return nil
}

// Update dependencies in pypconfig.toml
func updateConfigDeps(pm *PipManager) error {
	pkgs, err := pm.ListInstalled()
	if err != nil {
		return err
	}

	var cfg PypConfig
	configFile, err := os.ReadFile(PYPCONFIG_TOML)
	if err != nil {
		return fmt.Errorf("failed to read pypconfig.toml: %v", err)
	}
	if _, err := toml.Decode(string(configFile), &cfg); err != nil {
		return fmt.Errorf("failed to decode pypconfig.toml: %v", err)
	}

	cfg.Project.Dependencies = pkgs

	file, err := os.Create(PYPCONFIG_TOML)
	if err != nil {
		fmt.Printf("failed to create config file: %s", err)
		return err
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(cfg); err != nil {
		fmt.Printf("failed to encode config to file: %s", err)
		return err
	}

	return nil
}
