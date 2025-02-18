/*
Basic json-schema generator based on Go types, for easy interchange of Go
structures between diferent languages.
*/
package jsonschema

import (
	"encoding/json"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"reflect"
	"strings"
)

const DEFAULT_SCHEMA = "http://json-schema.org/schema#"

type Document struct {
	Schema string `json:"$schema,omitempty"`
	property
}

// Reads the variable structure into the JSON-Schema Document
func (d *Document) Read(variable interface{}) {
	d.setDefaultSchema()

	value := reflect.ValueOf(variable)
	d.read(value.Type(), "")
}

// ReadDeep reads the variable structure into the JSON-Schema Document
func (d *Document) ReadDeep(variable interface{}) {
	d.setDefaultSchema()

	value := reflect.ValueOf(variable)
	d.readDeep(value, "")
}

func (d *Document) setDefaultSchema() {
	if d.Schema == "" {
		d.Schema = DEFAULT_SCHEMA
	}
}

// Marshal returns the JSON encoding of the Document
func (d *Document) Marshal() ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

// String return the JSON encoding of the Document as a string
func (d *Document) String() string {
	jsonBytes, _ := d.Marshal()
	return string(jsonBytes)
}

type property struct {
	Type                 string                                    `json:"type,omitempty"`
	Format               string                                    `json:"format,omitempty"`
	Items                *property                                 `json:"items,omitempty"`
	Properties           *orderedmap.OrderedMap[string, *property] `json:"properties,omitempty"`
	Required             []string                                  `json:"required,omitempty"`
	AdditionalProperties bool                                      `json:"additionalProperties,omitempty"`
}

func (p *property) read(t reflect.Type, opts tagOptions) {
	jsType, format, kind := getTypeFromMapping(t)
	if jsType != "" {
		p.Type = jsType
	}
	if format != "" {
		p.Format = format
	}

	switch kind {
	case reflect.Slice:
		p.readFromSlice(t)
	case reflect.Map:
		p.readFromMap(t)
	case reflect.Struct:
		p.readFromStruct(t)
	case reflect.Ptr:
		p.read(t.Elem(), opts)
	}
}

func (p *property) readDeep(v reflect.Value, opts tagOptions) {
	if !v.IsValid() {
		p.Type = "null"
		return
	}
	jsType, format, kind := getTypeFromMapping(v.Type())
	if jsType != "" {
		p.Type = jsType
	}
	if format != "" {
		p.Format = format
	}

	if v.Kind() == reflect.Struct && v.Type() == reflect.TypeOf(orderedmap.OrderedMap[string, interface{}]{}) {
		p.readFromMapDeep(v)
		return
	}

	switch kind {
	case reflect.Slice:
		p.readFromSliceDeep(v)
	case reflect.Map:
		p.readFromMapDeep(v)
	case reflect.Struct:
		p.readFromStructDeep(v)
	case reflect.Ptr, reflect.Interface:
		p.readDeep(v.Elem(), opts)
	}
}

func (p *property) readFromSlice(t reflect.Type) {
	jsType, _, kind := getTypeFromMapping(t.Elem())
	if kind == reflect.Uint8 {
		p.Type = "string"
	} else if jsType != "" {
		p.Items = &property{}
		p.Items.read(t.Elem(), "")
	}
}

func (p *property) readFromSliceDeep(v reflect.Value) {
	if v.Len() == 0 {
		t := v.Type()
		jsType, _, kind := getTypeFromMapping(t.Elem())
		if kind == reflect.Uint8 {
			p.Type = "string"
		} else if jsType != "" {
			p.Items = &property{}
			if v.Len() == 0 {
				p.Items.read(t.Elem(), "")
				return
			}
			p.Items.readDeep(v.Index(0), "")
		}
		return
	}

	_, _, kind := getTypeFromMapping(v.Index(0).Type())
	if kind == reflect.Uint8 {
		p.Type = "string"
	} else {
		p.Items = &property{}
		p.Items.readDeep(v.Index(0), "")
	}
}

