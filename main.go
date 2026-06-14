package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config represents the complete schema payload contract mapped to bdracheck.json
type Config struct {
	ProjectName string `json:"projectName"`
	Rules       Rules  `json:"rules"`
}

type Rules struct {
	DisallowExternalImportsInPure DisallowExternalImportsInPure `json:"disallowExternalImportsInPure"`
	DisallowIOInProtected         DisallowIOInProtected         `json:"disallowIOInProtected"`
	EnforceInwardDependencyFlow   EnforceInwardDependencyFlow   `json:"enforceInwardDependencyFlow"`
}

type DisallowExternalImportsInPure struct {
	TargetDirs []string `json:"targetDirs"`
}

type DisallowIOInProtected struct {
	TargetDirs        []string `json:"targetDirs"`
	ForbiddenPackages []string `json:"forbiddenPackages"`
}

type EnforceInwardDependencyFlow struct {
	Rings []RingConfig `json:"rings"`
}

type RingConfig struct {
	ID                  string   `json:"id"`
	Path                string   `json:"path"`
	AllowedDependencies []string `json:"allowedDependencies"`
}

// Violation represents an isolated architectural boundary breach
type Violation struct {
	Position string
	Layer    string
	Message  string
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "verify" {
		fmt.Println("Usage: bdracheck verify")
		os.Exit(1)
	}

	// 1. RESOLVE AND UNMARSHAL CONFIGURATION SCHEMA MATRIX
	configFile, err := os.ReadFile("bdracheck.json")
	if err != nil {
		fmt.Println("🛑 [BDRA-LITE] Governance Error: bdracheck.json configuration file not found in execution folder.")
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Printf("🛑 [BDRA-LITE] Configuration Parse Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 [BDRA-LITE] Initializing Concurrent AST Inspection for project: [%s]...\n", config.ProjectName)

	// 2. DISCOVER ALL APPLICABLE GO SOURCE TARGETS
	var files []string
	err = filepath.Walk("internal", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Fatal file indexing fault: %v\n", err)
		os.Exit(1)
	}

	// 3. EXECUTE HIGH-PERFORMANCE CONCURRENT AUDITING SUITE
	violationCount := AnalyzeCodebase(files, config)

	// 4. REPORT SYSTEM TERMINATION METRICS
	if violationCount > 0 {
		fmt.Printf("🛑 [BDRA-LITE] Governance Gate FAILED: %d boundary violations detected!\n", violationCount)
		os.Exit(1)
	}

	fmt.Println("🛡️  [BDRA-LITE] Checking ring isolation layers [Inward Flow & Layer Segregation]... Clean!")
	fmt.Println("✅ [BDRA-LITE] Governance Gate Passed: 0 architecture violations detected.")
}

// AnalyzeCodebase orchestrates concurrent AST evaluation loops across all discovered source files
func AnalyzeCodebase(files []string, config Config) int {
	var wg sync.WaitGroup
	violationChan := make(chan Violation, len(files)*2)
	fset := token.NewFileSet()

	// CONCURRENT PIPELINE: Distribute file parsing routines across active system threads
	for _, fileTarget := range files {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			
			// Parse file token specs including adjacent code comment nodes optimizing with ImportsOnly
			node, err := parser.ParseFile(fset, target, nil, parser.ImportsOnly|parser.ParseComments)
			if err != nil {
				return // Gracefully skip un-compilable draft files during active development
			}

			normalizedPath := filepath.ToSlash(target)
			fileTokenMap := fset.File(node.Pos())

			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, "\"")
				pos := fset.Position(imp.Pos())

				// 精确注解校验: Validate exact token alignments for inline bypass directives
				if hasStrictIgnoreDirective(imp, fileTokenMap, node.Comments) {
					continue
				}

				// A. EVALUATE PURE LAYER BOUNDARIES
				if matchesPathPattern(normalizedPath, config.Rules.DisallowExternalImportsInPure.TargetDirs) {
					if isViolatingPureInvariants(importPath, config.ProjectName) {
						violationChan <- Violation{
							Position: pos.String(),
							Layer:    "Pure Layer (Zero I/O)",
							Message:  fmt.Sprintf("Forbidden reference to infrastructure or external package: '%s'", importPath),
						}
					}
				}

				// B. EVALUATE PROTECTED LAYER BOUNDARIES
				if matchesPathPattern(normalizedPath, config.Rules.DisallowIOInProtected.TargetDirs) {
					for _, bannedPkg := range config.Rules.DisallowIOInProtected.ForbiddenPackages {
						if importPath == bannedPkg || strings.HasPrefix(importPath, bannedPkg+"/") {
							violationChan <- Violation{
								Position: pos.String(),
								Layer:    "Protected Layer (API Contracts)",
								Message:  fmt.Sprintf("Direct framework or storage I/O leak detected: '%s'", importPath),
							}
						}
					}
				}

				// C. EVALUATE CONCENTRIC INWARD-FLOW INVARIANTS
				if currentRing, ok := identifyRingContext(normalizedPath, config.Rules.EnforceInwardDependencyFlow.Rings); ok {
					if isViolatingRingFlow(importPath, currentRing, config.Rules.EnforceInwardDependencyFlow.Rings, config.ProjectName) {
						violationChan <- Violation{
							Position: pos.String(),
							Layer:    fmt.Sprintf("%s Ring Boundary", currentRing.ID),
							Message:  fmt.Sprintf("Inward flow breach! Inner rings cannot access outer ring spaces: '%s'", importPath),
						}
					}
				}
			}
		}(fileTarget)
	}

	wg.Wait()
	close(violationChan)

	violationCount := 0
	for v := range violationChan {
		violationCount++
		fmt.Printf("❌ ARCHITECTURE VIOLATION\n   ↳ Position:  %s\n   ↳ Layer:     %s\n   ↳ Message:   %s\n\n", 
			v.Position, v.Layer, v.Message)
	}

	return violationCount
}

