package godefault

import (
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// GenerateCode will generate default filler implementation that avoid reflection
func GenerateCode(data interface{}, out io.Writer) error {
	if data == nil {
		return ErrNilValue
	}
	var dataValue reflect.Value
	if val, ok := data.(reflect.Value); ok {
		dataValue = val
	} else {
		dataValue = reflect.ValueOf(data)
	}

	switch dataValue.Kind() {
	case reflect.Interface, reflect.Ptr:
		return GenerateCode(dataValue.Elem(), out)
	default:
		dataType := dataValue.Type()
		generator := _NewCodeGenerator(dataType.Name(), out)
		return generator.GenerateCode(dataValue, dataType)
	}
}

type _CodeGenerator struct {
	writer      io.Writer
	indentation uint

	receiver string
	name     string
}

func _NewCodeGenerator(name string, writer io.Writer) *_CodeGenerator {
	return &_CodeGenerator{
		writer:   writer,
		receiver: strings.ToLower(string(name[0])),
		name:     name,
	}
}

func (c *_CodeGenerator) Writef(format string, datas ...interface{}) error {
	return c.WriteString(fmt.Sprintf(format, datas...))
}

func (c *_CodeGenerator) WriteString(data string) error {
	_, err := c.writer.Write([]byte(data))
	return err
}

func (c *_CodeGenerator) writeIdentation() error {
	for i := 0; i < int(c.indentation); i++ {
		if _, err := c.writer.Write([]byte{'\t'}); err != nil {
			return nil
		}
	}
	return nil
}

type beginBlockFunc func(c *_CodeGenerator) error

func (c *_CodeGenerator) BeginBlock(callback beginBlockFunc) error {
	c.indentation++
	if err := c.NewLine(); err != nil {
		return err
	}
	if err := callback(c); err != nil {
		return err
	}
	c.indentation--
	return c.NewLine()
}

func (c *_CodeGenerator) NewLine() error {
	if _, err := c.writer.Write([]byte{'\n'}); err != nil {
		return err
	}
	return c.writeIdentation()
}

func (c *_CodeGenerator) WriteSpace() error {
	_, err := c.writer.Write([]byte{' '})
	return err
}

func (c *_CodeGenerator) GenerateCode(dataValue reflect.Value, dataType reflect.Type) error {
	if err := c.Writef(`func (%s *%s) Default() {`, c.receiver, c.name); err != nil {
		return err
	}
	if err := c.BeginBlock(func(c *_CodeGenerator) (gerr error) {
		numField := dataValue.NumField()
		for i := 0; i < numField; i++ {
			fieldType := dataType.Field(i)
			// Check if field is exported
			if fieldName := fieldType.Name[0]; fieldName < 'A' || fieldName > 'Z' {
				continue
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
			fieldValue := dataValue.Field(i)
			if !fieldValue.CanSet() {
				return fmt.Errorf(`Tag is defined for field name "%s" but field is unsetable`, fieldType.Name)
			}
			if err := c._ExecuteFlag(fieldValue, fieldType, tagValue, optionSet); err != nil {
				return err
			}
			if i < numField-1 {
				if err := c.NewLine(); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return c.WriteString("}")
}

func (c *_CodeGenerator) _ExecuteField(value reflect.Value, structField reflect.StructField, tagValue string, optionSet _KeySet) error {
	if err := c.Writef(`%s.%s = `, c.receiver, structField.Name); err != nil {
		return err
	}
	return c._GenerateCodeValue(value, structField, tagValue, optionSet)
}

func (c *_CodeGenerator) _GenerateCodeValue(value reflect.Value, structField reflect.StructField, tagValue string, optionSet _KeySet) error {
	realValue := value.Interface()
	switch realValue.(type) {
	case time.Duration:
		duration, err := time.ParseDuration(tagValue)
		if err != nil {
			return err
		}
		return c.Writef(`time.Duration(%d)`, duration.Nanoseconds())
	case []byte:
		res, err := base64.StdEncoding.DecodeString(tagValue)
		if err != nil {
			return err
		}
		if err := c.WriteString("[]byte { "); err != nil {
			return err
		}
		if err := c.BeginBlock(func(c *_CodeGenerator) error {
			for i, data := range res {
				if i > 0 && i%10 == 0 {
					if err := c.NewLine(); err != nil {
						return err
					}
				}
				if err := c.Writef("0x%x, ", data); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return c.WriteString("}")
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		val, err := strconv.ParseInt(tagValue, 10, 0)
		if err != nil {
			return err
		}
		return c.Writef(`%d`, val)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		val, err := strconv.ParseUint(tagValue, 10, 0)
		if err != nil {
			return err
		}
		return c.Writef(`%d`, val)
	case reflect.String:
		return c.Writef(`"%s"`, tagValue)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(tagValue, 0)
		if err != nil {
			return err
		}
		return c.Writef(`%f`, val)
	case reflect.Array, reflect.Slice:
		return c._GenerateSliceValue(value, structField, tagValue, optionSet)
	default:
		return fmt.Errorf(`Unsupported Value Given for field name "%s"`, value.Type().Name())
	}
}

func (c *_CodeGenerator) _GenerateSliceValue(value reflect.Value, structField reflect.StructField, tagValue string, optionSet _KeySet) error {
	valueType := value.Type()
	splitted := strings.Split(tagValue, ",")
	newSlice := reflect.MakeSlice(valueType, 1, 1)
	dummyValue := newSlice.Index(0)
	if err := c.Writef(`[]%s {`, dummyValue.Type().Name()); err != nil {
		return err
	}
	if err := c.BeginBlock(func(c *_CodeGenerator) error {

		for i, split := range splitted {
			if err := c._GenerateCodeValue(dummyValue, structField, split, _KeySet{}); err != nil {
				return nil
			}
			if err := c.WriteString(", "); err != nil {
				return err
			}
			if i < len(splitted)-1 {
				if err := c.NewLine(); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return c.WriteString(`}`)

}

func (c *_CodeGenerator) _ExecuteFlag(value reflect.Value, structField reflect.StructField, tagValue string, optionSet _KeySet) error {
	if optionSet.HasKey("nonzero") {
		var zeroValStr string
		switch value.Kind() {
		case reflect.String:
			zeroValStr = `""`
		case reflect.Slice, reflect.Array:
			zeroValStr = `nil`
		default:
			zeroValStr = `0`
		}
		if err := c.Writef(`if %s.%s == %s {`, c.receiver, structField.Name, zeroValStr); err != nil {
			return err
		}
		if err := c.BeginBlock(func(c *_CodeGenerator) error {
			return c._ExecuteField(value, structField, tagValue, optionSet)
		}); err != nil {
			return err
		}
		return c.Writef(`}`)
	}
	return c._ExecuteField(value, structField, tagValue, optionSet)

}
