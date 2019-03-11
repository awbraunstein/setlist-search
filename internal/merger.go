package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/labstack/echo/v4"
)

// MergeJSONBody takes request params and injects them into the json body type.
// It is expected that the json body is either empty or top-level object.
//
// Currently only supports string values.
func MergeJSONBody(c echo.Context, v interface{}) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	// If there is a body, try to unmarshal the json.
	if len(body) > 0 {
		if err = json.Unmarshal(body, v); err != nil {
			return err
		}
	}
	// We know that v is a non-nil pointer because that is required by json.Unmarshal.
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("MergeJSONBody expects a non-nil pointer to a struct, but was: %v", reflect.TypeOf(v))
	}

	sType := rv.Elem().Type()
	for i := 0; i < sType.NumField(); i++ {
		sf := sType.Field(i)
		jsonField := sf.Tag.Get("json")
		if jsonField != "" {
			queryVal := c.QueryParam(jsonField)
			if queryVal != "" {
				rv.Elem().Field(i).SetString(queryVal)
			}
		}
	}
	return nil
}
