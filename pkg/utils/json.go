package utils

import (
	"bytes"
	"encoding/json"
)

func PrettyPrintJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// PrettyPrintToJSON receives an object of any type and returns a pretty-formatted
// output
func PrettyPrintToJSON(obj interface{}) (string, error) {
	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}
