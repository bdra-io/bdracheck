package main

import (
	"encoding/json"
	"flag"
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
	// Define supported CLI flags
	configFlag := flag.String("config", "bdracheck.json", "Path to the BDRA architecture governance configuration file")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || args[0] != "verify" {
		fmt.Println("Usage: bdracheck verify [--config=path/to/bdracheck.json]")
		os.Exit(1)
	}

	fmt.Printf("🚀 BDRA Static Analysis Linter Engine Engine Initializing using [%s]...\n", *configFlag)

	// 1. Load the rule matrix configuration
	configData, err := os.ReadFile(*configFlag)
	if err != nil {
		fmt.Printf("❌ Fatal: Failed to read compliance configuration file (%s): %v\n", *configFlag, err)
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("❌ Fatal: Malformed JSON parsing profile schema: %v\n", err)
		os.Exit(1)
	}

	// 2. Discover and harvest target Go source targets dynamically inside internal/
	var goFiles []string
	searchDir := "internal"
	
	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Target validation directory '%s' not found in current execution path.\n", searchDir)
		os.Exit(1)
	}

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
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