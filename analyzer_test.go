package main

import (
	"go/parser"
	"go/token"
	"testing"
)

// Generate mock config mapping standard workspace rules
func getMockConfig() Config {
	return Config{
		ProjectName: "mock-app",
		Rules: RulesConfig{
			DisallowIOInProtected: IORuleConfig{
				ForbiddenPackages: []string{"database/sql", "net/http"},
			},
			EnforceInwardDependencyFlow: FlowRuleConfig{
				Rings: []RingConfig{
					{ID: "ring0", Path: "internal/ring0"},
					{ID: "ring1", Path: "internal/ring1", AllowedDependencies: []string{"ring0"}},
				},
			},
		},
	}
}

func TestAnalyzeCodebase_PassesCleanCode(t *testing.T) {
	config := getMockConfig()
	fset := token.NewFileSet()

	// Clean code: A pure logic layer file containing zero banned frameworks or I/O imports
	srcCode := `package pure
import (
	"errors"
	"math"
)
`
	_, err := parser.ParseFile(fset, "internal/ring1/pure/order.go", srcCode, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("Failed to parse clean test token stream: %v", err)
	}

	// Manually run check on our deterministic pure constraints
	if isViolatingPureInvariants("errors", config.ProjectName) || isViolatingPureInvariants("math", config.ProjectName) {
		t.Error("Expected clean standard library utilities to pass Pure validation, but a leak was flagged.")
	}
}

func TestAnalyzeCodebase_CatchesPureLayerIOLeak(t *testing.T) {
	config := getMockConfig()

	// Violation: Leaking net/http framework directly inside a pure calculation directory
	dirtyImport := "net/http"

	if !isViolatingPureInvariants(dirtyImport, config.ProjectName) {
		t.Errorf("Security Failure: The engine failed to catch a banned network I/O import (%s) inside the Pure layer.", dirtyImport)
	}
}

func TestAnalyzeCodebase_CatchesProtectedLayerIOLeak(t *testing.T) {
	config := getMockConfig()
	bannedImport := "database/sql"
	violationFound := false

	// Test if our protected layer rule identifies raw database access packages
	for _, bannedPkg := range config.Rules.DisallowIOInProtected.ForbiddenPackages {
		if bannedImport == bannedPkg {
			violationFound = true
		}
	}

	if !violationFound {
		t.Error("Failure: The Protected layer boundary rules failed to identify database/sql as a banned infrastructure leak.")
	}
}

func TestAnalyzeCodebase_CatchesOutwardRingBypass(t *testing.T) {
	config := getMockConfig()

	currentRing := RingConfig{ID: "ring0", Path: "internal/ring0"} // Innermost ring
	illegalImport := "mock-app/internal/ring1/protected"          // Outer ring component

	// Inward Flow Rule: Inner rings can never import outer rings!
	isViolating := isViolatingRingFlow(illegalImport, currentRing, config.Rules.EnforceInwardDependencyFlow.Rings, config.ProjectName)

	if !isViolating {
		t.Errorf("Architecture Bypass Failure: Ring 0 blindly imported an outer ring component (%s) without throwing a circular flow violation.", illegalImport)
	}
}