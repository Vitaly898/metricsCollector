package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"paniccalls", "termination", "mainpkg", "aliases", "dotimports", "shadowed",
		"mainaliases", "maindot", "resetcli", "testharness.test", "ordinary.test", "generated")
}
