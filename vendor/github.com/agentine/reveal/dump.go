package reveal

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// dumpState tracks the state of a single dump operation.
type dumpState struct {
	w        io.Writer
	cs       *ConfigState
	seen     map[uintptr]bool
	depth    int
	indent   string
	written  int
	exceeded bool
}

// write writes a string to the output. Respects MaxSize.
func (d *dumpState) write(s string) {
	if d.exceeded {
		return
	}
	if d.cs.MaxSize > 0 && d.written+len(s) > d.cs.MaxSize {
		remaining := d.cs.MaxSize - d.written
		if remaining > 0 {
			d.w.Write([]byte(s[:remaining]))
			d.written += remaining
		}
		d.w.Write([]byte("\n... <output truncated, MaxSize reached>"))
		d.exceeded = true
		return
	}
	d.w.Write([]byte(s))
	d.written += len(s)
}

// writeIndent writes the current indentation.
func (d *dumpState) writeIndent() {
	d.write(d.indent)
}

// pushIndent increases the indent level.
func (d *dumpState) pushIndent() {
	d.indent += d.cs.Indent
}

// popIndent decreases the indent level.
func (d *dumpState) popIndent() {
	if len(d.indent) >= len(d.cs.Indent) {
		d.indent = d.indent[:len(d.indent)-len(d.cs.Indent)]
	}
}

// dump formats a single value recursively.
func (d *dumpState) dump(v interface{}) {
	if v == nil {
		d.write("(<nil>) <nil>")
		return
	}
	d.dumpValue(reflect.ValueOf(v))
}

// dumpValue formats a reflect.Value.
func (d *dumpState) dumpValue(v reflect.Value) {
	if !v.IsValid() {
		d.write("<invalid>")
		return
	}

	// Check max depth.
	if d.cs.MaxDepth > 0 && d.depth >= d.cs.MaxDepth {
		d.write("(" + typeString(v.Type()) + ") <max depth reached>")
		return
	}

	kind := v.Kind()

	// Handle interface values by unwrapping.
	if kind == reflect.Interface {
		if v.IsNil() {
			d.write("(" + typeString(v.Type()) + ") <nil>")
			return
		}
		d.dumpValue(v.Elem())
		return
	}

	// Handle pointer types.
	if kind == reflect.Ptr {
		d.dumpPointer(v)
		return
	}

	// Try methods (error/Stringer) before falling through to raw dump.
	if methStr, ok := handleMethods(d.cs, v); ok {
		d.write("(" + typeString(v.Type()) + ") ")
		if d.cs.ContinueOnMethod {
			d.write("(" + methStr + ") ")
			d.dumpRawValue(v)
		} else {
			d.write(methStr)
		}
		return
	}

	d.dumpRawValue(v)
}

// dumpPointer formats a pointer value.
func (d *dumpState) dumpPointer(v reflect.Value) {
	if v.IsNil() {
		d.write("(" + typeString(v.Type()) + ") <nil>")
		return
	}

	// Print the pointer type and address.
	d.write("(*" + typeString(v.Type().Elem()) + ")")
	if !d.cs.DisablePointerAddresses {
		d.write("(" + hexPointer(v) + ")")
	}

	// Check for circular reference.
	ptr := v.Pointer()
	if d.seen[ptr] {
		d.write("(<already shown>)")
		return
	}
	d.seen[ptr] = true
	defer func() { delete(d.seen, ptr) }()

	d.write("(")
	d.dumpValue(v.Elem())
	d.write(")")
}

