package config

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"time"
// 	"github.com/BurntSushi/toml"
// 	"vanish/internal/types"
// )

// func ExportConfig(exportPath string) error {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return fmt.Errorf("failed to get home directory: %v", err)
// 	}

// 	// Path to current config (fix directory consistency)
// 	configDir := filepath.Join(homeDir, ".config", "vanish")
// 	configPath := filepath.Join(configDir, "vanish.toml")

// 	// Check if config file exists
// 	if _, err := os.Stat(configPath); os.IsNotExist(err) {
// 		return fmt.Errorf("config file does not exist: %s", configPath)
// 	}

// 	// Create backup in the same config directory
// 	backupPath := filepath.Join(configDir, "vanish-bck.toml")
// 	if err := copyConfigFile(configPath, backupPath); err != nil {
// 		return fmt.Errorf("failed to create backup: %v", err)
// 	}
// 	fmt.Printf("Config backed up to: %s\n", backupPath)

// 	// Determine export path
// 	finalExportPath := exportPath

// 	// Check if exportPath is a directory
// 	if info, err := os.Stat(exportPath); err == nil && info.IsDir() {
// 		// Append default filename if directory provided
// 		finalExportPath = filepath.Join(exportPath, "vanish.toml")
// 	} else {
// 		// Make sure parent directory exists
// 		parentDir := filepath.Dir(exportPath)
// 		if err := os.MkdirAll(parentDir, 0755); err != nil {
// 			return fmt.Errorf("failed to create export directory: %v", err)
// 		}
// 	}

// 	// Copy config file to export path
// 	if err := copyConfigFile(configPath, finalExportPath); err != nil {
// 		return fmt.Errorf("failed to export config: %v", err)
// 	}

// 	fmt.Printf("Configuration exported successfully to: %s\n", finalExportPath)
// 	return nil
// }

// // importConfig imports configuration from a specified file
// func ImportConfig(importPath string) error {
// 	// Check if import file exists
// 	if _, err := os.Stat(importPath); os.IsNotExist(err) {
// 		return fmt.Errorf("import file does not exist: %s", importPath)
// 	}

// 	// Get current config path (fix directory consistency)
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return fmt.Errorf("failed to get home directory: %v", err)
// 	}

// 	configPath := filepath.Join(homeDir, ".config", "vanish", "vanish.toml") // Fixed to match ExportConfig
// 	configDir := filepath.Dir(configPath)

// 	// Create config directory if it doesn't exist
// 	if err := os.MkdirAll(configDir, 0755); err != nil {
// 		return fmt.Errorf("failed to create config directory: %v", err)
// 	}

// 	// Create backup of existing config if it exists
// 	if _, err := os.Stat(configPath); err == nil {
// 		timestamp := time.Now().Format("2006-01-02-15-04-05")
// 		backupPath := filepath.Join(configDir, fmt.Sprintf("vanish.toml.backup-%s", timestamp))

// 		// Copy current config to backup
// 		if err := copyConfigFile(configPath, backupPath); err != nil {
// 			return fmt.Errorf("failed to create config backup: %v", err)
// 		}
// 		fmt.Printf("Existing config backed up to: %s\n", backupPath)
// 	}

// 	// Test the import file by trying to parse it
// 	var testConfig types.Config
// 	if _, err := toml.DecodeFile(importPath, &testConfig); err != nil {
// 		return fmt.Errorf("invalid config file format: %v", err)
// 	}

// 	// Copy import file to config location
// 	if err := copyConfigFile(importPath, configPath); err != nil {
// 		return fmt.Errorf("failed to import config: %v", err)
// 	}

// 	fmt.Printf("Configuration imported successfully from: %s\n", importPath)
// 	fmt.Printf("New config location: %s\n", configPath)

// 	// Validate the imported config by loading it
// 	if _, err := LoadConfig(); err != nil {
// 		return fmt.Errorf("imported config is invalid: %v", err)
// 	}

// 	fmt.Println("Imported configuration validated successfully!")
// 	return nil
// }

// // Enhanced version with better error handling and options
// func ExportConfigWithOptions(exportPath string, createBackup bool, backupName string) error {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return fmt.Errorf("failed to get home directory: %v", err)
// 	}

// 	configDir := filepath.Join(homeDir, ".config", "vanish")
// 	configPath := filepath.Join(configDir, "vanish.toml")

// 	// Check if config file exists
// 	if _, err := os.Stat(configPath); os.IsNotExist(err) {
// 		return fmt.Errorf("config file does not exist: %s", configPath)
// 	}

// 	// Create backup if requested
// 	if createBackup {
// 		if backupName == "" {
// 			backupName = "vanish-bck.toml"
// 		}
// 		backupPath := filepath.Join(configDir, backupName)

// 		if err := copyConfigFile(configPath, backupPath); err != nil {
// 			return fmt.Errorf("failed to create backup: %v", err)
// 		}
// 		fmt.Printf("Config backed up to: %s\n", backupPath)
// 	}

// 	// Handle export path
// 	finalExportPath := exportPath
// 	if info, err := os.Stat(exportPath); err == nil && info.IsDir() {
// 		finalExportPath = filepath.Join(exportPath, "vanish.toml")
// 	} else {
// 		if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
// 			return fmt.Errorf("failed to create export directory: %v", err)
// 		}
// 	}

// 	// Export the config
// 	if err := copyConfigFile(configPath, finalExportPath); err != nil {
// 		return fmt.Errorf("failed to export config: %v", err)
// 	}

// 	fmt.Printf("Configuration exported successfully to: %s\n", finalExportPath)
// 	return nil
// }

// // Helper function to copy config files
// func copyConfigFile(src, dst string) error {
// 	// Read source file
// 	data, err := os.ReadFile(src)
// 	if err != nil {
// 		return err
// 	}

// 	// Write to destination
// 	return os.WriteFile(dst, data, 0644)
// }
