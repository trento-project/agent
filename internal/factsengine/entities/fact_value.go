package entities

import (
	"encoding/gob"
	"strconv"
)

func init() {
	gob.Register(&FactValueInt{})
	gob.Register(&FactValueFloat{})
	gob.Register(&FactValueString{})
	gob.Register(&FactValueBool{})
	gob.Register(&FactValueList{})
	gob.Register(&FactValueMap{})
}

// `FactValue` represents a dynamically typed value which can be either
// an int, a float, a string, a boolean, a recursive map[string] value, or a
// list of values.
// A producer of FactValue is expected to set one of that variants.
type FactValue interface {
	isFactValue()
	AsInterface() interface{}
}

type FactValueInt struct {
	Value int
}

func (v *FactValueInt) isFactValue() {}

// AsInterface converts a FactValueInt internal value to an interface{}.
func (v *FactValueInt) AsInterface() interface{} {
	return v.Value
}

type FactValueFloat struct {
	Value float64
}

func (v *FactValueFloat) isFactValue() {}

// AsInterface converts a FactValueFloat internal value to an interface{}.
func (v *FactValueFloat) AsInterface() interface{} {
	return v.Value
}

type FactValueBool struct {
	Value bool
}

func (v *FactValueBool) isFactValue() {}

// AsInterface converts a FactValueBool internal value to an interface{}.
func (v *FactValueBool) AsInterface() interface{} {
	return v.Value
}

type FactValueString struct {
	Value string
}

func (v *FactValueString) isFactValue() {}

// AsInterface converts a FactValueString internal value to an interface{}.
func (v *FactValueString) AsInterface() interface{} {
	return v.Value
}

type FactValueMap struct {
	Value map[string]FactValue
}

func (v *FactValueMap) isFactValue() {}

// AsInterface converts a FactValueMap internal value to an interface{}.
func (v *FactValueMap) AsInterface() interface{} {
	result := make(map[string]interface{})
	for key, value := range v.Value {
		result[key] = value.AsInterface()
	}
	return result
}

type FactValueList struct {
	Value []FactValue
}

func (v *FactValueList) isFactValue() {}

// AsInterface converts a FactValueList internal value to an interface{}.
func (v *FactValueList) AsInterface() interface{} {
	result := []interface{}{}
	for _, item := range v.Value {
		result = append(result, item.AsInterface())
	}
	return result
}

// ParseStringToFactValue parses a string to a FactValue type.
func ParseStringToFactValue(str string) FactValue {
	if i, err := strconv.Atoi(str); err == nil {
		return &FactValueInt{Value: i}
	} else if b, err := strconv.ParseBool(str); err == nil {
		return &FactValueBool{Value: b}
	} else if f, err := strconv.ParseFloat(str, 64); err == nil {
		return &FactValueFloat{Value: f}
	}

	return &FactValueString{Value: str}
}
