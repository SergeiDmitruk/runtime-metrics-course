// Package main provides a comprehensive static analysis tool that combines multiple
// Go linters into a single executable. It integrates standard Go analyzers,
// selected StaticCheck analyzers, and custom analyzers to perform extensive
// code quality checks.
//
// Analyzers Included:
//
// Standard Go Analyzers:
//   - atomic: Check for common mistakes with sync/atomic
//   - bools: Check for common mistakes with boolean operators
//   - buildtag: Check build tags
//   - copylock: Check for locks erroneously passed by value
//   - errorsas: Report passing non-pointer values to errors.As
//   - fieldalignment: Find structs that would use less memory if their fields were sorted
//   - httpresponse: Check for mistakes using HTTP responses
//   - loopclosure: Check for references to enclosing loop variables
//   - lostcancel: Check for failure to call context cancellation functions
//   - nilfunc: Check for useless comparisons between functions and nil
//   - printf: Check consistency of Printf format strings and arguments
//   - shadow: Check for possible unintended shadowing of variables
//   - shift: Check for shifts that exceed the width of an integer
//   - sortslice: Check for calls to sort.Slice that don't use a slice type
//   - stdmethods: Check signature of methods like String() and Error()
//   - structtag: Check that struct field tags conform to reflect.StructTag.Get
//   - testinggoroutine: Report calls to (*testing.T).Fatal from goroutines
//   - unmarshal: Report passing non-pointer values to unmarshal functions
//   - unreachable: Check for unreachable code
//   - unsafeptr: Check for invalid conversions of uintptr to unsafe.Pointer
//   - unusedresult: Check for unused results of pure functions
//
// StaticCheck Analyzers:
//   - All SA-class analyzers from staticcheck
//   - Selected analyzers from simple (S1000), stylecheck (ST1000), and quickfix (QF1001)
//
// Custom Analyzers:
//   - noosexit: Checks for direct calls to os.Exit in main package
//
// Usage:
//
//	Compile and run the tool against Go packages:
//	  go build -o staticlint main.go
//	  ./staticlint ./...
//
// The tool can be extended by adding additional analyzers to the appropriate
// analyzer lists in the main function.
package main

import (
	"github.com/runtime-metrics-course/cmd/staticlint/noosexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	standardAnalyzers := []*analysis.Analyzer{
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	var staticcheckAnalyzers []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			staticcheckAnalyzers = append(staticcheckAnalyzers, v.Analyzer)
		}
	}

	staticcheckAnalyzers = append(staticcheckAnalyzers,
		findAnalyzerByName(simple.Analyzers, "S1000"),
		findAnalyzerByName(stylecheck.Analyzers, "ST1000"),
		findAnalyzerByName(quickfix.Analyzers, "QF1001"),
	)

	myAnalyzers := []*analysis.Analyzer{
		noosexit.Analyzer,
	}

	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers, standardAnalyzers...)
	analyzers = append(analyzers, staticcheckAnalyzers...)
	analyzers = append(analyzers, myAnalyzers...)

	multichecker.Main(analyzers...)
}

func findAnalyzerByName(analyzers []*lint.Analyzer, name string) *analysis.Analyzer {
	for _, a := range analyzers {
		if a.Analyzer.Name == name {
			return a.Analyzer
		}
	}
	return nil
}
