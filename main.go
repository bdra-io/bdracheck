package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ArchitectureConfig maps the incoming bdracheck.json contract matrix
type ArchitectureConfig struct {
	ProjectName string `json:"projectName"`
	RingRules   []Rule `json:"ringRules"`
	LayerRules  []Rule `json:"layerRules"`
}

type Rule struct {
	TargetSegment    string   `json:"targetSegment"`
	ForbiddenImports []string `json:"forbiddenImports"`
	ErrorMessage     string   `json:"errorMessage"`
}

type Violation struct {
	File   string
	Line   int
	Import string
	Reason string
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "verify" {
		fmt.Println("Usage: bdracheck verify")
		os.Exit(1)
	}

	// 1. RESOLVE LOCAL CONFIGURATION MATRICES
	configFile, err := os.ReadFile("bdracheck.json")
	if err != nil {
		fmt.Println("🛑 [BDRA-LITE] Governance Error: bdracheck.json configuration file not found in current execution folder.")
		os.Exit(1)
	}

	var config ArchitectureConfig
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Printf("🛑 [BDRA-LITE] Configuration Parse Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 [BDRA-LITE] Initializing AST inspection for project: [%s]...\n", config.ProjectName)
	
	fset := token.NewFileSet()
	var violations []Violation
	allRules := append(config.RingRules, config.LayerRules...)

	// 2. RECURSIVELY SWEEP INTERNAL LAYERS
	err = filepath.Walk("internal", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		normalizedPath := filepath.ToSlash(path)

		for _, imp := range node.Imports {
			if imp.Path == nil {
				continue
			}

			importPath := strings.Trim(imp.Path.Value, `"`)

			// Respect manual inline bypass parameters
			if imp.Comment != nil && strings.Contains(imp.Comment.Text(), "bdracheck:ignore") {
				continue
			}

			// 3. RUN EVALUATION GATES
			for _, rule := range allRules {
				if strings.Contains(normalizedPath, rule.TargetSegment) {
					for _, forbidden := range rule.ForbiddenImports {
						if strings.Contains(importPath, forbidden) {
							violations = append(violations, Violation{
								File:   path,
								Line:   fset.Position(imp.Pos()).Line,
								Import: importPath,
								Reason: rule.ErrorMessage,
							})
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Fatal code verification fault: %v\n", err)
		os.Exit(1)
	}

	// 4. ARCHITECTURE RESULTS EVALUATION REPORTING
	if len(violations) > 0 {
		fmt.Printf("\n🛑 [BDRA-LITE] Governance Gate FAILED: %d boundary violations detected!\n", len(violations))
		fmt.Println("--------------------------------------------------------------------------------")
		for _, v := range violations {
			fmt.Printf("📍 File:   %s (Line %d)\n", v.File, v.Line)
			fmt.Printf("⚠️ Illegal Import:  \"%s\"\n", v.Import)
			fmt.Printf("📝 Rule Broken:     %s\n\n", v.Reason)
		}
		os.Exit(1)
	}

	fmt.Println("✅ [BDRA-LITE] Governance Gate Passed: 0 architecture violations detected.")
}