package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdtoc/internal/version"
)

func TestRunHelpDoesNotRequireInputFile(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := runWithBuildIdentity([]string{"--help"}, &stdout, &stderr, version.BuildIdentity{})

	if got != statusOK {
		t.Fatalf("runWithBuildIdentity() status = %d, want %d", got, statusOK)
	}
	if !strings.Contains(stdout.String(), "mdtoc <markdown-file>") {
		t.Fatalf("stdout = %q, want usage", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunVersionUsesBuildIdentity(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := runWithBuildIdentity(
		[]string{"--version"},
		&stdout,
		&stderr,
		version.BuildIdentity{SourceRef: "test", SourceRevision: "abc123"},
	)

	if got != statusOK {
		t.Fatalf("runWithBuildIdentity() status = %d, want %d", got, statusOK)
	}
	if !strings.Contains(stdout.String(), "mdtoc source ref=test revision=abc123") {
		t.Fatalf("stdout = %q, want supplied build identity", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunReadsMarkdownFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "input.md")
	if err := os.WriteFile(path, []byte("# A\n## B\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := runWithBuildIdentity([]string{path}, &stdout, &stderr, version.BuildIdentity{})

	if got != statusOK {
		t.Fatalf("runWithBuildIdentity() status = %d, want %d; stderr = %q", got, statusOK, stderr.String())
	}
	want := "1-2    '# ^\"A\"$'\n2-2    '# ^\"A\"$ | # ^\"B\"$'\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsUnsupportedShape(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := runWithBuildIdentity(nil, &stdout, &stderr, version.BuildIdentity{})

	if got != statusUsage {
		t.Fatalf("runWithBuildIdentity() status = %d, want %d", got, statusUsage)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage: mdtoc <markdown-file>") {
		t.Fatalf("stderr = %q, want usage diagnostic", stderr.String())
	}
}
