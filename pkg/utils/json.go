package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func PrettyPrintJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// ExtractValueFromJSON returns the value of a key from a json string. Assumes there will be only one matching key.
func ExtractValueFromJSON(body string, key string) (string, error) {
	keystr := "\"" + key + "\":[^,;\\]}]*"
	r, _ := regexp.Compile(keystr)
	match := r.FindString(body)
	keyValMatch := strings.Split(match, ":")
	if len(keyValMatch) != 2 {
		return "", fmt.Errorf("unable to extract value of %s from json: %s", key, body)
	}
	val := strings.ReplaceAll(keyValMatch[1], "\"", "")
	return val, nil
}
