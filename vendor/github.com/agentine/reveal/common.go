package reveal

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

// valueLen returns a printable length/cap string for array-like types.
func valueLen(v reflect.Value) (int, int) {
	return v.Len(), v.Cap()
}

// sortMapKeys sorts map keys for deterministic output.
// Keys must all be the same type. Supports string, int, uint, float, and bool keys.
func sortMapKeys(keys []reflect.Value) {
	if len(keys) == 0 {
		return
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return mapKeyLess(keys[i], keys[j])
	})
}

// mapKeyLess compares two map keys for sorting.
func mapKeyLess(a, b reflect.Value) bool {
	if a.Kind() != b.Kind() {
		return false
	}
	switch a.Kind() {
	case reflect.String:
		return a.String() < b.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() < b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return a.Uint() < b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() < b.Float()
	case reflect.Bool:
		return !a.Bool() && b.Bool()
	default:
		return false
	}
}

// typeString returns a human-friendly type name for a reflect.Type.
func typeString(t reflect.Type) string {
	return t.String()
}

// hexPointer returns a hex representation of a pointer address.
func hexPointer(v reflect.Value) string {
	var ptr uintptr
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Func, reflect.Chan, reflect.Slice, reflect.UnsafePointer:
		ptr = v.Pointer()
	default:
		return ""
	}
	return "0x" + strconv.FormatUint(uint64(ptr), 16)
}

// stringerRecursion tracks active Stringer/Error invocations to detect
// re-entrant calls (e.g., a Stringer that calls Sdump on itself).
// For addressable values, we key by pointer address. For non-addressable
// values, we key by reflect.Type to prevent unbounded type-based recursion.
var (
	stringerMu          sync.Mutex
	stringerActive      = make(map[uintptr]bool)       // addressable values: keyed by pointer
	stringerTypeDepth   = make(map[reflect.Type]int)    // non-addressable values: keyed by type
)

const maxStringerTypeDepth = 1

// handleMethods attempts to invoke error/Stringer interfaces on the value
// and returns the result string and true if a method was called.
// It supports RecoverPanics (catches panics in Stringer/Error) and
// detects Stringer recursion (a Stringer that calls back into reveal).
func handleMethods(cs *ConfigState, v reflect.Value) (string, bool) {
	if cs.DisableMethods {
		return "", false
	}

	if !v.IsValid() {
		return "", false
	}

	// Check for Stringer recursion.
	if v.CanAddr() {
		// Addressable values: use pointer-based tracking.
		ptr := v.UnsafeAddr()
		stringerMu.Lock()
		if stringerActive[ptr] {
			stringerMu.Unlock()
			return "", false
		}
		stringerActive[ptr] = true
		stringerMu.Unlock()
		defer func() {
			stringerMu.Lock()
			delete(stringerActive, ptr)
			stringerMu.Unlock()
		}()
	} else {
		// Non-addressable values: use type-based depth tracking to prevent
		// unbounded recursion when a Stringer calls back into reveal.
		t := v.Type()
		stringerMu.Lock()
		if stringerTypeDepth[t] >= maxStringerTypeDepth {
			stringerMu.Unlock()
			return "", false
		}
		stringerTypeDepth[t]++
		stringerMu.Unlock()
		defer func() {
			stringerMu.Lock()
			stringerTypeDepth[t]--
			if stringerTypeDepth[t] == 0 {
				delete(stringerTypeDepth, t)
			}
			stringerMu.Unlock()
		}()
	}

	// Try to call Error/String methods, with optional panic recovery.
	if result, ok := tryCallMethods(cs, v); ok {
		return result, true
	}

	return "", false
}

// tryCallMethods attempts to invoke error/Stringer on v, recovering from panics if configured.
func tryCallMethods(cs *ConfigState, v reflect.Value) (result string, ok bool) {
	if cs.RecoverPanics {
		defer func() {
			if r := recover(); r != nil {
				result = fmt.Sprintf("<panic in method: %v>", r)
				ok = true
			}
		}()
	}

	// Try pointer receiver methods first.
	if !cs.DisablePointerMethods && v.CanAddr() {
		pv := v.Addr()
		if pv.CanInterface() {
			if e, eok := pv.Interface().(error); eok {
				return e.Error(), true
			}
			if s, sok := pv.Interface().(interface{ String() string }); sok {
				return s.String(), true
			}
		}
	}

	if v.CanInterface() {
		if e, eok := v.Interface().(error); eok {
			return e.Error(), true
		}
		if s, sok := v.Interface().(interface{ String() string }); sok {
			return s.String(), true
		}
	}

	return "", false
}
