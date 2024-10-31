package cmd

import (
	"fmt"
	"os"
	"pyp/pip"

	"github.com/spf13/cobra"
)

var pythonPath string

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "pyp",
		Short: "A CLI tool to manage Python projects",
		Long:  `A Command Line Interface built to manage Python projects the simple way.`,
	}

	// Global flags
	// ! THE DEFAULT PYTHON PATH USED BY pyp. Defaults to "python3"
	rootCmd.PersistentFlags().StringVar(&pythonPath, "python", "python3", "Python executable path")

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "version of pyp cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pip.PypVersion()
		},
	}

	// Init command
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a Python project",
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}

			return pm.Init(cmd)
		},
	}
	initCmd.Flags().String("name", "my_project", "Project name")
	initCmd.Flags().String("version", "0.1.0", "Project version")
	initCmd.Flags().String("license", "ISC", "Project license")

	// Install command
	var installCmd = &cobra.Command{
		Use:     "install [package-name]",
		Aliases: []string{"add", "i", "in", "ins", "inst", "insta", "instal", "isnt", "isnta", "isntal", "isntall"},
		Short:   "Install a Python package",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			return pm.Install(args...)
		},
	}

	// Uninstall command
	var uninstallCmd = &cobra.Command{
		Use:     "uninstall [package-name]",
		Short:   "Uninstall a Python package",
		Aliases: []string{"unlink", "remove", "rm", "r", "un"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			return pm.Uninstall(args[0])
		},
	}

	// List command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List installed Python packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			packages, err := pm.ListInstalled()
			if err != nil {
				return err
			}
			for _, pkg := range packages {
				fmt.Println(pkg)
			}
			return nil
		},
	}

	// Check command
	var checkCmd = &cobra.Command{
		Use:   "check [package-name]",
		Short: "Check if a Python package is installed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			installed, err := pm.IsInstalled(args[0])
			if err != nil {
				return err
			}
			if installed {
				fmt.Printf("Package %s is installed\n", args[0])
			} else {
				fmt.Printf("Package %s is not installed\n", args[0])
			}
			return nil
		},
	}

	// Upgrade command
	var updateCmd = &cobra.Command{
		Use:     "update [package-name]",
		Short:   "Update a Python package to the latest version",
		Aliases: []string{"up", "upgrade", "udpate"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			return pm.Update(args[0])
		},
	}

	// Run Script command
	var runCmd = &cobra.Command{
		Use:     "run-script [script-name]",
		Short:   "Run a pyp script from pypconfig.toml",
		Aliases: []string{"rn", "run", "rum", "urn"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := pip.NewPipManager(pythonPath)
			if err != nil {
				return err
			}
			return pm.RunScript(args[0])
		},
	}

	// Add commands to root command
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
