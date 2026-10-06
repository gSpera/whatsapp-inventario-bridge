package reveal

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// bufPool is a sync.Pool for reusing bytes.Buffer instances to reduce allocations.
var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ConfigState holds the configuration options for reveal.
// All go-spew compatible fields are supported, plus additional reveal-only fields.
type ConfigState struct {
	// Indent specifies the string to use for each indentation level. Default: " " (single space).
	Indent string

	// MaxDepth controls the maximum depth to recurse into nested types. 0 means unlimited.
	MaxDepth int

	// DisableMethods disables invocation of error and Stringer interface methods.
	DisableMethods bool

	// DisablePointerMethods disables invocation of error and Stringer interface
	// methods on pointer receivers. This only applies when DisableMethods is false.
	DisablePointerMethods bool

	// DisablePointerAddresses disables printing of pointer addresses.
	DisablePointerAddresses bool

	// DisableCapacities disables printing of capacities for arrays, slices, and maps.
	DisableCapacities bool

	// ContinueOnMethod causes the dump to continue after calling a Stringer/Error method,
	// showing the internal structure as well.
	ContinueOnMethod bool

	// SortKeys causes map keys to be sorted before printing to produce deterministic output.
	SortKeys bool

	// SpewKeys causes map keys to be formatted using the Dump format for the key.
	SpewKeys bool

	// MaxSize is the maximum number of bytes the output may be. 0 means unlimited.
	// Default: 10485760 (10 MB). Prevents memory exhaustion with deeply nested structures.
	MaxSize int

	// RecoverPanics causes reveal to recover from panics in Stringer/Error methods
	// and fall back to raw value dumping. Default: true.
	RecoverPanics bool

	// OmitNilPointers suppresses nil pointer fields in struct output.
	OmitNilPointers bool

	// OmitUnexported suppresses unexported struct fields from output.
	OmitUnexported bool

	// HexIntegers formats integer values in hexadecimal.
	HexIntegers bool
}

// Config is the active default configuration for reveal.
// Modify this to change the default behavior of package-level functions.
var Config = ConfigState{
	Indent:        " ",
	MaxSize:       10 * 1024 * 1024, // 10 MB
	RecoverPanics: true,
}

// Dump displays the passed parameters to standard output with type info, pointer addresses,
// and nested structure using the configuration options in c.
func (c *ConfigState) Dump(a ...interface{}) {
	c.Fdump(os.Stdout, a...)
}

// Sdump returns a string with the passed parameters formatted using Dump style.
func (c *ConfigState) Sdump(a ...interface{}) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	c.Fdump(buf, a...)
	return buf.String()
}

// Fdump formats and writes the passed parameters to the given io.Writer.
func (c *ConfigState) Fdump(w io.Writer, a ...interface{}) {
	for i, arg := range a {
		if i > 0 {
			w.Write([]byte("\n"))
		}
		d := dumpState{
			w:      w,
			cs:     c,
			seen:   make(map[uintptr]bool),
			depth:  0,
			indent: "",
		}
		d.dump(arg)
		w.Write([]byte("\n"))
	}
}
