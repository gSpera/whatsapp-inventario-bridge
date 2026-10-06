# Reveal — Deep Pretty Printer for Go

## Overview

**Reveal** is a modern, go-spew-compatible deep pretty printer for Go data structures. It replaces `davecgh/go-spew` (18,454 importers, 6.4K stars, last commit October 2017) with a maintained, bug-fixed, and enhanced alternative.

**Package name:** `github.com/agentine/reveal`
**Registry:** pkg.go.dev (Go modules — publish by pushing a tagged release to GitHub)
**License:** MIT
**Go version:** Go 1.21+

## Target

| | go-spew | reveal |
|---|---|---|
| Maintainer | Single (davecgh), inactive since 2017 | @agentine org |
| Last release | v1.1.1 (Feb 2018) | Active |
| Open issues | 45 (unaddressed) | — |
| Open PRs | 23 (ignored) | — |
| Known panics | 4+ unpatched crash bugs | Fixed |
| Memory safety | Memory exhaustion bug (#145) | Bounded recursion + size limits |
| API | spew.Dump, Sdump, Sprint, Sprintf, etc. | Identical + extras |
| Circular refs | Detected | Detected (improved) |
| Performance | Baseline | Optimized (sync.Pool, reduced allocs) |
| Module support | go.mod added late, no proper versioning | Proper v1 module from day one |

## Design Principles

1. **API-compatible with go-spew** — same package-level functions, same ConfigState type, same output format. Users change one import path and everything works.
2. **Fix all known bugs** — panics with custom Stringers, unexported fields, memory exhaustion.
3. **Safety by default** — bounded recursion depth, maximum output size, recover from Stringer panics.
4. **Zero dependencies** — pure Go, only stdlib imports.
5. **Modern Go** — proper module structure, generics where helpful (Go 1.21+), comprehensive tests.

## API Surface (go-spew compatible)

### Package-level functions (use DefaultConfig)

```go
// Dump formats and displays to stdout (like fmt.Println but deep)
func Dump(a ...interface{})

// Sdump returns the formatted string
func Sdump(a ...interface{}) string

// Sprint/Sprintf/Sprintln — deep-aware fmt replacements
func Sprint(a ...interface{}) string
func Sprintf(format string, a ...interface{}) string
func Sprintln(a ...interface{}) string

// Fprint/Fprintf/Fprintln — deep-aware fmt to io.Writer
func Fprint(w io.Writer, a ...interface{}) (int, error)
func Fprintf(w io.Writer, format string, a ...interface{}) (int, error)
func Fprintln(w io.Writer, a ...interface{}) (int, error)

// Fdump — Dump to io.Writer
func Fdump(w io.Writer, a ...interface{})

// NewFormatter — returns a fmt.Formatter for use with fmt verbs
func NewFormatter(v interface{}) fmt.Formatter
```

### ConfigState (go-spew compatible fields)

```go
type ConfigState struct {
    // go-spew compatible fields
    Indent                  string // default: " "
    MaxDepth                int    // 0 = unlimited
    DisableMethods          bool
    DisablePointerMethods   bool
    DisablePointerAddresses bool
    DisableCapacities       bool
    ContinueOnMethod        bool
    SortKeys                bool
    SpewKeys                bool

    // New reveal-only fields
    MaxSize                 int    // max output bytes (0 = unlimited, default: 10MB)
    RecoverPanics           bool   // recover from Stringer/Error panics (default: true)
    OmitNilPointers         bool   // suppress nil pointer fields
    OmitUnexported          bool   // suppress unexported fields
    HexIntegers             bool   // format integers as hex
}
```

### Compatibility layer

```go
// For trivial migration: import reveal "github.com/agentine/reveal"
// Then use reveal.Dump(), reveal.Sdump(), etc. — same as spew.Dump(), spew.Sdump()
```

## Architecture

```
reveal/
├── reveal.go          # Package-level convenience functions (Dump, Sdump, Sprint, etc.)
├── config.go          # ConfigState type and default config
├── dump.go            # Core dump formatting (multi-line, recursive traversal)
├── format.go          # fmt.Formatter implementation (Sprint/Sprintf integration)
├── common.go          # Shared types: value walker, pointer tracker, cycle detection
├── bypass.go          # Unsafe reflection bypass for unexported fields (go-spew compat)
├── bypass_safe.go     # Safe fallback (build tag: safe/disableunsafe)
├── doc.go             # Package documentation
├── reveal_test.go     # Core tests
├── dump_test.go       # Dump-specific tests
├── format_test.go     # Formatter tests
├── compat_test.go     # go-spew output compatibility tests (golden file comparison)
├── fuzz_test.go       # Fuzz testing for panic safety
├── benchmark_test.go  # Performance benchmarks vs go-spew
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### Core Components

1. **Value Walker** (`common.go`): Recursive reflection-based traversal of Go values. Tracks visited pointers for circular reference detection. Respects MaxDepth and MaxSize limits.

2. **Dump Formatter** (`dump.go`): Multi-line "Dump" style output that shows type info, pointer addresses, and nested structure. Handles all Go types: basic types, arrays, slices, maps, structs, channels, functions, interfaces, pointers, unsafe pointers.

3. **fmt.Formatter** (`format.go`): Implements `fmt.Formatter` interface so reveal values can be used with `%v`, `%+v`, `%#v`, and custom `%d` (dump format) verb — exactly matching go-spew's behavior.

4. **Unsafe Bypass** (`bypass.go`): Uses `unsafe` package to read unexported struct fields (same technique as go-spew). Provides `bypass_safe.go` alternative when build tag `safe` or `disableunsafe` is set.

5. **Config** (`config.go`): `ConfigState` with all go-spew fields plus new safety fields. Package-level `Config` variable as default. Each ConfigState has its own set of Dump/Sdump/Sprint/etc methods.

## Bug Fixes (vs go-spew)

1. **Memory exhaustion (#145)**: Add `MaxSize` config (default 10MB) — stop dumping when output exceeds limit.
2. **Panic with custom Stringer on maps (#141, #115)**: Safely invoke Stringer/Error methods with recover; on panic, fall back to raw dump.
3. **Panic with wrapped custom errors (#144)**: Same recover-based approach for Error() method calls.
4. **Panic with unsorted private fields (#108)**: Fix comparison logic for unexported field sorting.
5. **Recursive Stringer loop**: Detect when a Stringer calls back into spew/reveal and break the cycle.

## Implementation Phases

### Phase 1: Core Engine & Dump

- Set up Go module (`go.mod`), project structure, CI
- Implement value walker with circular reference detection
- Implement `ConfigState` with all go-spew-compatible fields
- Implement `Dump`, `Sdump`, `Fdump` — multi-line deep dump format
- Implement unsafe bypass for unexported fields (+ safe build tag)
- Handle all Go types: bool, int/uint variants, float, complex, string, array, slice, map, struct, chan, func, interface, pointer, unsafe pointer
- Add `MaxDepth` support
- Tests for all types, circular references, unexported fields

### Phase 2: fmt.Formatter & Sprint Family

- Implement `fmt.Formatter` interface (`NewFormatter`)
- Implement `Sprint`, `Sprintf`, `Sprintln`
- Implement `Fprint`, `Fprintf`, `Fprintln`
- Support format verbs: `%v` (inline), `%+v` (with fields), `%#v` (Go syntax), `%d` (dump)
- Handle width, precision, and flag characters
- Tests for all format verbs, edge cases

### Phase 3: Safety & Bug Fixes

- Add `MaxSize` output limiting
- Add `RecoverPanics` — safely call Stringer/Error with deferred recover
- Fix map key sorting with unexported fields
- Add Stringer recursion detection
- Fuzz tests for panic safety (every Go type, nil values, circular refs, custom Stringers)
- Verify no panics on adversarial inputs

### Phase 4: Polish & Ship

- Add new reveal-only config options: `OmitNilPointers`, `OmitUnexported`, `HexIntegers`
- go-spew output compatibility test suite (golden files comparing reveal vs spew output)
- Performance benchmarks vs go-spew
- Optimize hot paths: sync.Pool for buffers, reduce reflect allocations
- README with migration guide (single import path change)
- CI/CD: GitHub Actions for test, lint, fuzz, benchmark
- Tag v1.0.0 and publish to pkg.go.dev

## Migration Path

For existing go-spew users, migration is a single import change:

```go
// Before
import "github.com/davecgh/go-spew/spew"
spew.Dump(myStruct)

// After
import "github.com/agentine/reveal"
reveal.Dump(myStruct)
```

Output format is identical by default. New safety features (MaxSize, RecoverPanics) are enabled by default but can be disabled for exact go-spew behavior.

## Success Criteria

- 100% API compatibility with go-spew v1.1.1
- All known go-spew bugs (#108, #115, #141, #144, #145) fixed
- Zero panics on fuzz testing
- Performance equal to or better than go-spew
- Comprehensive test suite (>90% coverage)
- Published to pkg.go.dev with proper module versioning
