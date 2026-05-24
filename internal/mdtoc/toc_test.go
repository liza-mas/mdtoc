package mdtoc

import (
	"strings"
	"testing"
)

func TestBuildEmitsSourceOrderSectionsWithParentRanges(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"# A",
		"intro",
		"## B",
		"body",
		"# C",
		"tail",
	}, "\n")

	sections, err := Build(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := Format(sections)
	want := strings.Join([]string{
		`1-4    '# ^"A"$'`,
		`3-4    '# ^"A"$ | # ^"B"$'`,
		`5-6    '# ^"C"$'`,
		"",
	}, "\n")
	if got != want {
		t.Fatalf("Format(Build()) = %q, want %q", got, want)
	}
}

func TestFormatFilePrefixesRowsWithPath(t *testing.T) {
	t.Parallel()

	sections, err := Build(strings.NewReader("# A\n## B\n"))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := FormatFile("docs/example.md", sections)
	want := strings.Join([]string{
		`docs/example.md:1-2    '# ^"A"$'`,
		`docs/example.md:2-2    '# ^"A"$ | # ^"B"$'`,
		"",
	}, "\n")
	if got != want {
		t.Fatalf("FormatFile() = %q, want %q", got, want)
	}
}

func TestFormatFilePreservesRawPathPrefix(t *testing.T) {
	t.Parallel()

	sections, err := Build(strings.NewReader("# A\n"))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := FormatFile("space dir/file one:100%.md", sections)
	want := `space dir/file one:100%.md:1-1    '# ^"A"$'` + "\n"
	if got != want {
		t.Fatalf("FormatFile() = %q, want %q", got, want)
	}
}

func TestBuildIgnoresHeadingLikeTextInsideFencedCode(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"# A",
		"```",
		"# Not a heading",
		"```",
		"## B",
	}, "\n")

	sections, err := Build(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := Format(sections)
	want := strings.Join([]string{
		`1-5    '# ^"A"$'`,
		`5-5    '# ^"A"$ | # ^"B"$'`,
		"",
	}, "\n")
	if got != want {
		t.Fatalf("Format(Build()) = %q, want %q", got, want)
	}
}

func TestBuildReturnsNoRowsForNoHeadings(t *testing.T) {
	t.Parallel()

	sections, err := Build(strings.NewReader("plain\ntext\n"))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got := Format(sections); got != "" {
		t.Fatalf("Format(Build()) = %q, want empty output", got)
	}
}

func TestBuildOnlyTrimsClosingATXMarkerAfterWhitespace(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"# C#",
		"# Title ###",
		"# ###",
		"# ##",
		"# #",
	}, "\n")

	sections, err := Build(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := Format(sections)
	want := strings.Join([]string{
		`1-1    '# ^"C#"$'`,
		`2-2    '# ^"Title"$'`,
		`3-3    AMBIGUOUS '# ^""$'`,
		`4-4    AMBIGUOUS '# ^""$'`,
		`5-5    AMBIGUOUS '# ^""$'`,
		"",
	}, "\n")
	if got != want {
		t.Fatalf("Format(Build()) = %q, want %q", got, want)
	}
}

func TestFormatMarksDuplicateSelectorPathsAmbiguous(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"# A",
		"## B",
		"## B",
	}, "\n")

	sections, err := Build(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := Format(sections)
	want := strings.Join([]string{
		`1-3    '# ^"A"$'`,
		`2-2    AMBIGUOUS '# ^"A"$ | # ^"B"$'`,
		`3-3    AMBIGUOUS '# ^"A"$ | # ^"B"$'`,
		"",
	}, "\n")
	if got != want {
		t.Fatalf("Format(Build()) = %q, want %q", got, want)
	}
}

func TestFormatEscapesSelectorTextForMDQAndShell(t *testing.T) {
	t.Parallel()

	input := "# A \"quoted\" user's guide\n"

	sections, err := Build(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	got := Format(sections)
	want := "1-1    '# ^\"A \\\"quoted\\\" user'\\''s guide\"$'\n"
	if got != want {
		t.Fatalf("Format(Build()) = %q, want %q", got, want)
	}
}
