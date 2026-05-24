package mdtoc

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Section is one ATX Markdown heading with its inclusive source range.
type Section struct {
	StartLine int
	EndLine   int
	Level     int
	Title     string
	Selector  string
	Ambiguous bool
}

type heading struct {
	line     int
	level    int
	title    string
	selector string
}

// Build reads Markdown and returns source-order sections for every ATX heading.
func Build(reader io.Reader) ([]Section, error) {
	lines, headings, err := scan(reader)
	if err != nil {
		return nil, err
	}

	selectorCounts := make(map[string]int, len(headings))
	for _, heading := range headings {
		selectorCounts[heading.selector]++
	}

	sections := make([]Section, 0, len(headings))
	for index, heading := range headings {
		endLine := len(lines)
		for next := index + 1; next < len(headings); next++ {
			if headings[next].level <= heading.level {
				endLine = headings[next].line - 1
				break
			}
		}

		sections = append(sections, Section{
			StartLine: heading.line,
			EndLine:   endLine,
			Level:     heading.level,
			Title:     heading.title,
			Selector:  heading.selector,
			Ambiguous: selectorCounts[heading.selector] > 1,
		})
	}

	return sections, nil
}

// Format renders the stable agent-facing mdtoc output.
func Format(sections []Section) string {
	var builder strings.Builder
	for _, section := range sections {
		selector := shellSingleQuote(section.Selector)
		if section.Ambiguous {
			selector = "AMBIGUOUS " + selector
		}
		fmt.Fprintf(&builder, "%d-%d    %s\n", section.StartLine, section.EndLine, selector)
	}

	return builder.String()
}

func scan(reader io.Reader) ([]string, []heading, error) {
	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0)
	headings := make([]heading, 0)
	stack := make([]heading, 0, 6)

	inFence := false
	fenceMarker := byte(0)
	fenceLength := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		lineNumber := len(lines)

		if marker, length, ok := fenceBoundary(line); ok {
			if !inFence {
				inFence = true
				fenceMarker = marker
				fenceLength = length
				continue
			}
			if marker == fenceMarker && length >= fenceLength {
				inFence = false
				fenceMarker = 0
				fenceLength = 0
			}
			continue
		}
		if inFence {
			continue
		}

		level, title, ok := atxHeading(line)
		if !ok {
			continue
		}

		for len(stack) > 0 && stack[len(stack)-1].level >= level {
			stack = stack[:len(stack)-1]
		}

		current := heading{
			line:  lineNumber,
			level: level,
			title: title,
		}
		current.selector = selector(append(stack, current))
		headings = append(headings, current)
		stack = append(stack, current)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return lines, headings, nil
}

func fenceBoundary(line string) (byte, int, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || len(trimmed) < 3 {
		return 0, 0, false
	}

	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return 0, 0, false
	}

	length := 0
	for length < len(trimmed) && trimmed[length] == marker {
		length++
	}
	if length < 3 {
		return 0, 0, false
	}

	return marker, length, true
}

func atxHeading(line string) (int, string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || trimmed == "" || trimmed[0] != '#' {
		return 0, "", false
	}

	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level > 6 {
		return 0, "", false
	}
	if level < len(trimmed) && trimmed[level] != ' ' && trimmed[level] != '\t' {
		return 0, "", false
	}

	title := atxHeadingTitle(trimmed[level:])
	return level, title, true
}

func atxHeadingTitle(raw string) string {
	withoutTrailingSpace := strings.TrimRight(raw, " \t")
	index := len(withoutTrailingSpace) - 1
	for index >= 0 && withoutTrailingSpace[index] == '#' {
		index--
	}
	if index == len(withoutTrailingSpace)-1 {
		return strings.TrimSpace(raw)
	}
	if index >= 0 && (withoutTrailingSpace[index] == ' ' || withoutTrailingSpace[index] == '\t') {
		return strings.TrimSpace(withoutTrailingSpace[:index])
	}

	return strings.TrimSpace(raw)
}

func selector(path []heading) string {
	parts := make([]string, 0, len(path))
	for _, item := range path {
		parts = append(parts, `# ^"`+mdqDoubleQuotedText(item.title)+`"$`)
	}

	return strings.Join(parts, " | ")
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func mdqDoubleQuotedText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
