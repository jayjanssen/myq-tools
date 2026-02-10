package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

func TestScripts(t *testing.T) {
	// Build the binary for testing
	binary := filepath.Join(t.TempDir(), "myq_status")
	cmd := exec.Command("go", "build", "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	engine := &script.Engine{
		Cmds:  scripttest.DefaultCmds(),
		Conds: scripttest.DefaultConds(),
	}

	// Add the binary directory to PATH
	env := append(os.Environ(), "PATH="+filepath.Dir(binary)+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Run all script tests in testdata/*/*.* (both .txt and .txtar files)
	txtPattern := filepath.Join("testdata", "*", "*")
	txtFiles, err := filepath.Glob(txtPattern)
	if err != nil {
		t.Fatalf("failed to glob test files: %v", err)
	}

	if len(txtFiles) > 0 {
		scripttest.Test(t, t.Context(), engine, env, txtPattern)
	}

}
