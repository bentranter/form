// Package form provides a mechanism for automatically decoding HTTP request
// bodies into structs.
//
// Inspired by https://github.com/gorilla/schema.
package form

import (
	"errors"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/bentranter/terrible/internal/reflectx"
)

// Unmarshal maps an HTTP requests form data to the given interface. The given
// interface must be a pointer to a struct.
func Unmarshal(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Assumptions:
	//	1. v is a struct.
	//	2. That struct is a pointer receiver.
	//
	// With these assumptions in place, iterate over each field, and try to
	// parse it from the request.
	rv := reflect.ValueOf(v)
	elem, err := reflectx.GetStructPtr(rv)
	if err != nil {
		return err
	}

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)

		// If the field can't be set, any of the operations below will panic,
		// so we continue if that's the case.
		if !field.CanSet() {
			continue
		}

		key := elem.Type().Field(i).Name
		value := r.FormValue(key)

		// Set the field based on its kind. Its "kind" is its type.
		switch field.Kind() {
		case reflect.String:
			field.SetString(value)

		case reflect.Int, reflect.Int64:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return errors.New(`field ` + key + ` expects type int, but "` + value + `" is not an int`)
			}
			field.SetInt(i)

		case reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return errors.New(`field ` + key + ` expects type float64, but "` + value + `" is not a float64`)
			}
			field.SetFloat(f)

		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return errors.New(`field ` + key + ` expects type bool, but "` + value + `" is not a bool`)
			}
			field.SetBool(b)

		default:
			// ok
		}
	}

	return nil
}

type ForOpts struct {
	Method   string
	Action   string
	HTMLOpts map[string]string
}

// For generates the HTML form for v. If v is not a pointer to a struct, it
// **will** panic.
func For(v interface{}, opts ...*ForOpts) template.HTML {
	opt := &ForOpts{}
	for _, o := range opts {
		opt = o
	}

	// In Rails, form post to the current page, so we set that as the default
	// behaviour as well.
	if opt.Method == "" {
		opt.Method = http.MethodPost
	}
	if opt.Action == "" {
		opt.Action = "/"
	}

	rv := reflect.ValueOf(v)
	elem, err := reflectx.GetStructFromPtr(rv)
	if err != nil {
		panic(err)
	}

	// Write the top-level form tag.
	builder := &strings.Builder{}
	builder.WriteString(`<form method="`)
	builder.WriteString(opt.Method)
	builder.WriteString(`" action="`)
	builder.WriteString(opt.Action)
	builder.WriteString(`">
`)

	// Write each form field.
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)

		// If the field can't "interface", then that field isn't exported, and
		// will panic if we try to access it, so we don't touch these fields.
		// so we continue if that's the case.
		if !field.CanInterface() {
			continue
		}

		key := elem.Type().Field(i).Name
		// If the field isn't zeroed or unset, then assign a value.
		value := ""
		if field.Interface() != reflect.Zero(field.Type()).Interface() {
			value = field.String()
		}
		builder.WriteString(`  <label for="` + key + `">` + key + `</label>
`)
		builder.WriteString(`  <input type="text" id="` + key + `" name="` + key + `" value="` + value + `"/>

`)
	}

	// Close the top-level form tag.
	builder.WriteString(`  <input type="submit" value="Submit">
</form>
`)

	return template.HTML(builder.String())
}
