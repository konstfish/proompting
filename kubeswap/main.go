package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var kubeConfigPath string
var backupDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}

	kubeConfigPath = filepath.Join(home, ".kube", "config")
	backupDir = filepath.Join(home, ".kube", "kubeswap_backups")
}

var rootCmd = &cobra.Command{
	Use:   "kubeswap",
	Short: "Swap kubectl config files",
	Long:  `kubeswap is a CLI tool to manage and swap between different kubectl config files.`,
}

var swapCmd = &cobra.Command{
	Use:   "swap [file]",
	Short: "Swap to a new kubectl config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		newConfigPath := args[0]
		if err := swapConfig(newConfigPath); err != nil {
			fmt.Println("Error swapping config:", err)
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backed up kubectl config files",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listConfigs(); err != nil {
			fmt.Println("Error listing configs:", err)
			os.Exit(1)
		}
	},
}

func swapConfig(newConfigPath string) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup current config
	backupPath := filepath.Join(backupDir, fmt.Sprintf("config_backup_%s", time.Now().Format("20060102_150405")))
	if err := copyFile(kubeConfigPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current config: %w", err)
	}

	// Copy new config to kubectl config path
	if err := copyFile(newConfigPath, kubeConfigPath); err != nil {
		return fmt.Errorf("failed to set new config: %w", err)
	}

	fmt.Printf("Successfully swapped to new config: %s\n", newConfigPath)
	fmt.Printf("Old config backed up to: %s\n", backupPath)
	return nil
}

func listConfigs() error {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	fmt.Println("Available kubectl config backups:")
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "config_backup_") {
			fmt.Println(file.Name())
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func main() {
	rootCmd.AddCommand(swapCmd)
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

