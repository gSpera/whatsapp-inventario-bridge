//go:build safe || disableunsafe

package reveal

import "reflect"

// unsafeReflectValue is a no-op when unsafe is disabled.
// Unexported fields will show as "<unexported>" instead of their actual value.
func unsafeReflectValue(v reflect.Value) reflect.Value {
	return v
}

const unsafeEnabled = false
