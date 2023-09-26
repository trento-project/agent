package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
)

// nolint:gochecknoglobals
// ValueNotFoundError is an error returned when the wanted value in GetValue
// function is not found
var ValueNotFoundError = FactGatheringError{
	Type:    "value-not-found",
	Message: "error getting value",
}

type Conf struct {
	StringConversion bool // Enable string automatic conversion to to a numeric fact value
	SnakeCaseKeys    bool // snake_case keys if the provided value is a map
}

func NewDefaultConf() Conf {
	return Conf{
		StringConversion: true,
		SnakeCaseKeys:    false,
	}
}

// FactValue represents a dynamically typed value which can be either
// an int, a float, a string, a boolean, a recursive map[string] value, or a
// list of values.
// A producer of FactValue is expected to set one of that variants.
type FactValue interface {
	isFactValue()
	AsInterface() interface{}
}

// NewFactValue constructs a FactValue from a nested interface with a provided configuration
func NewFactValue(factInterface interface{}, conf *Conf) (FactValue, error) {
	switch value := factInterface.(type) {
	case []interface{}:
		newList := []FactValue{}
		for _, value := range value {
			newValue, err := NewFactValue(value, conf)
			if err != nil {
				return nil, err
			}
			newList = append(newList, newValue)
		}
		return &FactValueList{Value: newList}, nil
	case map[string]interface{}:
		newMap := make(map[string]FactValue)
		for key, mapValue := range value {
			newValue, err := NewFactValue(mapValue, conf)
			if err != nil {
				return nil, err
			}
			if conf.SnakeCaseKeys {
				newMap[strcase.ToSnake(key)] = newValue
			} else {
				newMap[key] = newValue
			}
		}
		return &FactValueMap{Value: newMap}, nil
	case bool, int, int32, int64, uint, uint32, uint64, float32, float64:
		return ParseStringToFactValue(fmt.Sprint(value)), nil
	case string:
		if conf.StringConversion {
			return ParseStringToFactValue(value), nil
		}
		return &FactValueString{Value: value}, nil
	default:
		return nil, fmt.Errorf("invalid type: %T for value: %v", value, factInterface)
	}
}

// NewFactValueWithDefaultConf constructs a FactValue from a nested interface.
func NewFactValueWithDefaultConf(factInterface interface{}) (FactValue, error) {
	conf := NewDefaultConf()
	return NewFactValue(factInterface, &conf)
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

// GetValue returns a value using a dot access key format from a FactValue.
// Examples:
// foo.bar.buz access the {"foo": {"bar": {"baz": "value"}}}
// foo.0.buz access the {"foo": [{"buz": "value"}]}
func (v *FactValueMap) GetValue(values string) (FactValue, *FactGatheringError) {
	// splitDotAccess returns and empty list if the coming argument is an empty string.
	// It is used to replace strings.Split as this 2nd returns a one element list with
	// and empty string in the same scenario
	splitDotAccess := func(c rune) bool {
		return c == '.'
	}

	value, err := getValue(v, strings.FieldsFunc(values, splitDotAccess))
	if err != nil {
		return value, ValueNotFoundError.Wrap(fmt.Sprintf("%s: %s", err.Error(), values))
	}
	return value, nil
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

// AsInterface converts a FactValueList internal value to an interface{}.
func (v *FactValueList) AppendValue(value FactValue) {
	v.Value = append(v.Value, value)
}

// ParseStringToFactValue parses a string to a FactValue type.
func ParseStringToFactValue(str string) FactValue {
	if i, err := strconv.Atoi(str); err == nil {
		return &FactValueInt{Value: i}
	} else if b, err := strconv.ParseBool(str); err == nil {
		return &FactValueBool{Value: b}
	} else if f, err := strconv.ParseFloat(str, 64); err == nil {
		if math.IsInf(f, 0) {
			return &FactValueString{Value: str}
		}
		return &FactValueFloat{Value: f}
	}

	return &FactValueString{Value: str}
}

func getValue(fact FactValue, values []string) (FactValue, error) {
	if len(values) == 0 {
		return fact, nil
	}
	switch value := fact.(type) {
	case *FactValueMap:
		if child, found := value.Value[values[0]]; found {
			return getValue(child, values[1:])
		}
		return nil, fmt.Errorf("requested field value not found")

	case *FactValueList:
		listIndex, err := strconv.Atoi(values[0])
		if err != nil {
			return nil, fmt.Errorf("list index must be of integer value, %s provided", values[0])
		}
		if listIndex > len(value.Value)-1 {
			return nil, fmt.Errorf("%d index is not available in the list", listIndex)
		}
		return getValue(value.Value[listIndex], values[1:])
	default:
		return nil, fmt.Errorf("requested field value not found")
	}
}

func Prettify(fact FactValue) (string, error) {
	if fact == nil {
		return "()", nil
	}

	interfaceValue := fact.AsInterface()

	jsonResult, err := json.Marshal(interfaceValue)
	if err != nil {
		return "", errors.Wrap(err, "Error building the response")
	}

	var prettyfiedJSON bytes.Buffer
	if err := json.Indent(&prettyfiedJSON, jsonResult, "", "  "); err != nil {
		return "", errors.Wrap(err, "Error indenting the json data")
	}

	prettifiedJSONString := prettyfiedJSON.String()

	rhaiValue := strings.ReplaceAll(prettifiedJSONString, "{", "#{")
	rhaiValue = strings.ReplaceAll(rhaiValue, "null", "()")

	return rhaiValue, nil
}
