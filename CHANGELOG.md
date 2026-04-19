# Changelog

## Unreleased

No unreleased changes

## [v0.13.0] - 2026-04-19

### Added

- Full module documentation is now available along with example usage
- Full unit tests for 100% code coverage have been created
- Options for capturing caller information are now passed via `context.Context` objects or directly when an `Error` is constructed
- `NewAs`, `NewfAs`, `WrapAs`, `WrapfAs` are all new functions designed to make it simpler to create concrete types in 3rd party libraries and applications
- `XError` is now a new, concrete type available from the module
- `NewXError` provides a convenient way to create a new `XError` to use with 3rd party error types

### Changed

- The format of this Changelog has been updated to comply with [keep a changelog v1.1.0](https://keepachangelog.com/en/1.1.0)

### Removed

- Caller capture options are no longer defined globally; instead they are passed as options in context or when the error itself is created

## [v0.7.0] - 2026-04-15

### Added

- Initial unit tests for code coverage

### Security

- Upgraded to require Go 1.25.9 or later for security purposes

## [v0.6.0] - 2026-02-13

### Changed

- Updated module name to `go.innotegrity.dev/mod/xerrors` for consistency with modules and apps

## [v0.4.0] - 2025-11-05

### Security

- Upgraded to require Go 1.23.1 or later for security purposes

## [v0.3.4] - 2025-10-29

### Fixed

- Fixed `MarshalJSON` so that it includes the wrapped error properly now as well

## [v0.3.3] - 2025-10-07

### Fixed

- Fixed nil pointer bug with wrapped errors

## [v0.3.2] - 2025-10-06

### Fixed

- Fixed bug when marshalling wrapped error to JSON

## [v0.3.1] - 2025-10-06

### Added

- Added `CallerInfo` type and `DefaultCallerInfo` and `GetCallerInfo` functions

### Changed

- Enhanced `MarshalJSON` to include wrapped errors

## [v0.2.0] - 2025-10-05

### Added

- Added `Wrap` and `Wrapf` functions

## [v0.1.0] - 2025-10-05

### Added

- Initial release of the module
