# Changelog

## Unreleased

No unreleased changes

## v0.5.0 (Released 2026-01-19)

* Updated module name to `go.innotegrity.dev/mod/xerrors` for consistency with modules and apps

## v0.4.0 (Released 2025-11-05)

* Updated required `go` version to `1.23.1`

## v0.3.4 (Released 2025-10-29)

* Fixed MarshalJSON so that it includes the wrapped error as well

## v0.3.3 (Released 2025-10-07)

* Fixed nil pointer bug with wrapped errors

## v0.3.2 (Released 2025-10-06)

* Fixed bug when marshalling wrapped error to JSON

## v0.3.1 (Released 2025-10-06)

* Added `CallerInfo` type and `DefaultCallerInfo` and `GetCallerInfo` functions
* Enhanced `MarshalJSON` to include wrapped errors

## v0.2.0 (Released 2025-10-05)

* Added `Wrap` and `Wrapf` functions

## v0.1.0 (Released 2025-10-05)

* Initial release of the module
