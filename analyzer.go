package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"sync"
)

// Violation represents an isolated architectural boundary breach
type Violation struct {
	Position string
	Layer    string
	Message  string
}

// AnalyzeCodebase orchestrates concurrent AST evaluation loops across all discovered source files
func AnalyzeCodebase(files []string, config Config) int {
	var wg sync.WaitGroup
	violationChan := make(chan Violation, len(files)*2)
	fset := token.NewFileSet()

	// 1. CONCURRENT PIPELINE: Distribute file parsing routines across active system threads
	for _, fileTarget := range files {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			
			// Parse file token specs including adjacent code comment nodes
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

	// 2. LIFECYCLE MANAGEMENT: Close transmission streams once processing routines finish execution
	wg.Wait()
	close(violationChan)

	// 3. DIAGNOSTIC COMPILATION: Harvest and format compiled violations safely
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
		
		// Strip glob qualifiers to isolate pure package roots cleanly
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

	// Dynamic Core Mapping: Permitted to interact with horizontal pure domains safely
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

	// Self dependencies within the same domain are natively legal
	if targetRing.ID == currentRing.ID {
		return false
	}

	// Cross-check pre-authorized allowed structures explicitly loaded from config file
	for _, allowedID := range currentRing.AllowedDependencies {
		if targetRing.ID == allowedID {
			return false
		}
	}

	return true
}