package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// PlainUnix renders data as classic unix-style textual output for an
// explicit --output plain invocation (see PlainOrUnix): a bare scalar
// prints as-is; a list of scalars (e.g. []string) prints one per line; a
// single multi-field object prints as sorted key=value lines (env-style);
// and a list of multi-field objects prints one record per line as its
// sorted field values joined by ':' (/etc/passwd-style).
func PlainUnix(cmd *cobra.Command, data any) error {
	v := reflect.ValueOf(data)
	for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			cmd.Println()
			return nil
		}
		v = v.Elem()
	}

	if !v.IsValid() {
		cmd.Println()
		return nil
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return plainUnixList(cmd, v)
	case reflect.Map, reflect.Struct:
		return plainUnixObject(cmd, data)
	default:
		cmd.Println(data)
		return nil
	}
}

func plainUnixList(cmd *cobra.Command, v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		if isPlainUnixScalar(item) {
			cmd.Println(item)
			continue
		}
		fields, err := flattenToSortedFields(item)
		if err != nil {
			return err
		}
		values := make([]string, len(fields))
		for i, f := range fields {
			values[i] = f.value
		}
		cmd.Println(strings.Join(values, ":"))
	}
	return nil
}

func plainUnixObject(cmd *cobra.Command, data any) error {
	fields, err := flattenToSortedFields(data)
	if err != nil {
		return err
	}
	for _, f := range fields {
		cmd.Printf("%s=%s\n", f.key, f.value)
	}
	return nil
}

func isPlainUnixScalar(v any) bool {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.Struct, reflect.Map:
		return false
	default:
		return true
	}
}

type plainField struct {
	key   string
	value string
}

// flattenToSortedFields round-trips data through JSON to normalize it into
// a flat map, then returns its fields sorted by key for deterministic
// output — the source struct/map field order isn't otherwise stable.
func flattenToSortedFields(data any) ([]plainField, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fields := make([]plainField, len(keys))
	for i, k := range keys {
		fields[i] = plainField{key: k, value: stringifyPlainValue(m[k])}
	}
	return fields, nil
}

func stringifyPlainValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	}
}
