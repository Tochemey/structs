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
	// DefaultTagName is the default tag name for struct fields which provides
	// a more granular to tweak certain structs. Lookup the necessary functions
	// for more info.
	DefaultTagName = "structs" // struct's field default tag name

	errNotSlice        = errors.New("not a slice")
	errNotMap          = errors.New("not a map")
	errNotStruct       = errors.New("not a struct")
	errNotArrayOrSlice = errors.New("not an array or slice")
)

// Struct encapsulates a struct type to provide several high level functions
// around the struct.
type Struct struct {
	raw     any
	value   reflect.Value
	TagName string
}

// New returns a new *Struct with the struct s. It panics if the s's kind is
// not struct.
func New(s any) *Struct {
	return &Struct{
		raw:     s,
		value:   structVal(s),
		TagName: DefaultTagName,
	}
}

// Map converts the given struct to a map[string]any, where the keys
// of the map are the field names and the values of the map the associated
// values of the fields. The default key string is the struct field name but
// can be changed in the struct field's tag value. The "structs" key in the
// struct's field tag value is the key name. Example:
//
//	// Field appears in map as key "myName".
//	Name string `structs:"myName"`
//
// A tag value with the content of "-" ignores that particular field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// A tag value with the content of "string" uses the stringer to get the value. Example:
//
//	// The value will be output of Animal's String() func.
//	// Map will panic if Animal does not implement String().
//	Field *Animal `structs:"field,string"`
//
// A tag value with the option of "flatten" used in a struct field is to flatten its fields
// in the output map. Example:
//
//	// The FieldStruct's fields will be flattened into the output map.
//	FieldStruct time.Time `structs:",flatten"`
//
// A tag value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//	// Field is not processed further by this package.
//	Field time.Time     `structs:"myName,omitnested"`
//	Field *http.Request `structs:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field if
// the field value is empty. Example:
//
//	// Field appears in map as key "myName", but the field is
//	// skipped if empty.
//	Field string `structs:"myName,omitempty"`
//
//	// Field appears in map as key "Field" (the default), but
//	// the field is skipped if empty.
//	Field string `structs:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields will be neglected.
func (s *Struct) Map() map[string]any {
	out := make(map[string]any)
	s.FillMap(out)
	return out
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func (s *Struct) FillMap(out map[string]any) {
	if out == nil {
		return
	}

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		isSubStruct := false
		var finalVal any

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		if !tagOpts.Has("omitnested") {
			finalVal = s.nested(val)

			v := reflect.ValueOf(val.Interface())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			switch v.Kind() {
			case reflect.Map, reflect.Struct:
				isSubStruct = true
			default:
				// pass
			}
		} else {
			finalVal = val.Interface()
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				out[name] = s.String()
			}
			continue
		}

		if isSubStruct && (tagOpts.Has("flatten")) {
			for k := range finalVal.(map[string]any) {
				out[k] = finalVal.(map[string]any)[k]
			}
		} else {
			out[name] = finalVal
		}
	}
}

// Values converts the given s struct's field values to a []any.  A
// struct tag with the content of "-" ignores the that particular field.
// Example:
//
//	// Field is ignored by this package.
//	Field int `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//	// Fields is not processed further by this package.
//	Field time.Time     `structs:",omitnested"`
//	Field *http.Request `structs:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field and
// is not added to the values if the field value is empty. Example:
//
//	// Field is skipped if empty
//	Field string `structs:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected.
func (s *Struct) Values() []any {
	fields := s.structFields()

	var t []any

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				t = append(t, s.String())
			}
			continue
		}

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			// look out for embedded structs, and convert them to a
			// []any to be added to the final values slice
			t = append(t, Values(val.Interface())...)
		} else {
			t = append(t, val.Interface())
		}
	}

	return t
}

// Fields returns a slice of Fields. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Fields() []*Field {
	return getFields(s.value, s.TagName)
}

// Names returns a slice of field names. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Names() []string {
	fields := getFields(s.value, s.TagName)

	names := make([]string, len(fields))

	for i, field := range fields {
		names[i] = field.Name()
	}

	return names
}

// Field returns a new Field struct that provides several high level functions
// around a single struct field entity. It panics if the field is not found.
func (s *Struct) Field(name string) *Field {
	f, ok := s.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return f
}

// FieldOk returns a new Field struct that provides several high level functions
// around a single struct field entity. The boolean returns true if the field
// was found.
func (s *Struct) FieldOk(name string) (*Field, bool) {
	t := s.value.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &Field{
		field:      field,
		value:      s.value.FieldByName(name),
		defaultTag: s.TagName,
	}, true
}

// IsZero returns true if all fields in a struct is a zero value (not
// initialized) A struct tag with the content of "-" ignores the checking of
// that particular field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//	// Field is not processed further by this package.
//	Field time.Time     `structs:"myName,omitnested"`
//	Field *http.Request `structs:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) IsZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			ok := IsZero(val.Interface())
			if !ok {
				return false
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if !reflect.DeepEqual(current, zero) {
			return false
		}
	}

	return true
}

