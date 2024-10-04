package compute

import "reflect"

func GetStructFields(s interface{}) []string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
			t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
			return nil
	}

	fields := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
			fields[i] = t.Field(i).Name
	}
	return fields
}