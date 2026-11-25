package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"honnef.co/go/tools/staticcheck"
)

func TestAnalyzerRegistry(t *testing.T) {
	checks := analyzers()
	if err := analysis.Validate(checks); err != nil {
		t.Fatal(err)
	}
	registered := make(map[string]*analysis.Analyzer)
	for _, check := range checks {
		if _, duplicate := registered[check.Name]; duplicate {
			t.Errorf("duplicate analyzer %s", check.Name)
		}
		registered[check.Name] = check
	}
	for _, check := range staticcheck.Analyzers {
		if registered[check.Analyzer.Name] != check.Analyzer {
			t.Errorf("Staticcheck analyzer %s is missing", check.Analyzer.Name)
		}
	}
	for _, name := range []string{"printf", "copylocks", "lostcancel", "S1000", "ST1019", "QF1004", "ineffassign", "bodyclose", "exitcheck"} {
		if registered[name] == nil {
			t.Errorf("required analyzer %s is missing", name)
		}
	}
}

func TestAnalyzerDocumentation(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "doc.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if file.Doc == nil {
		t.Fatal("staticlint has no package GoDoc")
	}
	doc := file.Doc.Text()
	for _, check := range analyzers() {
		if !strings.Contains(doc, check.Name+": ") {
			t.Errorf("GoDoc has no description for %s", check.Name)
		}
	}
}

func TestRegisteredChecksReportDiagnostics(t *testing.T) {
	for _, check := range analyzers() {
		if check.Name == "SA1000" || check.Name == "ineffassign" || check.Name == "bodyclose" {
			t.Run(check.Name, func(t *testing.T) {
				analysistest.Run(t, analysistest.TestData(), check, strings.ToLower(check.Name))
			})
		}
	}
}
