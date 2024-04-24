package mainexcheck_test

import (
	"testing"

	"github.com/dlomanov/mon/cmd/staticlint/checks/mainexcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMainExitCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), mainexcheck.Analyzer, "./...")
}
