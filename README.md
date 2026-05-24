# mdtoc

`mdtoc` prints a flat table of Markdown sections with inclusive line ranges and
complete [`mdq`](https://github.com/yshavit/mdq) selectors.

```bash
mdtoc path/to/file.md
```

Output shape:

```text
START-END    'MDQ_SELECTOR'
```

Example:

```text
1-12    '# ^"Architecture"$'
5-12    '# ^"Architecture"$ | # ^"Data Flow"$'
```

## Behavior Contract

`mdtoc` emits one row for every ATX heading (`#` through `######`) in source order.
Heading-like text inside fenced code blocks is ignored.

Line ranges are 1-based and inclusive. A section starts on its heading line and
ends on the line before the next heading of the same or higher level, or on the
final line of the file.

Selectors include the full heading path from ancestor to current heading. Each
segment is an anchored [`mdq`](https://github.com/yshavit/mdq) section selector,
and segments are chained with ` | `:

```text
'# ^"Parent"$ | # ^"Child"$ | # ^"Leaf"$'
```

Successful output has no explanatory prose. Every non-empty line starts with a
`START-END` token, followed by whitespace, followed by the selector field. The
range can be converted to `sed -n 'START,ENDp'`; the selector field is intended
to be copied directly into `mdq`.

If a heading path cannot be represented uniquely, the row marks the selector as
ambiguous while keeping the line range usable:

```text
12-18    AMBIGUOUS '# ^"A"$ | # ^"Duplicate"$'
19-25    AMBIGUOUS '# ^"A"$ | # ^"Duplicate"$'
```

Files with no headings produce no rows and exit successfully.

## Scope

The MVP intentionally has no behavior-changing options, JSON mode, tree mode,
interactive mode, document summary, link resolution, include handling, or file
mutation.

Setext headings are not parsed. Selector escaping follows `mdq` double-quoted
text syntax inside a shell single-quoted argument; if `mdq` syntax changes,
`mdtoc` should follow `mdq` rather than invent its own selector language.

## Installation

**Quick install (latest release, macOS/Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | bash
mdtoc --version
```

**Options:**

```bash
# Explicit release
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | VERSION=<release> bash
mdtoc --version

# Build from a branch with caller-provided Go and make
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | BRANCH=<branch> bash
mdtoc --version

# Custom install directory
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | INSTALL_DIR=<directory> bash
<directory>/mdtoc --version
```

**From a local clone:**

```bash
git clone https://github.com/liza-mas/mdtoc.git
cd mdtoc
make install
mdtoc --version
```

Local clone installs require caller-provided Go and make. Use
`INSTALL_DIR=<directory> make install` to install from a local clone into a
custom directory, then verify with `<directory>/mdtoc --version`.

Build and test:

```bash
make test
make build
```

Build a release artifact whose `--version` output identifies the selected
release:

```bash
make release-build RELEASE_IDENTITY=v1.0.0
./build/mdtoc --version
```