// HasZero returns true if a field in a struct is not initialized (zero value).
// A struct tag with the content of "-" ignores the checking of that particular
// field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//	// Field is not processed further by this package.
//	Field time.Time     `structs:"myName,omitnested"`
//	Field *http.Request `structs:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) HasZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			ok := HasZero(val.Interface())
			if ok {
				return true
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if reflect.DeepEqual(current, zero) {
			return true
		}
	}

	return false
}

// Name returns the structs's type name within its package. For more info refer
// to Name() function.
func (s *Struct) Name() string {
	return s.value.Type().Name()
}

// Original returns the underlying struct
func (s *Struct) Original() any {
	return s.raw
}

// structFields returns the exported struct fields for a given s struct. This
// is a convenient helper method to avoid duplicate code in some of the
// functions.
func (s *Struct) structFields() []reflect.StructField {
	t := s.value.Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(s.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

// nested retrieves recursively all types for the given value and returns the
// nested value.
func (s *Struct) nested(val reflect.Value) any {
	var finalVal any

	v := reflect.ValueOf(val.Interface())
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		n := New(val.Interface())
		n.TagName = s.TagName
		m := n.Map()

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		default:
			// pass
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]any, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}
			finalVal = m
			break
		}

		// TODO(arslan): should this be optional?
		finalVal = val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			finalVal = val.Interface()
			break
		}

		// TODO(arslan): should this be optional?
		// do not iterate of non struct types, just pass the value. Ie: []int,
		// []string, co... We only iterate further if it's a struct.
		// i.e []foo or []*foo
		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		slices := make([]any, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}
		finalVal = slices
	default:
		finalVal = val.Interface()
	}

	return finalVal
}

