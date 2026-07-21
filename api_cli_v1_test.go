package simpl

import (
	"encoding/json"
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

	limitedCmd := exec.Command("go", "run", "./cmd/simpl", "run", srcPath, "--max-output-bytes", "5", "--json")
	limitedCmd.Dir = "."
	limitedOut, err := limitedCmd.Output()
	if err == nil {
		t.Fatal("expected output-limited command to return a non-zero status")
	}
	var limitedResult RunResult
	if err := json.Unmarshal(limitedOut, &limitedResult); err != nil {
		t.Fatalf("decode limited result: %v\nOutput:\n%s", err, string(limitedOut))
	}
	if limitedResult.Stdout != "abcd|" {
		t.Fatalf("unexpected limited stdout: %q", limitedResult.Stdout)
	}
	if len(limitedResult.Diagnostics) != 1 || limitedResult.Diagnostics[0].Code != "LIMIT_OUTPUT_EXCEEDED" {
		t.Fatalf("unexpected limited diagnostics: %+v", limitedResult.Diagnostics)
	}

	invalidLimitCmd := exec.Command("go", "run", "./cmd/simpl", "run", srcPath, "--max-output-bytes", "-1", "--json")
	invalidLimitCmd.Dir = "."
	invalidLimitOut, err := invalidLimitCmd.Output()
	if err == nil {
		t.Fatal("expected invalid output limit command to return a non-zero status")
	}
	var invalidLimitResult RunResult
	if err := json.Unmarshal(invalidLimitOut, &invalidLimitResult); err != nil {
		t.Fatalf("decode invalid limit result: %v\nOutput:\n%s", err, string(invalidLimitOut))
	}
	if len(invalidLimitResult.Diagnostics) != 1 || invalidLimitResult.Diagnostics[0].Code != "LIMIT_INVALID_OUTPUT_LIMIT" {
		t.Fatalf("unexpected invalid limit diagnostics: %+v", invalidLimitResult.Diagnostics)
	}

	valueLimitedCmd := exec.Command("go", "run", "./cmd/simpl", "run", srcPath, "--max-value-allocation-bytes", "1", "--json")
	valueLimitedCmd.Dir = "."
	valueLimitedOut, err := valueLimitedCmd.Output()
	if err == nil {
		t.Fatal("expected value-allocation-limited command to return a non-zero status")
	}
	var valueLimitedResult RunResult
	if err := json.Unmarshal(valueLimitedOut, &valueLimitedResult); err != nil {
		t.Fatalf("decode value-allocation-limited result: %v\nOutput:\n%s", err, string(valueLimitedOut))
	}
	if len(valueLimitedResult.Diagnostics) != 1 || valueLimitedResult.Diagnostics[0].Code != "LIMIT_VALUE_ALLOCATION_EXCEEDED" {
		t.Fatalf("unexpected value allocation diagnostics: %+v", valueLimitedResult.Diagnostics)
	}
}
