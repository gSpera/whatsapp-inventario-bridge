//go:build !safe && !disableunsafe

package reveal

import (
	"reflect"
	"unsafe"
)

// reflectValue is the internal layout of reflect.Value.
// This must match the runtime struct layout.
type reflectValue struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
	flag uintptr
}

// unsafeReflectValue returns a reflect.Value that is usable even for unexported fields.
// It clears the read-only flag (flagRO) using unsafe pointer manipulation —
// the same technique used by go-spew.
func unsafeReflectValue(v reflect.Value) reflect.Value {
	if !v.IsValid() || v.CanInterface() {
		return v
	}

	// The flagRO bits in reflect.Value.flag prevent reading unexported fields.
	// We clear those bits by manipulating the internal struct directly.
	rv := (*reflectValue)(unsafe.Pointer(&v))
	// flagRO is bits 5 and 6 (flagStickyRO | flagEmbedRO) = 0x60.
	// Clear them to make the value readable.
	rv.flag &^= uintptr(0x60)
	return v
}

const unsafeEnabled = true
