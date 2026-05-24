#!/bin/sh
set -eu

REPO_OWNER=${MDTOC_REPO_OWNER:-liza-mas}
REPO_NAME=${MDTOC_REPO_NAME:-mdtoc}
INSTALL_DIR=${INSTALL_DIR:-"$HOME/.local/bin"}
SOURCE_REPO=${MDTOC_SOURCE_REPO:-"https://github.com/${REPO_OWNER}/${REPO_NAME}.git"}
SOURCE_TMPDIR=${MDTOC_SOURCE_TMPDIR:-"${TMPDIR:-/tmp}"}
RELEASES_FILE=${MDTOC_RELEASES_FILE:-}
if [ -n "${MDTOC_RELEASES_URL:-}" ]; then
	RELEASES_URL=$MDTOC_RELEASES_URL
elif [ -n "${VERSION:-}" ]; then
	RELEASES_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/releases.tsv"
else
	RELEASES_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/releases.tsv"
fi

fail() {
	printf 'mdtoc install: %s\n' "$*" >&2
	exit 1
}

install_branch() {
	source_branch=$1

	command -v go >/dev/null 2>&1 || fail "Go is required for branch builds"
	command -v make >/dev/null 2>&1 || fail "make is required for branch builds"
	command -v git >/dev/null 2>&1 || fail "git is required to fetch BRANCH=$source_branch"

	source_root=$(mktemp -d "${SOURCE_TMPDIR%/}/mdtoc-source.XXXXXX") || fail "could not create temporary source checkout"
	cleanup_source_root() {
		rm -rf "$source_root"
	}
	trap cleanup_source_root EXIT HUP INT TERM

	source_dir=$source_root/src
	if ! git clone --quiet --depth 1 --branch "$source_branch" "$SOURCE_REPO" "$source_dir"; then
		fail "BRANCH=$source_branch is unavailable from $SOURCE_REPO"
	fi

	if ! source_revision=$(git -C "$source_dir" rev-parse --short HEAD 2>/dev/null); then
		fail "could not identify source revision for BRANCH=$source_branch"
	fi

	if ! make -C "$source_dir" install INSTALL_DIR="$INSTALL_DIR" SOURCE_REF="branch:$source_branch" SOURCE_REVISION="$source_revision"; then
		fail "source install failed for BRANCH=$source_branch"
	fi

	installed_path=$INSTALL_DIR/mdtoc
	if [ ! -x "$installed_path" ]; then
		fail "source install did not create executable mdtoc at $installed_path"
	fi

	if ! version_output=$("$installed_path" --version 2>&1); then
		fail "installed mdtoc at $installed_path failed --version for BRANCH=$source_branch"
	fi
	case "$version_output" in
	*"source"*"branch:$source_branch"*"$source_revision"*)
		;;
	*)
		fail "installed mdtoc at $installed_path did not report source provenance for BRANCH=$source_branch"
		;;
	esac

	printf 'Installed mdtoc source branch=%s revision=%s to %s\n' "$source_branch" "$source_revision" "$installed_path"
}

normalize_os() {
	case "$1" in
	Linux | linux)
		printf 'linux\n'
		;;
	Darwin | darwin)
		printf 'darwin\n'
		;;
	*)
		printf '%s\n' "$1"
		;;
	esac
}

normalize_arch() {
	case "$1" in
	x86_64 | amd64)
		printf 'amd64\n'
		;;
	arm64 | aarch64)
		printf 'arm64\n'
		;;
	*)
		printf '%s\n' "$1"
		;;
	esac
}

detect_platform() {
	raw_os=${MDTOC_INSTALL_OS:-}
	raw_arch=${MDTOC_INSTALL_ARCH:-}

	if [ -z "$raw_os" ]; then
		raw_os=$(uname -s)
	fi
	if [ -z "$raw_arch" ]; then
		raw_arch=$(uname -m)
	fi

	platform_os=$(normalize_os "$raw_os")
	platform_arch=$(normalize_arch "$raw_arch")

	case "$platform_os/$platform_arch" in
	linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64)
		return 0
		;;
	*)
		fail "unsupported platform $platform_os/$platform_arch"
		;;
	esac
}

metadata_source() {
	if [ -n "$RELEASES_FILE" ]; then
		[ -r "$RELEASES_FILE" ] || fail "release metadata file is not readable: $RELEASES_FILE"
		printf '%s\n' "$RELEASES_FILE"
		return 0
	fi

	tmp_file=$(mktemp)
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$RELEASES_URL" -o "$tmp_file" || fail "latest release metadata unavailable at $RELEASES_URL"
	else
		fail "curl is required to fetch release metadata unless MDTOC_RELEASES_FILE is set"
	fi
	printf '%s\n' "$tmp_file"
}

select_release() {
	metadata=$1
	requested=${VERSION:-}
	best_version=
	best_published=
	best_artifact=

	while IFS='	' read -r release_version published_at release_os release_arch artifact_path; do
		[ -n "$release_version" ] || continue
		[ "$release_os" = "$platform_os" ] || continue
		[ "$release_arch" = "$platform_arch" ] || continue

		if [ -n "$requested" ]; then
			if [ "$release_version" = "$requested" ]; then
				best_version=$release_version
				best_artifact=$artifact_path
				break
			fi
			continue
		fi

		if [ -z "$best_published" ] || [ "$published_at" \> "$best_published" ]; then
			best_version=$release_version
			best_published=$published_at
			best_artifact=$artifact_path
		fi
	done <"$metadata"

	if [ -z "$best_version" ]; then
		if [ -n "$requested" ]; then
			fail "VERSION=$requested is unavailable for platform $platform_os/$platform_arch"
		fi
		fail "latest release is unavailable for platform $platform_os/$platform_arch"
	fi

	selected_version=$best_version
	selected_artifact=$best_artifact
}

install_artifact() {
	dest=$INSTALL_DIR/mdtoc
	mkdir -p "$INSTALL_DIR" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"

	case "$selected_artifact" in
	http://* | https://*)
		if command -v curl >/dev/null 2>&1; then
			curl -fsSL "$selected_artifact" -o "$dest" || fail "release artifact is unavailable for $selected_version"
		else
			fail "curl is required to fetch release artifacts"
		fi
		;;
	*)
		[ -r "$selected_artifact" ] || fail "release artifact is unavailable for $selected_version: $selected_artifact"
		cp "$selected_artifact" "$dest" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"
		;;
	esac

	chmod 0755 "$dest" || fail "INSTALL_DIR is not usable: $INSTALL_DIR"
	if ! version_output=$("$dest" --version 2>&1); then
		fail "installed mdtoc at $dest failed --version for $selected_version"
	fi
	case "$version_output" in
	*"$selected_version"*)
		;;
	*)
		fail "installed mdtoc at $dest did not identify $selected_version"
		;;
	esac
	printf 'Installed mdtoc %s to %s\n' "$selected_version" "$dest"
}

if [ -n "${BRANCH:-}" ]; then
	install_branch "$BRANCH"
	exit 0
fi

detect_platform
metadata=$(metadata_source)
select_release "$metadata"
install_artifact