func (p *property) readFromMap(t reflect.Type) {
	jsType, format, _ := getTypeFromMapping(t.Elem())

	if jsType != "" {
		p.Properties = orderedmap.New[string, *property]()
		p.Properties.Set(".*", &property{Type: jsType, Format: format})
	} else {
		p.AdditionalProperties = true
	}
}

func (p *property) readFromMapDeep(v reflect.Value) {

	properties := orderedmap.New[string, *property]()
	vv, ok := v.Interface().(orderedmap.OrderedMap[string, interface{}])
	if ok {
		for pair := vv.Oldest(); pair != nil; pair = pair.Next() {
			keyName := pair.Key
			value := pair.Value
			pro := &property{}
			pro.readDeep(reflect.ValueOf(value), "")
			properties.Set(keyName, pro)
		}

		if properties.Len() > 0 {
			p.Properties = properties
		}
	} else {
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			keyName := mapKeyToString(key)
			pro := &property{}
			pro.readDeep(value, "")
			properties.Set(keyName, pro)
		}

		if properties.Len() > 0 {
			p.Properties = properties
		}
	}

}

func mapKeyToString(key reflect.Value) string {
	keyKind := key.Kind()

	if keyKind == reflect.Interface {
		return mapKeyToString(key.Elem())
	}

	return key.String()
}

func (p *property) readFromStruct(t reflect.Type) {
	p.Type = "object"
	p.Properties = orderedmap.New[string, *property]()
	p.AdditionalProperties = false

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("json")
		name, opts := parseTag(tag)
		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}

		if field.Anonymous {
			embeddedProperty := &property{}
			embeddedProperty.read(field.Type, opts)

			for pair := embeddedProperty.Properties.Oldest(); pair != nil; pair = pair.Next() {
				p.Properties.Set(pair.Key, pair.Value)
			}

			p.Required = append(p.Required, embeddedProperty.Required...)

			continue
		}

		pro := &property{}
		pro.read(field.Type, opts)
		p.Properties.Set(name, pro)

		if !opts.Contains("omitempty") {
			p.Required = append(p.Required, name)
		}
	}
}

func (p *property) readFromStructDeep(v reflect.Value) {
	t := v.Type()
	p.Type = "object"
	p.Properties = orderedmap.New[string, *property]()
	p.AdditionalProperties = false

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("json")
		name, opts := parseTag(tag)
		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}

		if field.Anonymous {
			embeddedProperty := &property{}
			embeddedProperty.readDeep(v.Field(i), opts)

			for pair := embeddedProperty.Properties.Oldest(); pair != nil; pair = pair.Next() {
				p.Properties.Set(pair.Key, pair.Value)
			}
			p.Required = append(p.Required, embeddedProperty.Required...)

			continue
		}

		pro := &property{}
		pro.read(field.Type, opts)
		p.Properties.Set(name, pro)

		if !opts.Contains("omitempty") {
			p.Required = append(p.Required, name)
		}
	}
}

var formatMapping = map[string][]string{
	"time.Time": {"string", "date-time"},
}

var kindMapping = map[reflect.Kind]string{
	reflect.Bool:    "boolean",
	reflect.Int:     "integer",
	reflect.Int8:    "integer",
	reflect.Int16:   "integer",
	reflect.Int32:   "integer",
	reflect.Int64:   "integer",
	reflect.Uint:    "integer",
	reflect.Uint8:   "integer",
	reflect.Uint16:  "integer",
	reflect.Uint32:  "integer",
	reflect.Uint64:  "integer",
	reflect.Float32: "number",
	reflect.Float64: "number",
	reflect.String:  "string",
	reflect.Slice:   "array",
	reflect.Struct:  "object",
	reflect.Map:     "object",
}

func getTypeFromMapping(t reflect.Type) (string, string, reflect.Kind) {
	if v, ok := formatMapping[t.String()]; ok {
		return v[0], v[1], reflect.String
	}

	kind := t.Kind()
	if v, ok := kindMapping[kind]; ok {
		return v, "", kind
	}

	return "", "", kind
}

type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, ""
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}

	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