func structVal(s any) reflect.Value {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

// Map converts the given struct to a map[string]any. For more info
// refer to Struct types Map() method. It panics if s's kind is not struct.
func Map(s any) map[string]any {
	return New(s).Map()
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func FillMap(s any, out map[string]any) {
	New(s).FillMap(out)
}

// FillStruct a given struct with the provide map in place.
// It panics in case of error
func FillStruct(m map[string]any, s any) {
	if err := toStruct(m, structVal(s)); err != nil {
		panic(err)
	}
}

// Values converts the given struct to a []any. For more info refer to
// Struct types Values() method.  It panics if s's kind is not struct.
func Values(s any) []any {
	return New(s).Values()
}

// Fields returns a slice of *Field. For more info refer to Struct types
// Fields() method.  It panics if s's kind is not struct.
func Fields(s any) []*Field {
	return New(s).Fields()
}

// Names returns a slice of field names. For more info refer to Struct types
// Names() method.  It panics if s's kind is not struct.
func Names(s any) []string {
	return New(s).Names()
}

// IsZero returns true if all fields is equal to a zero value. For more info
// refer to Struct types IsZero() method.  It panics if s's kind is not struct.
func IsZero(s any) bool {
	return New(s).IsZero()
}

// HasZero returns true if any field is equal to a zero value. For more info
// refer to Struct types HasZero() method.  It panics if s's kind is not struct.
func HasZero(s any) bool {
	return New(s).HasZero()
}

// IsStruct returns true if the given variable is a struct or a pointer to
// struct.
func IsStruct(s any) bool {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// uninitialized zero value of a struct
	if v.Kind() == reflect.Invalid {
		return false
	}

	return v.Kind() == reflect.Struct
}

// Name returns the structs's type name within its package. It returns an
// empty string for unnamed types. It panics if s's kind is not struct.
func Name(s any) string {
	return New(s).Name()
}

func fromPtr(in any, t reflect.Type, out reflect.Value) error {
	child := reflect.New(t.Elem())
	if err := fromValue(in, child.Elem(), child.Elem().Type()); err != nil {
		return err
	}
	out.Set(child)
	return nil
}

func fromSlice(in any, out reflect.Value, t reflect.Type) (err error) {
	input := reflect.ValueOf(in)
	if input.Kind() != reflect.Slice {
		return errNotSlice
	}

	output := reflect.MakeSlice(t, input.Len(), input.Cap())
	for i := 0; i < input.Len(); i++ {
		inputValue := reflect.ValueOf(input.Index(i).Interface())
		elem := reflect.New(output.Index(i).Type()).Elem()
		if e := fromValue(inputValue.Interface(), elem, elem.Type()); e != nil {
			err = errors.Join(err, e)
			continue
		}

		output.Index(i).Set(elem)
	}

	if err == nil {
		out.Set(output)
	}

	return
}

func fromMap(in any, out reflect.Value, t reflect.Type) (err error) {
	input := reflect.ValueOf(in)
	if input.Kind() != reflect.Map {
		return errNotMap
	}

	output := reflect.MakeMap(t)
	for _, key := range input.MapKeys() {
		value := reflect.ValueOf(key.Interface())
		iface := value.Interface()
		outKey := reflect.New(value.Type()).Elem()
		if e := fromValue(iface, outKey, outKey.Type()); e != nil {
			err = errors.Join(err, e)
			continue
		}

		inputValue := reflect.ValueOf(input.MapIndex(value).Interface()).Interface()
		outputValue := reflect.New(output.Type().Elem()).Elem()
		if e := fromValue(inputValue, outputValue, outputValue.Type()); e != nil {
			err = errors.Join(err, e)
			continue
		}

		output.SetMapIndex(outKey, outputValue)
	}

	if err == nil {
		// Special case: out may be a struct or struct pointer...
		out.Set(output)
	}

	return
}

func fromArray(in any, out reflect.Value, t reflect.Type) (err error) {
	input := reflect.ValueOf(in)
	if input.Kind() != reflect.Array && input.Kind() != reflect.Slice {
		return errNotArrayOrSlice
	}

	output := reflect.New(t).Elem()
	for i := 0; i < input.Len(); i++ {
		outputValue := output.Index(i)
		inputValue := input.Index(i)
		if e := fromValue(inputValue.Interface(), outputValue, outputValue.Type()); e != nil {
			err = errors.Join(err, fmt.Errorf("%v:(%s)", e, fmt.Sprintf("@%d", i)))
			continue
		}
	}

	if err == nil {
		out.Set(output)
	}

	return
}

func fromValue(in any, out reflect.Value, t reflect.Type) error {
	switch out.Kind() {
	case reflect.Ptr:
		return fromPtr(in, t, out)
	case reflect.Struct:
		return toStruct(in, out)
	case reflect.Slice:
		return fromSlice(in, out, t)
	case reflect.Map:
		return fromMap(in, out, t)
	case reflect.Array:
		return fromArray(in, out, t)
	default:
		// pass
	}

	inputValue := reflect.ValueOf(in)
	inputType := inputValue.Type()
	outputType := reflect.ValueOf(out.Interface()).Type()

	if inputType == outputType {
		// default case: copy the value over
		out.Set(reflect.ValueOf(in))
		return nil
	}

	if inputType.AssignableTo(outputType) {
		// types are assignable
		out.Set(inputValue)
		return nil
	}

	if inputType.ConvertibleTo(outputType) {
		// types are convertible
		out.Set(inputValue.Convert(outputType))
		return nil
	}

	return fmt.Errorf("type mismatch: %s and %s are incompatible", outputType.String(), inputType.String())
}

// toStruct fills a given struct with the provided map values
func toStruct(in any, s reflect.Value) (err error) {
	if in == nil {
		return errors.New("input data is nil")
	}

	// make sure input is a map
	if reflect.ValueOf(in).Kind() != reflect.Map {
		return errNotMap
	}

	// if target is a pointer to a struct: create a new instance
	if s.Kind() == reflect.Ptr {
		s.Set(reflect.New(s.Type().Elem()))
		s = s.Elem()
	}

	if s.Kind() != reflect.Struct {
		return errNotStruct
	}

	// get the all the exported fields of th passed struct
	fields := getFields(s, DefaultTagName)

	// Hold the values of the modified fields in a map, which will be applied shortly before
	// this function returns.
	// This ensures we do not modify the target struct at all in case of an error
	modifiedFields := make(map[int]reflect.Value, len(fields))
	for i, field := range fields {
		name := field.Name()
		val := s.FieldByName(name)

		if field.IsEmbedded() {
			if e := toStruct(in, val); e != nil {
				err = errors.Join(err, e)
				continue
			}
			continue
		}
		// ignore unexported field
		if !field.IsExported() {
			continue
		}

		// handle value struct
		if field.Kind() == reflect.Struct {
			if e := toStruct(in, val); e != nil {
				err = errors.Join(err, e)
				continue
			}
			continue
		}

		// interfaces are not supported
		if field.Kind() == reflect.Interface {
			err = errors.Join(err, fmt.Errorf("interface not supported:(%s)", name))
			continue
		}

		// look up value of "fieldName" in map
		mapVal := reflect.ValueOf(in).MapIndex(reflect.ValueOf(name))
		if !mapVal.IsValid() {
			// value not in map, ignore it
			continue
		}

		fieldType := val.Type()
		elem := reflect.New(fieldType).Elem()
		value := mapVal.Interface()
		if e := fromValue(value, elem, fieldType); e != nil {
			err = errors.Join(err, fmt.Errorf("%v:(%s)", e, name))
			continue
		}

		modifiedFields[i] = elem
	}

	// Apply changes to all modified fields in case no error happened during processing.
	if err == nil {
		// Apply changes to all modified fields
		for index, value := range modifiedFields {
			s.Field(index).Set(value)
		}
	}
	return
}
