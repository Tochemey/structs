/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Fatih Arslan
 * Copyright (c) 2024 Arsene Tochemey
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package structs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errNotExported = errors.New("field is not exported")
	errNotSettable = errors.New("field is not settable")
)

// Field represents a single struct field that encapsulates high level
// functions around the field.
type Field struct {
	value      reflect.Value
	field      reflect.StructField
	defaultTag string
}

// Tag returns the value associated with key in the tag string. If there is no
// such key in the tag, Tag returns the empty string.
func (f *Field) Tag(key string) string {
	return f.field.Tag.Get(key)
}

// Value returns the underlying value of the field. It panics if the field
// is not exported.
func (f *Field) Value() any {
	return f.value.Interface()
}

// IsEmbedded returns true if the given field is an anonymous field (embedded)
func (f *Field) IsEmbedded() bool {
	return f.field.Anonymous
}

// IsExported returns true if the given field is exported.
func (f *Field) IsExported() bool {
	return f.field.PkgPath == ""
}

// IsZero returns true if the given field is not initialized (has a zero value).
// It panics if the field is not exported.
func (f *Field) IsZero() bool {
	zero := reflect.Zero(f.value.Type()).Interface()
	current := f.Value()

	return reflect.DeepEqual(current, zero)
}

// Name returns the name of the given field
func (f *Field) Name() string {
	return f.field.Name
}

// Kind returns the fields kind, such as "string", "map", "bool", etc ..
func (f *Field) Kind() reflect.Kind {
	return f.value.Kind()
}

// Set sets the field to given value v. It returns an error if the field is not
// settable (not addressable or not exported) or if the given value's type
// doesn't match the fields type.
func (f *Field) Set(val any) error {
	// we can't set unexported fields, so be sure this field is exported
	if !f.IsExported() {
		return errNotExported
	}

	// do we get here? not sure...
	if !f.value.CanSet() {
		return errNotSettable
	}

	given := reflect.ValueOf(val)

	if f.value.Kind() != given.Kind() {
		return fmt.Errorf("wrong kind. got: %s want: %s", given.Kind(), f.value.Kind())
	}

	f.value.Set(given)
	return nil
}

// Zero sets the field to its zero value. It returns an error if the field is not
// settable (not addressable or not exported).
func (f *Field) Zero() error {
	zero := reflect.Zero(f.value.Type()).Interface()
	return f.Set(zero)
}

// Fields returns a slice of Fields. This is particular handy to get the fields
// of a nested struct . A struct tag with the content of "-" ignores the
// checking of that particular field. Example:
//
//	// Field is ignored by this package.
//	Field *http.Request `structs:"-"`
//
// It panics if field is not exported or if field's kind is not struct
func (f *Field) Fields() []*Field {
	return getFields(f.value, f.defaultTag)
}

// Field returns the field from a nested struct. It panics if the nested struct
// is not exported or if the field was not found.
func (f *Field) Field(name string) *Field {
	field, ok := f.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return field
}

// FieldOk returns the field from a nested struct. The boolean returns whether
// the field was found (true) or not (false).
func (f *Field) FieldOk(name string) (*Field, bool) {
	value := &f.value
	// value must be settable so we need to make sure it holds the address of the
	// variable and not a copy, so we can pass the pointer to structVal instead of a
	// copy (which is not assigned to any variable, hence not settable).
	// see "https://blog.golang.org/laws-of-reflection#TOC_8."
	if f.value.Kind() != reflect.Ptr {
		a := f.value.Addr()
		value = &a
	}
	v := structVal(value.Interface())
	t := v.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &Field{
		field: field,
		value: v.FieldByName(name),
	}, true
}

func getFields(v reflect.Value, tagName string) []*Field {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	var fields []*Field

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}

		f := &Field{
			field: field,
			value: v.FieldByName(field.Name),
		}

		fields = append(fields, f)
	}

	return fields
}
