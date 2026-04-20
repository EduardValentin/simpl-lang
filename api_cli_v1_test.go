package simpl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFileWorks(t *testing.T) {
	tmp := t.TempDir()
	srcPath := filepath.Join(tmp, "prog.simpl")
	if err := os.WriteFile(srcPath, []byte("write \"ok\""), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	res := RunFile(srcPath, "", RunOptions{})
	if len(res.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", res.Diagnostics)
	}
	if res.Stdout != "ok" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}

func TestCLICheckAndRun(t *testing.T) {
	tmp := t.TempDir()
	srcPath := filepath.Join(tmp, "prog.simpl")
	stdinPath := filepath.Join(tmp, "in.txt")
	source := `var s string = "abc"
var a array[int] = [1,2,3]
push s, "d", "e"
push a, 4, 5
pop s
pop a
write s, "|", size s, "|", a, "|", size a`
	if err := os.WriteFile(srcPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.WriteFile(stdinPath, []byte("unused"), 0o600); err != nil {
		t.Fatalf("write stdin file: %v", err)
	}

	checkCmd := exec.Command("go", "run", "./cmd/simpl", "check", srcPath)
	checkCmd.Dir = "."
	checkOut, err := checkCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(checkOut))
	}
	if !strings.Contains(string(checkOut), "OK") {
		t.Fatalf("unexpected check output: %s", string(checkOut))
	}

	runCmd := exec.Command("go", "run", "./cmd/simpl", "run", srcPath, "--stdin", stdinPath)
	runCmd.Dir = "."
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run command failed: %v\nOutput:\n%s", err, string(runOut))
	}
	if strings.TrimSpace(string(runOut)) != "abcd|4|[1, 2, 3, 4]|4" {
		t.Fatalf("unexpected run output: %q", string(runOut))
	}
}
