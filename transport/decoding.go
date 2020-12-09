package transport

import (
	"encoding/json"
	"io"
	"reflect"
)

func Decode(r io.Reader, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

func DecodeFields(v interface{}) map[string]string {
	fields := map[string]string{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Tag.Get("json")
		fieldType := t.Field(i).Type.Name()
		fields[fieldName] = fieldType
	}
	return fields
}
