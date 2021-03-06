package godefault

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	RECURSIVE = 1
)

// Fill will fill given struct with it's default value
func Fill(data interface{}, flag uint) error {
	if data == nil {
		return ErrNilValue
	}
	var dataValue reflect.Value
	if val, ok := data.(reflect.Value); ok {
		dataValue = val
	} else {
		dataValue = reflect.ValueOf(data)
	}

	indirectData := _FollowIndirection(dataValue)
	if indirectData == nil {
		return nil
	}
	dataValue = *indirectData

	if dataValue.Kind() != reflect.Struct {
		return ErrUnsupportedValue
	}
	return _Fill(dataValue, dataValue.Type(), flag)
}

func _Fill(dataValue reflect.Value, dataType reflect.Type, flag uint) error {
	numField := dataValue.NumField()
	for i := 0; i < numField; i++ {

		fieldType := dataType.Field(i)
		// Check if field is exported
		if fieldName := fieldType.Name[0]; fieldName < 'A' || fieldName > 'Z' {
			continue
		}

		fieldValue := dataValue.Field(i)
		if (flag & RECURSIVE) > 0 {
			if err := Fill(fieldValue, flag); err != nil && err != ErrUnsupportedValue {
				return err
			}
		}

		tagValue, hasTag := fieldType.Tag.Lookup("default")
		if !hasTag {
			continue
		}

		optionSet := _KeySet{}
		optionTagValue, hasOption := fieldType.Tag.Lookup("default-opt")
		if hasOption {
			options := strings.Split(optionTagValue, ",")
			for _, option := range options {
				optionSet.SetKey(option)
			}
		}
		if !fieldValue.CanSet() {
			return fmt.Errorf(`Tag is defined for field name "%s" but field is unsetable`, fieldType.Name)
		}
		if err := _FillValue(fieldValue, fieldType, tagValue, optionSet); err != nil {
			return err
		}
	}
	return nil
}

func _FillValue(value reflect.Value, fieldType reflect.StructField, tagValue string, optionSet _KeySet) error {
	if !value.IsZero() {
		return nil
	}

	realValue := value.Interface()
	switch realValue.(type) {
	case time.Duration:
		duration, err := time.ParseDuration(tagValue)
		if err != nil {
			return err
		}
		value.SetInt(int64(duration))
		return nil
	case []byte:
		res, err := base64.StdEncoding.DecodeString(tagValue)
		if err != nil {
			return err
		}
		value.SetBytes(res)
		return nil
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		val, err := strconv.ParseInt(tagValue, 10, 0)
		if err != nil {
			return err
		}
		value.SetInt(val)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		val, err := strconv.ParseUint(tagValue, 10, 0)
		if err != nil {
			return err
		}
		value.SetUint(val)
	case reflect.String:
		value.SetString(tagValue)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(tagValue, 0)
		if err != nil {
			return err
		}
		value.SetFloat(val)
	case reflect.Array, reflect.Slice:
		if err := _FillSlice(value, fieldType, tagValue); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`Unsupported Value Given for field name "%s"`, fieldType.Name)
	}
	return nil
}

func _FillSlice(value reflect.Value, fieldType reflect.StructField, tagValue string) error {
	valueType := value.Type()
	splitted := strings.Split(tagValue, ",")
	splittedLen := len(splitted)
	newSlice := reflect.MakeSlice(valueType, splittedLen, splittedLen)
	for i, split := range splitted {
		currentValue := newSlice.Index(i)
		if err := _FillValue(currentValue, fieldType, split, _KeySet{}); err != nil {
			return nil
		}
	}
	value.Set(newSlice)
	return nil

}

func _FollowIndirection(dataValue reflect.Value) *reflect.Value {
	switch dataValue.Kind() {
	case reflect.Interface, reflect.Ptr:
		if dataValue.IsNil() {
			return nil
		}
		return _FollowIndirection(dataValue.Elem())
	default:
		return &dataValue
	}
}
