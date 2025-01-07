# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to [Semantic
Versioning].

## v0.3.0 - 2025-03-06

**Breaking changes!**

Big rewrite.

### Added

-   Helpers for checking the health of some standard library things

-   Tracing, using [OpenTelemetry]

### Changed

-   Restructured packages

    The entire API is now in the top-level package.

-   More use of `context.Context`

    E.g. in health check calls.

### Fixed

-   The HTTP server now responds with 500 if the overall status is failure

## v0.2.5 - 2024-08-01

## v0.2.4 - 2023-10-30

## v0.2.3 - 2023-10-18

## v0.2.2 - 2023-05-22

## v0.2.1 - 2022-05-31

## v0.2.0 - 2022-05-18

## v0.1.0 - 2021-12-06

## v0.0.1 - 2019-11-28

Initial version.

[Keep a Changelog]: https://keepachangelog.com
[OpenTelemetry]: https://opentelemetry.io
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
