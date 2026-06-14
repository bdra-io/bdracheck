package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config mirrors the structural schema defined in bdracheck.json
type Config struct {
	ProjectName          string               `json:"projectName"`
	ArchitectureStrategy string               `json:"architectureStrategy"`
	Rules                RulesConfig          `json:"rules"`
}

type RulesConfig struct {
	DisallowExternalImportsInPure PureRuleConfig `json:"disallowExternalImportsInPure"`
	DisallowIOInProtected         IORuleConfig   `json:"disallowIOInProtected"`
	EnforceInwardDependencyFlow   FlowRuleConfig `json:"enforceInwardDependencyFlow"`
}

type PureRuleConfig struct {
	TargetDirs      []string `json:"targetDirs"`
	AllowedPrefixes []string `json:"allowedPrefixes"`
}

type IORuleConfig struct {
	TargetDirs        []string `json:"targetDirs"`
	ForbiddenPackages []string `json:"forbiddenPackages"`
}

type FlowRuleConfig struct {
	Rings []RingConfig `json:"rings"`
}

type RingConfig struct {
	ID                  string   `json:"id"`
	Path                string   `json:"path"`
	AllowedDependencies []string `json:"allowedDependencies,omitempty"`
}

func main() {
	fmt.Println("🚀 BDRA Static Analysis Linter Engine Engine Initializing...")

	// 1. Load local rule matrix configuration
	configFile := "bdracheck.json"
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("❌ Fatal: Failed to read compliance configuration file (%s): %v\n", configFile, err)
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("❌ Fatal: Malformed JSON parsing profile schema: %v\n", err)
		os.Exit(1)
	}

	// 2. Discover and harvest target Go source targets
	var goFiles []string
	err = filepath.Walk("internal", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Error scanning internal package nodes: %v\n", err)
		os.Exit(1)
	}

	// 3. Hand off the found source files to our AST Analyzer engine
	violationsFound := AnalyzeCodebase(goFiles, config)

	if violationsFound > 0 {
		fmt.Printf("\n🛑 Build Stopped: %d Architecture Boundary Violations Detected.\n", violationsFound)
		os.Exit(1)
	}

	fmt.Println("\n✅ Verification Successful: All Ring and Layer boundaries are perfectly clean.")
	os.Exit(0)
}