package utils

import (
	"bytes"
	"encoding/json"
	"strings"
)

func PrettyPrintJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// PrettyPrintToJSON receives an object of any type and returns a pretty-formatted
// output
func PrettyPrintToJSON(obj interface{}) (string, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(obj)
	if err != nil {
		return "", err
	}

	res := string(buf.Bytes())

	// Remove the extra newline added when using a plain Encoder
	res = strings.TrimSuffix(res, "\n")

	return res, nil
}
