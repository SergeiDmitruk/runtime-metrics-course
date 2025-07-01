package noosexit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExit(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "testdata/src/...")
}
