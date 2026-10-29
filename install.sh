#!/bin/sh
# Install the latest tplay release. Compatible with: curl -fsSL .../install.sh | sh
set -eu

die() {
    printf 'tplay: %s\n' "$*" >&2
    exit 1
}

download() {
    curl --fail --silent --show-error --location --retry 3 \
        --proto '=https' --proto-redir '=https' "$@"
}

main() {
    for tool in curl tar awk mktemp chmod mv mkdir rm uname cat; do
        command -v "$tool" >/dev/null 2>&1 || die "Required command missing: $tool"
    done

    case "$(uname -s)" in
        Linux) os=linux ;;
        Darwin) os=darwin ;;
        *) die "Unsupported operating system: $(uname -s)" ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) arch=amd64 ;;
        aarch64|arm64) arch=arm64 ;;
        *) die "Unsupported architecture: $(uname -m)" ;;
    esac

    if command -v sha256sum >/dev/null 2>&1; then
        checksum_tool=sha256sum
    elif command -v shasum >/dev/null 2>&1; then
        checksum_tool=shasum
    else
        die 'Required command missing: sha256sum or shasum'
    fi
    [ -n "${HOME:-}" ] || die 'HOME is not set'
    install_dir=$HOME/.local/bin
    repo=https://github.com/dzhenetl/tplay

    # Follow GitHub's latest-release redirect instead of parsing JSON or HTML.
    release_url=$(download -o /dev/null -w '%{url_effective}' "$repo/releases/latest") ||
        die 'Could not find the latest release'
    case "$release_url" in
        "$repo/releases/tag/"*) tag=${release_url##*/} ;;
        *) die 'Unexpected latest release URL' ;;
    esac
    case "$tag" in
        ''|*[!a-zA-Z0-9._-]*) die 'Invalid release tag' ;;
    esac
    version=${tag#v}
    archive=tplay_${version}_${os}_${arch}.tar.gz
    base_url=$repo/releases/download/$tag

    work_dir=$(mktemp -d "${TMPDIR:-/tmp}/tplay-install.XXXXXX") || die 'Cannot create temporary directory'
    staged_binary=
    trap 'rm -rf "$work_dir"; if [ -n "$staged_binary" ]; then rm -f "$staged_binary"; fi' 0
    trap 'exit 1' HUP INT TERM

    printf 'Downloading tplay %s for %s/%s...\n' "$version" "$os" "$arch"
    download -o "$work_dir/$archive" "$base_url/$archive" || die 'Archive download failed'
    download -o "$work_dir/checksums.txt" "$base_url/checksums.txt" || die 'Checksum download failed'

    expected=$(awk -v name="$archive" '$2 == name { count++; hash=$1 } END { if (count != 1) exit 1; print hash }' \
        "$work_dir/checksums.txt") || die 'Missing or duplicate checksum for archive'
    case "$expected" in
        ''|*[!0-9a-fA-F]*) die 'Invalid checksum' ;;
    esac
    [ "${#expected}" -eq 64 ] || die 'Invalid checksum length'
    if [ "$checksum_tool" = sha256sum ]; then
        actual=$(sha256sum "$work_dir/$archive") || die 'Could not compute checksum'
    else
        actual=$(shasum -a 256 "$work_dir/$archive") || die 'Could not compute checksum'
    fi
    actual=${actual%% *}
    [ "$actual" = "$expected" ] || die 'Archive checksum mismatch'

    # Extract only the binary, never the rest of the release archive.
    tar -xzf "$work_dir/$archive" -C "$work_dir" tplay || die 'Could not extract binary'
    [ -f "$work_dir/tplay" ] && [ ! -L "$work_dir/tplay" ] || die 'Archive does not contain a regular tplay binary'
    chmod 755 "$work_dir/tplay"
    mkdir -p "$install_dir" || die "Cannot create $install_dir"
    [ ! -d "$install_dir/tplay" ] || die "Destination is a directory: $install_dir/tplay"
    # Stage on the destination filesystem so replacement is atomic.
    staged_binary=$(mktemp "$install_dir/.tplay.XXXXXX") || die 'Cannot stage binary'
    cat "$work_dir/tplay" > "$staged_binary"
    chmod 755 "$staged_binary"
    mv -f "$staged_binary" "$install_dir/tplay" || die 'Could not install binary'
    staged_binary=
    printf 'Installed tplay to %s/tplay\n' "$install_dir"

    case ":${PATH:-}:" in
        *":$install_dir:"*) ;;
        *) printf 'Add %s to your PATH to run tplay.\n' "$install_dir" >&2 ;;
    esac
    for tool in ffmpeg ffprobe; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            printf 'Warning: %s is missing; install ffmpeg to play audio.\n' "$tool" >&2
        fi
    done
}

main "$@"
