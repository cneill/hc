package hc

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

var (
	ErrNotStruct       = fmt.Errorf("provided input was not a struct")
	ErrInvalidURLValue = fmt.Errorf("invalid value in 'url' tag")
)

// URLValuesFromStruct takes any struct as 'input' and returns a url.Values object containing any URL variables specified based
// on the following struct tag rules:
//
//   - `url:"<name[,omitempty]|->"` -- specify the name of the url.Values key and optionally "omitempty", or "-" to ignore
//     the field entirely
//   - `url_val_sep:"<separator string>"` -- if the field is a slice, each value will be included in the specified url.Values
//     key as a single string joined by the separator; otherwise, error
//
// It will also parse values from embedded structs that follow the same rules.
func URLValuesFromStruct(input any) (url.Values, error) {
	results := url.Values{}
	inputType := reflect.TypeOf(input)
	inputValue := reflect.ValueOf(input)

	// if we have a pointer to a struct, set inputType and inputValue to the underlying struct
	if inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
		inputValue = inputValue.Elem()
	}

	if inputType.Kind() != reflect.Struct {
		return results, ErrNotStruct
	}

	// walk each struct field and check for 'url' tags
	for i := 0; i < inputValue.NumField(); i++ {
		fieldType := inputType.Field(i)
		fieldValue := inputValue.Field(i)

		// check embedded structs for variables to include
		if fieldType.IsExported() && fieldType.Type.Kind() == reflect.Struct && fieldType.Anonymous {
			raw := fieldValue.Interface()
			embeddedValues, err := URLValuesFromStruct(raw)
			if err != nil {
				return results, fmt.Errorf("failed to get embedded values from struct %q: %w", fieldType.Name, err)
			}

			for key, values := range embeddedValues {
				for _, value := range values {
					results.Add(key, value)
				}
			}

			continue
		}

		if !fieldType.IsExported() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		details, err := getURLTagDetails(fieldType, fieldValue)
		if err != nil {
			return results, fmt.Errorf("failed to parse url struct tags: %w", err)
		}

		if details.OmitAlways {
			continue
		} else if details.OmitEmpty && len(details.Values) == 0 {
			continue
		}

		for _, val := range details.Values {
			results.Add(details.Key, val)
		}
	}

	return results, nil
}

type urlTagDetails struct {
	Key        string
	Values     []string
	OmitEmpty  bool
	OmitAlways bool
}

func getURLTagDetails(sf reflect.StructField, val reflect.Value) (urlTagDetails, error) {
	result := urlTagDetails{
		Values: []string{},
	}

	// if no "url" tag is specified, don't do anything else
	urlTag, ok := sf.Tag.Lookup("url")
	if !ok {
		return result, nil
	}

	tagFields := strings.Split(urlTag, ",")
	// if we just have "-" or "<name>"
	if len(tagFields) == 1 {
		// if we got "-", we don't need to do anything else (unless I want to allow "-" key?)
		if urlTag == "-" {
			result.OmitAlways = true
			return result, nil
		}

		result.Key = urlTag
	} else {
		if tagFields[1] == "omitempty" {
			result.OmitEmpty = true
		} else {
			return result, fmt.Errorf("failed to parse 'url' tag %q: second field can only be 'omitempty'", urlTag)
		}

		result.Key = tagFields[0]
	}

	// check for a zero-valued reflect.Value or IsZero
	if result.OmitEmpty {
		if val == (reflect.Value{}) {
			return result, nil
		}

		if val.IsZero() {
			return result, nil
		}
	}

	stringValues, err := getStringValues(sf, val)
	if err != nil {
		return result, fmt.Errorf("failed to marshal values of field %q: %w", result.Key, err)
	}

	// getStringValues will return []string{""} since it doesn't know about omitempty - check that here
	if result.OmitEmpty && len(stringValues) > 0 {
		newVals := []string{}

		for _, strVal := range stringValues {
			if strVal != "" {
				newVals = append(newVals, strVal)
			}
		}

		stringValues = newVals
	}

	result.Values = stringValues

	return result, nil
}

func getStringValues(sf reflect.StructField, val reflect.Value) ([]string, error) {
	initialValues := []string{}

	kind := sf.Type.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			iVal := val.Index(i)
			initialValues = append(initialValues, iVal.String())
		}
	case reflect.String:
		initialValues = append(initialValues, val.String())
	case reflect.Bool:
		initialValues = append(initialValues, fmt.Sprintf("%t", val.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		initialValues = append(initialValues, fmt.Sprintf("%d", val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		initialValues = append(initialValues, fmt.Sprintf("%d", val.Uint()))
	case reflect.Float32, reflect.Float64:
		initialValues = append(initialValues, fmt.Sprintf("%.2f", val.Float()))
	default:
		invalid := true

		if val.CanInterface() {
			raw := val.Interface()

			if s, ok := raw.(fmt.Stringer); ok {
				initialValues = append(initialValues, s.String())
				invalid = false
			}
		}

		if invalid {
			return []string{}, fmt.Errorf("%q is an invalid type for a URL value", sf.Type)
		}
	}

	if separator := sf.Tag.Get("url_val_sep"); separator != "" {
		return []string{strings.Join(initialValues, separator)}, nil
	}

	return initialValues, nil
}
