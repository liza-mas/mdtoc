# Release Validation

This document lists the maintainer validation commands for supported `mdtoc`
distribution workflows. A workflow passes when it installs an executable and
that executable succeeds with `--version`.

These checks validate distribution packaging only. They do not publish hosted
releases, verify package managers, or perform signing or notarization.

## Latest Release

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | bash
mdtoc --version
```

## Explicit Release

Replace `<release>` with the released version under validation.

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | VERSION=<release> bash
mdtoc --version
```

## Custom Install Directory

Replace `<directory>` with the requested install directory. Verify the
executable directly from that directory because it may not be on `PATH`.

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | INSTALL_DIR=<directory> bash
<directory>/mdtoc --version
```

## Branch Source Install

Replace `<branch>` with the source branch under validation. This workflow
requires caller-provided Go and make.

```bash
curl -fsSL https://raw.githubusercontent.com/liza-mas/mdtoc/main/install.sh | BRANCH=<branch> bash
mdtoc --version
```

## Local Clone Source Install

This workflow requires caller-provided Go and make.

```bash
git clone https://github.com/liza-mas/mdtoc.git
cd mdtoc
make install
mdtoc --version
```

For a custom local-clone install directory:

```bash
make install INSTALL_DIR=<directory>
<directory>/mdtoc --version
```
