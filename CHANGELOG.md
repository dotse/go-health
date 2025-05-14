# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to [Semantic
Versioning].

## 1.0.1 - 2025-05-14

### Changed

-   Upgraded dependencies

### Fixed

-   Use IPv4 as the default server address in the client

    I.e. don’t rely on `localhost` since it might resolve to IPv6 and the server
    only listens on IPv4.

## 1.0.0 - 2025-04-23

No changes; releasing as v1.0.0 and thus stabilising the API.

## 0.4.0 - 2025-04-02

### Added

-   `DeregisterAll()`: Remove all previously registered health checkers

## 0.3.0 - 2025-03-06

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

## 0.2.5 - 2024-08-01

## 0.2.4 - 2023-10-30

## 0.2.3 - 2023-10-18

## 0.2.2 - 2023-05-22

## 0.2.1 - 2022-05-31

## 0.2.0 - 2022-05-18

## 0.1.0 - 2021-12-06

## 0.0.1 - 2019-11-28

Initial version.

[Keep a Changelog]: https://keepachangelog.com
[OpenTelemetry]: https://opentelemetry.io
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
