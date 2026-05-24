package main

import (
	"fmt"
	"io"
	"os"

	"mdtoc/internal/mdtoc"
	"mdtoc/internal/version"
)

type status int

const (
	statusOK    status = 0
	statusUsage status = 2
	statusInput status = 3
)

const helpText = `Description:
  Print a flat Markdown section table with line ranges and mdq selectors.

Usage:
  mdtoc --help
  mdtoc --version
  mdtoc <markdown-file> [<markdown-file>...]

Output:
  FILE:START-END    'MDQ_SELECTOR'

Exit codes:
  0  success
  2  usage error
  3  input read error
`

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

func run(args []string, stdout io.Writer, stderr io.Writer) status {
	return runWithBuildIdentity(args, stdout, stderr, version.Current())
}

func runWithBuildIdentity(args []string, stdout io.Writer, stderr io.Writer, buildIdentity version.BuildIdentity) status {
	if len(args) == 1 && args[0] == "--help" {
		if _, err := fmt.Fprint(stdout, helpText); err != nil {
			return writeDiagnostic(stderr, statusUsage, err.Error())
		}
		return statusOK
	}

	if len(args) == 1 && args[0] == "--version" {
		if _, err := fmt.Fprintln(stdout, version.Format(buildIdentity)); err != nil {
			return writeDiagnostic(stderr, statusUsage, err.Error())
		}
		return statusOK
	}

	if len(args) == 0 {
		return writeDiagnostic(stderr, statusUsage, "usage: mdtoc <markdown-file> [<markdown-file>...]")
	}

	for _, path := range args {
		file, err := os.Open(path)
		if err != nil {
			return writeDiagnostic(stderr, statusInput, err.Error())
		}

		sections, err := mdtoc.Build(file)
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			return writeDiagnostic(stderr, statusInput, err.Error())
		}

		if _, err := fmt.Fprint(stdout, mdtoc.FormatFile(path, sections)); err != nil {
			return writeDiagnostic(stderr, statusUsage, err.Error())
		}
	}

	return statusOK
}

func writeDiagnostic(stderr io.Writer, code status, message string) status {
	fmt.Fprintf(stderr, "mdtoc: %s\n", message)
	return code
}
