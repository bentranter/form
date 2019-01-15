package reflectx

import (
	"errors"
	"reflect"
)

// GetStructFromPtr checks to see if rv is a pointer to a struct. If it is, it
// returns its underlying value. If not, it returns the original value, along
// with an error message describing what the expected type for rv is, and what
// the actual type is.
func GetStructFromPtr(rv reflect.Value) (reflect.Value, error) {
	kind := rv.Kind()
	if kind != reflect.Ptr {
		return rv, errors.New("expected interface to be a pointer to a struct, but got a " + kind.String())
	}

	elem := rv.Elem()
	underlyingKind := elem.Kind()
	if underlyingKind != reflect.Struct {
		return rv, errors.New("expected interface to be a pointer to a struct, but got a pointer to a " + underlyingKind.String())
	}

	return elem, nil
}
