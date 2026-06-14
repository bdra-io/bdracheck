package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// AnalyzeCodebase inspects harvested source streams, parsing layout rules using AST nodes
func AnalyzeCodebase(files []string, config Config) int {
	violations := 0
	fset := token.NewFileSet() // Tracks position coordinates across the code workspace

	for _, fileTarget := range files {
		// Convert text code streams directly into an AST node graph
		node, err := parser.ParseFile(fset, fileTarget, nil, parser.ImportsOnly)
		if err != nil {
			continue // Skip un-compilable files safely during validation sweeps
		}

		// Normalize paths to ensure cross-platform compatibility (Windows vs Unix)
		normalizedPath := filepath.ToSlash(fileTarget)

		// Loop through every import spec declared in the current file header
		for _, imp := range node.Imports {
			// Unquote the raw literal value string (e.g., "\"net/http\"" -> "net/http")
			importPath := strings.Trim(imp.Path.Value, "\"")

			// Check Rule 1: Pure Layer Zero-Dependency Isolation
			if isPureLayerPath(normalizedPath) {
				if isViolatingPureInvariants(importPath, config.ProjectName) {
					fmt.Printf("❌ BOUNDARY VIOLATION [%s]: Pure layer is forbidden from importing external or I/O utilities: '%s'\n", fileTarget, importPath)
					violations++
				}
			}

			// Check Rule 2: Protected Layer Banned I/O Packages
			if isProtectedLayerPath(normalizedPath) {
				for _, bannedPkg := range config.Rules.DisallowIOInProtected.ForbiddenPackages {
					if importPath == bannedPkg || strings.HasPrefix(importPath, bannedPkg+"/") {
						fmt.Printf("❌ LAYER INFRASTRUCTURE LEAK [%s]: Protected layer cannot import framework I/O: '%s'\n", fileTarget, importPath)
						violations++
					}
				}
			}

			// Check Rule 3: Concentric Inward Ring Flow Invariant
			if currentRing, ok := identifyRingContext(normalizedPath, config.Rules.EnforceInwardDependencyFlow.Rings); ok {
				if isViolatingRingFlow(importPath, currentRing, config.Rules.EnforceInwardDependencyFlow.Rings, config.ProjectName) {
					fmt.Printf("❌ RING DEPENDENCY BYPASS [%s]: Inward flow violation! Ring '%s' cannot access this cross-ring node: '%s'\n", fileTarget, currentRing.ID, importPath)
					violations++
				}
			}
		}
	}

	return violations
}

func isPureLayerPath(path string) bool {
	return strings.Contains(path, "/pure/")
}

func isProtectedLayerPath(path string) bool {
	return strings.Contains(path, "/protected/")
}

// Check if a pure file imports standard library I/O or unauthorized cross-domain libraries
func isViolatingPureInvariants(importPath string, projectName string) bool {
	// Standard library utility packages allowed natively within the Pure Layer
	allowedStdLib := map[string]bool{
		"errors":  true,
		"math":    true,
		"strings": true,
		"time":    true,
		"fmt":     true,
		"sort":    true,
		"json":    true, // Serialization models are pure mutations
	}

	if allowedStdLib[importPath] || strings.HasPrefix(importPath, "encoding/") {
		return false
	}

	// Permitted to import matching sub-packages of its EXACT horizontal domain ring path
	if strings.HasPrefix(importPath, projectName+"/internal/ring") && strings.Contains(importPath, "/pure") {
		return false
	}

	return true
}

func identifyRingContext(path string, rings []RingConfig) (RingConfig, bool) {
	for _, ring := range rings {
		if strings.Contains(path, ring.Path+"/") {
			return ring, true
		}
	}
	return RingConfig{}, false
}

// Verifies cross-ring dependency alignments matching the JSON configuration file rules
func isViolatingRingFlow(importPath string, currentRing RingConfig, allRings []RingConfig, projectName string) bool {
	// Skip validation loops if the target import does not belong to internal domains
	if !strings.HasPrefix(importPath, projectName+"/internal/") {
		return false
	}

	// Identify which ring the imported package belongs to
	var targetRing *RingConfig
	for _, ring := range allRings {
		if strings.Contains(importPath, ring.Path+"/") {
			targetRing = &ring
			break
		}
	}

	if targetRing == nil {
		return false
	}

	// Self-imports inside the exact same ring are always legal
	if targetRing.ID == currentRing.ID {
		return false
	}

	// Scan through pre-authorized allowed dependencies
	for _, allowedID := range currentRing.AllowedDependencies {
		if targetRing.ID == allowedID {
			return false // Explicit clearance rule hit
		}
	}

	return true // If not explicitly cleared, it's a boundary bypass violation
}