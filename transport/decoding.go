package transport

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/chill-and-code/event-bus/models"
)

func Decode(r io.Reader, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

func DecodeFields(v interface{}) []models.RequiredField {
	fields := make([]models.RequiredField, 0)
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		jsonTag := strings.Split(t.Field(i).Tag.Get("json"), ",")
		typeTag := strings.Split(t.Field(i).Tag.Get("type"), ",")
		isRequired := true
		if len(jsonTag) == 2 && jsonTag[1] == "omitempty" {
			isRequired = false
		}
		field := models.RequiredField{
			Name: jsonTag[0],
			Type: typeTag[0],
			Required: isRequired,
		}
		fields = append(fields, field)
	}
	return fields
}
