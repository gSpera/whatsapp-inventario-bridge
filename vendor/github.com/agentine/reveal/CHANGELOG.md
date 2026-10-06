# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-12

Initial stable release of **reveal** — a zero-dependency, drop-in replacement for [go-spew](https://github.com/davecgh/go-spew) with improved safety, correctness, and performance.

### Added

- **Core dump engine** (`dump.go`) — deep-pretty-prints any Go value with type annotations, pointer tracking, indentation, and configurable depth. Handles structs, maps, slices, arrays, interfaces, pointers, and function values.
- **Format support** (`format.go`) — implements `fmt.Formatter` on all types, enabling `%v`, `%+v`, `%#v`, `%s` verbs in `fmt.Printf` and friends.
- **Sprint family** — `Sdump`, `Sprintf`, `Sprint`, `Sprintln` for formatted string output without printing.
- **Safety features** (`bypass.go` / `bypass_safe.go`) — unsafe memory access for unexported field inspection, with a `-tags safe` build flag that disables unsafe access and uses reflection fallbacks only. Both modes are tested.
- **Panic recovery** — `RecoverPanics` config option (default `true`) wraps Stringer/GoStringer calls to prevent panics from crashing the caller.
- **Circular reference tracking** — `seen` pointer map prevents infinite recursion on self-referential data structures (maps, slices, structs with pointer cycles).
- **Configurable output** — `ConfigState` with `Indent`, `MaxDepth`, `DisableMethods`, `DisablePointerMethods`, `ContinueOnMethod`, `SortKeys`, `RecoverPanics`, `MaxSize` (default 10 MB output limit).
- **Global config** — `Config` global for `Dump`/`Fdump`/`Sdump` and the `NewFormatter` convenience functions.
- **Instance config** — `ConfigState.Dump` / `ConfigState.Fdump` / `ConfigState.Sdump` for per-call configuration.
- **Fuzz tests** — `FuzzDump` covering a broad range of value types.
- **Benchmarks** — comparison against go-spew baseline.
- **Compat tests** (`compat_test.go`) — golden-output comparisons against go-spew for struct, map, and pointer values.
- **`-tags safe` CI pass** — all tests run under both normal and safe builds.

### Fixed

- **Circular reference crash in maps** (#169) — `dumpMap()` and `dumpSliceOrArray()` now track map/slice pointers in `d.seen` before iterating, preventing infinite recursion.
- **Recursive Stringer stack overflow** (#170) — non-addressable values passed to Stringer are no longer promoted to pointer receivers, preventing infinite method dispatch loops.
- **Safe-build test failure** (#171) — `TestDumpUnexportedFields` field renamed `Exported` (was `exported`, which is unexported and thus inaccessible under `-tags safe`).
- Dead code removed: `pointerTracker`, `isPointerType`, `printableKey`, `indentString` helper functions eliminated.
