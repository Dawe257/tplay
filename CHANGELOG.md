# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-10-29

### Added

- `install.sh` installs the latest GitHub release into `~/.local/bin` on Linux
  and macOS (amd64 and arm64) via `curl | sh`, verifies SHA-256 checksums, and
  warns if `ffmpeg`/`ffprobe` or the install directory in `PATH` is missing.

### Changed

- CI checks the installer with offline integration tests on pushes to `main`
  and on pull requests.

## [0.2.0] - 2026-10-28

### Added

- Prebuilt binaries for Linux and macOS (amd64 and arm64) on GitHub Releases,
  built by GoReleaser when a `v*` tag is pushed; the release notes are taken
  from this changelog.
- `tplay` without an argument plays the current directory instead of printing
  the usage.

## [0.1.1] - 2026-10-28

### Changed

- The module moved to `github.com/dzhenetl/tplay` after the GitHub username
  change; install with `go install github.com/dzhenetl/tplay/cmd/tplay@latest`.

## [0.1.0] - 2026-10-26

### Added

- Playback of any audio format `ffmpeg` can decode; `tplay <file or directory>`
  starts with the given file or the first file of the directory.
- Playlist of every audio file in the same directory, sorted by name with
  numbers in natural order (`2` before `10`).
- Keys: `space` pause, `←`/`→` (`h`/`l`) seek 5 seconds, `n`/`p` next and
  previous track (`p` restarts the track after 3 seconds), `↑`/`↓` (`k`/`j`,
  `g`/`G`) move in the playlist, `enter` play the selection, `+`/`-` volume,
  `q`/`esc` quit.
- The next track starts when one ends; a file that fails to play is skipped in
  the direction you last moved.
- `tplay --version` prints the version embedded at build time.
- Install with `go install github.com/dzhenetl/tplay/cmd/tplay@latest`.
- CI on GitHub Actions runs `make lint` and `make test` on pushes to `main` and
  on pull requests.

[0.3.0]: https://github.com/dzhenetl/tplay/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/dzhenetl/tplay/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/dzhenetl/tplay/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/dzhenetl/tplay/releases/tag/v0.1.0
