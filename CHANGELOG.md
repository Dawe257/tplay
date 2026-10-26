# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Install with `go install github.com/Dawe257/tplay/cmd/tplay@latest`.
- CI on GitHub Actions runs `make lint` and `make test` on pushes to `main` and
  on pull requests.

[Unreleased]: https://github.com/Dawe257/tplay/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Dawe257/tplay/releases/tag/v0.1.0
