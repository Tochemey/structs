# Structs

[![build](https://img.shields.io/github/actions/workflow/status/Tochemey/structs/build.yml?branch=main)](https://github.com/Tochemey/structs/actions/workflows/build.yml)
[![GitHub go.mod Go version](https://badges.chse.dev/github/go-mod/go-version/Tochemey/structs)](https://go.dev/doc/install)


## Overview
Structs is a forked copy of the original structs library written by [Fatih Arslan](https://github.com/fatih/structs)
Structs contains various utilities to work with Go (Golang) structs. 

## Addition to the original library

- `FillStruct`: this method helps fill a given struct from a map.

## Install

```bash
go get github.com/tochemey/structs
```

## Usage and Examples

Just like the standard lib `strings`, `bytes` and co packages, `structs` has
many global functions to manipulate or organize your struct data. Lets define
and declare a struct:

```go
type Server struct {
	Name        string `json:"name,omitempty"`
	ID          int
	Enabled     bool
	users       []string // not exported
	http.Server          // embedded
}

server := &Server{
	Name:    "gopher",
	ID:      123456,
	Enabled: true,
}
```

```go
// Convert a struct to a map[string]interface{}
// => {"Name":"gopher", "ID":123456, "Enabled":true}
m := structs.Map(server)

// Convert the values of a struct to a []interface{}
// => ["gopher", 123456, true]
v := structs.Values(server)

// Convert the names of a struct to a []string
// (see "Names methods" for more info about fields)
n := structs.Names(server)

// Convert the values of a struct to a []*Field
// (see "Field methods" for more info about fields)
f := structs.Fields(server)

// Return the struct name => "Server"
n := structs.Name(server)

// Check if any field of a struct is initialized or not.
h := structs.HasZero(server)

// Check if all fields of a struct is initialized or not.
z := structs.IsZero(server)

// Check if server is a struct or a pointer to struct
i := structs.IsStruct(server)

// Sets the fields of a given struct from a map
m := map[string]any{"Name":"gopher", "ID":123456, "Enabled":false}
structs.FillStruct(m, server)
```

### Struct methods

The structs functions can be also used as independent methods by creating a new
`*structs.Struct`. This is handy if you want to have more control over the
structs (such as retrieving a single Field).

```go
// Create a new struct type:
s := structs.New(server)

m := s.Map()              // Get a map[string]interface{}
v := s.Values()           // Get a []interface{}
f := s.Fields()           // Get a []*Field
n := s.Names()            // Get a []string
f := s.Field(name)        // Get a *Field based on the given field name
f, ok := s.FieldOk(name)  // Get a *Field based on the given field name
n := s.Name()             // Get the struct name
h := s.HasZero()          // Check if any field is uninitialized
z := s.IsZero()           // Check if all fields are uninitialized
```

### Field methods

We can easily examine a single Field for more detail. Below you can see how we
get and interact with various field methods:


```go
s := structs.New(server)

// Get the Field struct for the "Name" field
name := s.Field("Name")

// Get the underlying value,  value => "gopher"
value := name.Value().(string)

// Set the field's value
name.Set("another gopher")

// Get the field's kind, kind =>  "string"
name.Kind()

// Check if the field is exported or not
if name.IsExported() {
	fmt.Println("Name field is exported")
}

// Check if the value is a zero value, such as "" for string, 0 for int
if !name.IsZero() {
	fmt.Println("Name is initialized")
}

// Check if the field is an anonymous (embedded) field
if !name.IsEmbedded() {
	fmt.Println("Name is not an embedded field")
}

// Get the Field's tag value for tag name "json", tag value => "name,omitempty"
tagValue := name.Tag("json")
```

Nested structs are supported too:

```go
addrField := s.Field("Server").Field("Addr")

// Get the value for addr
a := addrField.Value().(string)

// Or get all fields
httpServer := s.Field("Server").Fields()
```

We can also get a slice of Fields from the Struct type to iterate over all
fields. This is handy if you wish to examine all fields:

```go
s := structs.New(server)

for _, f := range s.Fields() {
	fmt.Printf("field name: %+v\n", f.Name())

	if f.IsExported() {
		fmt.Printf("value   : %+v\n", f.Value())
		fmt.Printf("is zero : %+v\n", f.IsZero())
	}
}
```

## Credits

 * [Arsene Gandote](https://github.com/Tochemey)
 * [Fatih Arslan](https://github.com/fatih)
 * [Cihangir Savas](https://github.com/cihangir)
 * [Stephan Peijnik-Steinwender ](https://github.com/speijnik): Borrowed some of his work [here](https://github.com/anexia-it/go-structmapper)

## License

The MIT License (MIT) - see LICENSE.md for more details
