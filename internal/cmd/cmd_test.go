package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"knot": Execute,
	}))
}

func TestScript(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	scriptDir := filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "script")
	testscript.Run(t, testscript.Params{
		Dir:           scriptDir,
		UpdateScripts: os.Getenv("UPDATE_SCRIPT") != "",
	})
}
