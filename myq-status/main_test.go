package main

import (
	"context"
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

	// Run all script tests in testdata/script (both .txt and .txtar files)
	pattern := filepath.Join("testdata", "script", "*.txt")
	scripttest.Test(t, context.Background(), engine, env, pattern)

	pattern = filepath.Join("testdata", "script", "*.txtar")
	scripttest.Test(t, context.Background(), engine, env, pattern)
}
