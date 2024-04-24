package main

import (
	"github.com/dlomanov/mon/cmd/staticlint/checks/mainexcheck"
	"github.com/jingyugao/rowserrcheck/passes/rowserr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	cs := []*analysis.Analyzer{}
	cs = append(cs, stdchecks()...)
	cs = append(cs, staticchecks()...)
	cs = append(cs, misc()...)
	multichecker.Main(cs...)
}

// stdchecks returns default vet tool analyzers.
func stdchecks() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		// check for missing values after append
		appends.Analyzer,

		// report mismatches between assembly files and Go declarations
		asmdecl.Analyzer,

		// check for useless assignments
		assign.Analyzer,

		// check for common mistakes using the sync/atomic package
		atomic.Analyzer,

		// checks for non-64-bit-aligned arguments to sync/atomic functions
		atomicalign.Analyzer,

		// check for common mistakes involving boolean operators
		bools.Analyzer,

		// check that +build tags are well-formed and correctly located
		buildtag.Analyzer,

		// detect some violations of the cgo pointer passing rules. Not working in Arcadia for now
		cgocall.Analyzer,

		// check for unkeyed composite literals
		composite.Analyzer,

		// check for locks erroneously passed by value
		copylock.Analyzer,

		// check for the use of reflect.DeepEqual with error values
		deepequalerrors.Analyzer,

		// check that the second argument to errors.As is a pointer to a type implementing error
		errorsas.Analyzer,

		// check for mistakes using HTTP responses
		httpresponse.Analyzer,

		// check references to loop variables from within nested functions
		loopclosure.Analyzer,

		// check cancel func returned by context.WithCancel is called
		lostcancel.Analyzer,

		// check for useless comparisons between functions and nil
		nilfunc.Analyzer,

		// inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons
		nilness.Analyzer,

		// check consistency of Printf format strings and arguments
		printf.Analyzer,

		// check for possible unintended shadowing of variables
		shadow.Analyzer,

		// check for shifts that equal or exceed the width of the integer
		shift.Analyzer,

		// check signature of methods of well-known interfaces
		stdmethods.Analyzer,

		// check that struct field tags conform to reflect.StructTag.Get
		structtag.Analyzer,

		// check for common mistaken usages of tests and examples
		tests.Analyzer,

		// report passing non-pointer or non-interface values to unmarshal
		unmarshal.Analyzer,

		// check for unreachable code
		unreachable.Analyzer,

		// check for invalid conversions of uintptr to unsafe.Pointer
		unsafeptr.Analyzer,

		// check for unused results of calls to some functions
		unusedresult.Analyzer,

		// check for unused writes
		unusedwrite.Analyzer,

		// check for string(int) conversions
		stringintconv.Analyzer,

		// check for impossible interface-to-interface type assertions
		ifaceassert.Analyzer,
	}
}

// staticchecks returns staticcheck analyzers.
func staticchecks() []*analysis.Analyzer {
	checks := []*analysis.Analyzer{}

	// simple - checks that are concerned with simplifying code
	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	// staticcheck - checks that are concerned with the correctness of code
	for _, v := range staticcheck.Analyzers {
		switch v.Analyzer.Name {
		// storing non-pointer values in sync.Pool allocates memory
		case "SA6002":
			continue
		// field assignment that will never be observed
		case "SA4005":
			continue
		}

		checks = append(checks, v.Analyzer)
	}

	// stylecheck - contains all checks that are concerned with stylistic issues
	for _, v := range stylecheck.Analyzers {
		switch v.Analyzer.Name {
		// incorrect or missing package comment
		case "ST1000":
			continue
		// the documentation of an exported function should start with the function's name
		case "ST1020":
			continue
		// the documentation of an exported type should start with type's name
		case "ST1021":
			continue
		// the documentation of an exported variable or constant should start with variable's name
		case "ST1022":
			continue
		}
		checks = append(checks, v.Analyzer)
	}

	return checks
}

func misc() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		// checks whether HTTP response body is closed successfully
		bodyclose.Analyzer,
		// rowserrcheck checks whether Rows.Err is checked"
		// database/sql included by default
		rowserr.NewAnalyzer("github.com/jmoiron/sqlx"),

		// check that main function does not exit via os.Exit
		mainexcheck.Analyzer,
	}
}