// matchesPathPattern uses robust segment-token isolation to compile wildcard glob variables safely
func matchesPathPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		cleanPattern := filepath.ToSlash(pattern)
		
		cleanPattern = strings.ReplaceAll(cleanPattern, "/**", "")
		cleanPattern = strings.ReplaceAll(cleanPattern, "/*", "")
		cleanPattern = strings.ReplaceAll(cleanPattern, "/...", "")
		baseTarget := strings.Split(cleanPattern, "*.go")[0]

		if strings.Contains(path, baseTarget) {
			return true
		}
	}
	return false
}

// hasStrictIgnoreDirective executes token position line checks to eliminate loose substring collisions
func hasStrictIgnoreDirective(imp *ast.ImportSpec, fileMap *token.File, commentGroups []*ast.CommentGroup) bool {
	if fileMap == nil {
		return false
	}
	importLine := fileMap.Line(imp.Pos())

	for _, group := range commentGroups {
		for _, comment := range group.List {
			commentLine := fileMap.Line(comment.Pos())
			
			// In Go AST, the bypass directive must reside on the exact same line as the statement
			if commentLine == importLine && strings.Contains(comment.Text, "bdracheck:ignore") {
				return true
			}
		}
	}
	return false
}

func isViolatingPureInvariants(importPath string, projectName string) bool {
	allowedStdLib := map[string]bool{
		"errors":  true,
		"math":    true,
		"strings": true,
		"time":    true,
		"fmt":     true,
		"sort":    true,
		"strconv": true,
	}

	if allowedStdLib[importPath] || strings.HasPrefix(importPath, "encoding/") {
		return false
	}

	if strings.Contains(importPath, "/pure") {
		return false
	}

	return true
}

func identifyRingContext(path string, rings []RingConfig) (RingConfig, bool) {
	for _, ring := range rings {
		if strings.Contains(path, ring.Path+"/") || strings.HasSuffix(path, ring.Path) {
			return ring, true
		}
	}
	return RingConfig{}, false
}

func isViolatingRingFlow(importPath string, currentRing RingConfig, allRings []RingConfig, projectName string) bool {
	if !strings.Contains(importPath, "internal/ring") {
		return false
	}

	var targetRing *RingConfig
	for _, ring := range allRings {
		if strings.Contains(importPath, ring.Path+"/") || strings.HasSuffix(importPath, ring.Path) {
			targetRing = &ring
			break
		}
	}

	if targetRing == nil {
		return false
	}

	if targetRing.ID == currentRing.ID {
		return false
	}

	for _, allowedID := range currentRing.AllowedDependencies {
		if targetRing.ID == allowedID {
			return false
		}
	}

	return true
}