// dumpRawValue formats the raw value without method invocation or pointer handling.
func (d *dumpState) dumpRawValue(v reflect.Value) {
	t := v.Type()
	kind := v.Kind()

	switch kind {
	case reflect.Bool:
		d.write("(bool) " + strconv.FormatBool(v.Bool()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if d.cs.HexIntegers {
			d.write("(" + typeString(t) + ") 0x" + strconv.FormatInt(v.Int(), 16))
		} else {
			d.write("(" + typeString(t) + ") " + strconv.FormatInt(v.Int(), 10))
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if d.cs.HexIntegers {
			d.write("(" + typeString(t) + ") 0x" + strconv.FormatUint(v.Uint(), 16))
		} else {
			d.write("(" + typeString(t) + ") " + strconv.FormatUint(v.Uint(), 10))
		}

	case reflect.Float32, reflect.Float64:
		d.write("(" + typeString(t) + ") " + strconv.FormatFloat(v.Float(), 'f', -1, int(t.Size())*8))

	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		d.write("(" + typeString(t) + ") " + fmt.Sprintf("(%g+%gi)", real(c), imag(c)))

	case reflect.String:
		d.write("(string) (len=" + strconv.Itoa(v.Len()) + ") " + strconv.Quote(v.String()))

	case reflect.Slice:
		if v.IsNil() {
			d.write("(" + typeString(t) + ") <nil>")
			return
		}
		d.dumpSliceOrArray(v)

	case reflect.Array:
		d.dumpSliceOrArray(v)

	case reflect.Map:
		d.dumpMap(v)

	case reflect.Struct:
		d.dumpStruct(v)

	case reflect.Chan:
		d.dumpChan(v)

	case reflect.Func:
		d.dumpFunc(v)

	case reflect.UnsafePointer:
		d.write("(unsafe.Pointer) " + hexPointer(v))

	default:
		d.write("(" + typeString(t) + ") " + fmt.Sprintf("%v", v.Interface()))
	}
}

// dumpSliceOrArray formats a slice or array value.
func (d *dumpState) dumpSliceOrArray(v reflect.Value) {
	t := v.Type()

	// Track slice pointer for circular reference detection (slices only, not arrays).
	if v.Kind() == reflect.Slice && v.Len() > 0 {
		ptr := v.Pointer()
		if d.seen[ptr] {
			d.write("(" + typeString(t) + ") <already shown>")
			return
		}
		d.seen[ptr] = true
		defer func() { delete(d.seen, ptr) }()
	}

	l, c := valueLen(v)

	header := "(" + typeString(t) + ") (len=" + strconv.Itoa(l)
	if !d.cs.DisableCapacities && v.Kind() == reflect.Slice {
		header += " cap=" + strconv.Itoa(c)
	}
	header += ") {\n"
	d.write(header)

	d.depth++
	d.pushIndent()
	for i := 0; i < l; i++ {
		d.writeIndent()
		d.dumpValue(v.Index(i))
		if i < l-1 {
			d.write(",")
		}
		d.write("\n")
	}
	d.popIndent()
	d.depth--

	d.writeIndent()
	d.write("}")
}

// dumpMap formats a map value.
func (d *dumpState) dumpMap(v reflect.Value) {
	t := v.Type()

	if v.IsNil() {
		d.write("(" + typeString(t) + ") <nil>")
		return
	}

	// Track map pointer for circular reference detection.
	ptr := v.Pointer()
	if d.seen[ptr] {
		d.write("(" + typeString(t) + ") <already shown>")
		return
	}
	d.seen[ptr] = true
	defer func() { delete(d.seen, ptr) }()

	keys := v.MapKeys()
	if d.cs.SortKeys {
		sortMapKeys(keys)
	}

	d.write("(" + typeString(t) + ") (len=" + strconv.Itoa(len(keys)) + ") {\n")

	d.depth++
	d.pushIndent()
	for i, key := range keys {
		d.writeIndent()
		d.write("(")
		d.dumpMapKey(key)
		d.write(") ")
		d.write(": (")
		d.dumpValue(v.MapIndex(key))
		d.write(")")
		if i < len(keys)-1 {
			d.write(",")
		}
		d.write("\n")
	}
	d.popIndent()
	d.depth--

	d.writeIndent()
	d.write("}")
}

// dumpMapKey formats a map key.
func (d *dumpState) dumpMapKey(v reflect.Value) {
	if d.cs.SpewKeys {
		d.dumpValue(v)
	} else {
		// Simple format for map keys.
		if v.CanInterface() {
			d.write(fmt.Sprintf("%v", v.Interface()))
		} else {
			d.dumpValue(v)
		}
	}
}

// dumpStruct formats a struct value.
func (d *dumpState) dumpStruct(v reflect.Value) {
	t := v.Type()
	d.write("(" + typeString(t) + ") {\n")

	d.depth++
	d.pushIndent()
	numFields := v.NumField()
	printed := 0
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		fv := v.Field(i)

		// OmitUnexported: skip unexported fields entirely.
		if d.cs.OmitUnexported && !field.IsExported() {
			continue
		}

		// OmitNilPointers: skip nil pointer fields.
		if d.cs.OmitNilPointers && fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}

		if printed > 0 {
			d.write(",\n")
		}
		printed++

		d.writeIndent()
		d.write(field.Name + ": ")

		// Handle unexported fields.
		if !fv.CanInterface() {
			if unsafeEnabled {
				fv = unsafeReflectValue(fv)
				if fv.CanInterface() {
					d.dumpValue(fv)
				} else {
					d.write("(" + typeString(fv.Type()) + ") <unexported>")
				}
			} else {
				d.write("(" + typeString(fv.Type()) + ") <unexported>")
			}
		} else {
			d.dumpValue(fv)
		}
	}
	if printed > 0 {
		d.write("\n")
	}
	d.popIndent()
	d.depth--

	d.writeIndent()
	d.write("}")
}

// dumpChan formats a channel value.
func (d *dumpState) dumpChan(v reflect.Value) {
	t := v.Type()
	if v.IsNil() {
		d.write("(" + typeString(t) + ") <nil>")
		return
	}

	var dir string
	switch t.ChanDir() {
	case reflect.BothDir:
		dir = "chan"
	case reflect.SendDir:
		dir = "chan<-"
	case reflect.RecvDir:
		dir = "<-chan"
	}

	_ = dir // dir is already part of typeString
	d.write("(" + typeString(t) + ")")
	if !d.cs.DisablePointerAddresses {
		d.write("(" + hexPointer(v) + ")")
	}
}

// dumpFunc formats a function value.
func (d *dumpState) dumpFunc(v reflect.Value) {
	t := v.Type()
	if v.IsNil() {
		d.write("(" + typeString(t) + ") <nil>")
		return
	}
	d.write("(" + typeString(t) + ")")
	if !d.cs.DisablePointerAddresses {
		d.write(" " + hexPointer(v))
	}
}


