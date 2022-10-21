package utils

import (
	"bytes"
	"encoding/json"
)

func ObjectToMapStringString(content interface{}) (map[string]string, error) {
	var name map[string]string
	marshalContent, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(marshalContent))
	d.UseNumber()
	if err := d.Decode(&name); err != nil {
		return nil, err
	}
	return name, nil
}
