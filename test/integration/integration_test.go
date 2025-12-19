package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// Setup paths
	wd, _ := os.Getwd() // test/integration
	repoRoot := filepath.Dir(filepath.Dir(wd)) // ../..
	fixturesDir := filepath.Join(repoRoot, "test", "fixtures", "valid")
	cmdSrc := filepath.Join(repoRoot, "cmd", "morph", "main.go")
	binPath := filepath.Join(repoRoot, "morph_test_bin")

	// 1. Build the compiler binary
	// We build it once to save time and ensure it runs as a standalone executable
	buildCmd := exec.Command("go", "build", "-o", binPath, cmdSrc)
	buildCmd.Dir = repoRoot
	// buildCmd.Stdout = os.Stdout
	// buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build morph compiler: %v", err)
	}
	defer os.Remove(binPath)

	// 2. Find test files
	files, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to read fixtures dir: %v", err)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".fox") {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			path := filepath.Join(fixturesDir, f.Name())

			// Parse expectation
			content, _ := os.ReadFile(path)
			expected := parseExpectedOutput(string(content))
			if expected == "" {
				t.Skip("No '# EXPECT:' comments found, skipping verification")
			}

			// Run binary
			// We set Dir to fixturesDir so relative imports work naturally
			cmd := exec.Command(binPath, "--vm", f.Name())
			cmd.Dir = fixturesDir

			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Execution failed: %v\nStdout: %s\nStderr: %s", err, out.String(), stderr.String())
			}

			output := strings.TrimSpace(out.String())
			if output != expected {
				t.Errorf("Output mismatch for %s.\nExpected:\n%s\nGot:\n%s", f.Name(), expected, output)
			}
		})
	}
}

func parseExpectedOutput(content string) string {
	lines := strings.Split(content, "\n")
	var expected []string
	found := false
	for _, line := range lines {
		if strings.Contains(line, "# EXPECT:") {
			found = true
			parts := strings.SplitN(line, "# EXPECT:", 2)
			expected = append(expected, strings.TrimSpace(parts[1]))
		}
	}
	if !found {
		return ""
	}
	return strings.Join(expected, "\n")
}
