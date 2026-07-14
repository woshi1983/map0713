// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	yaml "gopkg.in/yaml.v2"
)

// EZTemplate provides robust template functions for rendering text templates.
func EZTemplate(tpl string) (*template.Template, error) {
	t := template.New("__main__")
	return t.Funcs(template.FuncMap{
		"uuidgen":    uuidgen,
		"now":        now,
		"timestamp":  timestamp,
		"quote":      strconv.Quote,
		"unquote":    unquote,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.ToTitle,
		"replace":    replace,
		"sha256":     toSha256,
		"trim":       trim,
		"trimPrefix": trimPrefix,
		"trimSuffix": trimSuffix,
		"trimSpace":  trimSpace,
		"indent":     indent,
		"split":      split,
		"join":       join,
		"first":      first,
		"last":       last,
		"contains":   contains,
		"atob":       atob,
		"btoa":       btoa,
		"json":       toJson,
		"parseJson":  parseJSON,
		"yaml":       toYAML,
		"parseYaml":  parseYAML,
		"fetch":      fetch,
		"import":     withImport(t),

		"bool":     toBool,
		"int":      toInt,
		"uint":     toUint,
		"float":    toFloat,
		"string":   toString,
		"time":     toTime,
		"duration": toDuration,
		"semver":   toSemver,

		"add":       add,
		"subtract":  subtract,
		"multiply":  multiply,
		"divide":    divide,
		"modulo":    modulo,
		"min":       minimum,
		"minimum":   minimum,
		"max":       maximum,
		"maximum":   maximum,
		"before":    before,
		"after":     after,
		"timeAdd":   timeAdd,
		"timeSub":   timeSub,
		"olderThan": olderThan,
		"newerThan": newerThan,
	}).Parse(tpl)
}

func uuidgen() string {
	return uuid.New().String()
}

func now() time.Time {
	return time.Now()
}

func timestamp(layout string, val any) (string, error) {
	v := reflect.ValueOf(val)
	var t time.Time

	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		t = time.Unix(v.Int(), 0)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		secs := v.Uint()
		if v.Uint() > math.MaxInt64 {
			secs = math.MaxInt64
		}
		t = time.Unix(int64(secs), 0)
	case reflect.Float32, reflect.Float64:
		secs := v.Float()
		if v.Uint() > math.MaxInt64 {
			secs = math.MaxInt64
		}
		t = time.Unix(int64(secs), 0)
	case reflect.String:
		tp, err := time.Parse(time.RFC3339Nano, v.String())
		if err != nil {
			return "", errors.Wrap(err, "timestamp")
		}
		t = tp
	default:
		var ok bool
		t, ok = val.(time.Time)
		if !ok {
			return "", fmt.Errorf("timestamp: type of %q (%T) is not convertable", v, val)
		}
	}

	if layout == "" {
		layout = time.RFC3339Nano
	}
	return t.Format(layout), nil
}

func unquote(s string) (string, error) {
	ret, err := strconv.Unquote(s)
	if err != nil {
		return "", errors.Wrap(err, "unquote")
	}
	return ret, nil
}

func replace(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

func toSha256(val any) (string, error) {
	v := reflect.ValueOf(val)
	h := sha256.New()

	switch v.Kind() {
	case reflect.Bool:
		h.Write([]byte(strconv.FormatBool(v.Bool())))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		h.Write([]byte(strconv.FormatInt(v.Int(), 10)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		h.Write([]byte(strconv.FormatUint(v.Uint(), 10)))
	case reflect.Float32, reflect.Float64:
		h.Write([]byte(strconv.FormatFloat(v.Float(), 'f', -1, 64)))
	case reflect.String:
		h.Write([]byte(v.String()))
	default:
		return "", fmt.Errorf("sha256: type of %q (%T) is not hashable", v, val)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func trim(cutset string, s string) (string, error) {
	return strings.Trim(s, cutset), nil
}

func trimPrefix(prefix string, s string) (string, error) {
	return strings.TrimPrefix(s, prefix), nil
}

func trimSuffix(suffix string, s string) (string, error) {
	return strings.TrimSuffix(s, suffix), nil
}

func trimSpace(s string) (string, error) {
	return strings.TrimSpace(s), nil
}

func indent(spaces int, s string) (string, error) {
	if spaces < 0 {
		return "", fmt.Errorf("indent: %d is not a positive integer", spaces)
	}

	var (
		output []byte
		size   int
	)
	sp := true
	prefix := []byte(strings.Repeat(" ", spaces))
	for _, c := range []byte(s) {
		if sp && c != '\n' {
			output = append(output, prefix...)
			size += spaces
		}
		output = append(output, c)
		sp = c == '\n'
		size++
	}
	return string(output[:size]), nil
}

func split(sep, s string) []string {
	return strings.Split(s, sep)
}

func join(sep string, a []string) (string, error) {
	return strings.Join(a, sep), nil
}

func first(val any) (reflect.Value, error) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Array:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("first: length of array must be greater than zero")
		}
		return v.Index(0), nil
	case reflect.Slice:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("first: length of array must be greater than zero")
		}
		return v.Index(0), nil
	case reflect.String:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("first: length of array must be greater than zero")
		}
		return v.Index(0), nil
	default:
		return reflect.Value{}, fmt.Errorf("first: type of %q (%T) cannot be indexed", v, val)
	}
}

func last(val any) (reflect.Value, error) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Array:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("last: length of array must be greater than zero")
		}
		return v.Index(v.Len() - 1), nil
	case reflect.Slice:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("last: length of array must be greater than zero")
		}
		return v.Index(v.Len() - 1), nil
	case reflect.String:
		if v.Len() < 1 {
			return reflect.Value{}, fmt.Errorf("last: length of array must be greater than zero")
		}
		return v.Index(v.Len() - 1), nil
	default:
		return reflect.Value{}, fmt.Errorf("last: type of %q (%T) cannot be indexed", v, val)
	}
}

func contains(slice, val any) (bool, error) {
	sv := reflect.ValueOf(slice)
	v := reflect.ValueOf(val)

	switch sv.Kind() {
	case reflect.Array, reflect.Slice:
		var intfs []any
		if reflect.TypeOf(slice).Elem().Kind() == reflect.Interface {
			intfs = slice.([]any)
		}

		for i := 0; i < sv.Len(); i++ {
			var vv reflect.Value
			if intfs != nil {
				vv = reflect.ValueOf(intfs[i])
			} else {
				vv = sv.Index(i)
			}

			switch vv.Kind() {
			case reflect.Bool:
				if v.Kind() == reflect.Bool && v.Bool() == vv.Bool() {
					return true, nil
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch v.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if v.Int() == vv.Int() {
						return true, nil
					}
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				switch v.Kind() {
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if v.Uint() == vv.Uint() {
						return true, nil
					}
				}
			case reflect.Float32, reflect.Float64:
				switch v.Kind() {
				case reflect.Float32, reflect.Float64:
					if v.Float() == vv.Float() {
						return true, nil
					}
				}
			case reflect.String:
				if v.Type() == vv.Type() && v.String() == vv.String() {
					return true, nil
				}
			}
		}
	case reflect.String:
		if v.Type() == sv.Type() && strings.Contains(sv.String(), v.String()) {
			return true, nil
		}
	}
	return false, nil
}

func atob(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func btoa(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.Wrap(err, "btoa")
	}
	return string(b), nil
}

func toBool(val any) (bool, error) {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > 0 {
			return true, nil
		}
		return false, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() > 0 {
			return true, nil
		}
		return false, nil
	case reflect.Float32, reflect.Float64:
		if v.Float() > 0 {
			return true, nil
		}
		return false, nil
	case reflect.String:
		ret, err := strconv.ParseBool(v.String())
		if err != nil {
			return ret, errors.Wrap(err, "bool")
		}
		return ret, nil
	default:
		return false, fmt.Errorf("bool: type of %q (%T) is not convertable", v, val)
	}
}

func toInt(val any) (int64, error) {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret := v.Uint()
		if ret > math.MaxInt64 {
			return math.MaxInt64, nil
		}
		return int64(ret), nil
	case reflect.Float32, reflect.Float64:
		ret := v.Float()
		if ret > math.MaxInt64 {
			return math.MaxInt64, nil
		}
		return int64(ret), nil
	case reflect.String:
		ret, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "int")
		}
		return ret, nil
	default:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return val.(time.Time).Unix(), nil
		}
		return 0, fmt.Errorf("int: type of %q (%T) is not convertable", v, val)
	}
}

func toUint(val any) (uint64, error) {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < 0 {
			return 0, nil
		}
		return uint64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Float32, reflect.Float64:
		ret := v.Float()
		if ret > math.MaxUint64 {
			return math.MaxUint64, nil
		}
		return uint64(ret), nil
	case reflect.String:
		ret, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "uint")
		}
		return ret, nil
	default:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return uint64(val.(time.Time).Unix()), nil
		}
		return 0, fmt.Errorf("uint: type of %q (%T) is not convertable", v, val)
	}
}

func toFloat(val any) (float64, error) {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		ret, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, errors.Wrap(err, "float")
		}
		return ret, nil
	default:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return float64(val.(time.Time).Unix()), nil
		}
		return 0, fmt.Errorf("float: type of %q (%T) is not convertable", v, val)
	}
}

func toString(val any) (string, error) {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.String:
		return v.String(), nil
	default:
		return "", fmt.Errorf("string: type of %q (%T) is not convertable", v, val)
	}
}

var nullTime = time.Time{}

func toTime(val any) (time.Time, error) {
	str, err := toString(val)
	if err != nil {
		return nullTime, errors.Wrapf(err, "time")
	}

	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return nullTime, errors.Wrapf(err, "time")
	}
	return t, nil
}

func toDuration(val any) (float64, error) {
	str, err := toString(val)
	if err != nil {
		return 0, errors.Wrapf(err, "duration")
	}

	d, err := time.ParseDuration(str)
	if err != nil {
		return 0, errors.Wrapf(err, "duration")
	}
	return d.Seconds(), nil
}

func toSemver(val any) (*semver.Version, error) {
	str, err := toString(val)
	if err != nil {
		return nil, errors.Wrapf(err, "semver")
	}

	v, err := semver.NewVersion(str)
	if err != nil {
		return nil, errors.Wrapf(err, "semver")
	}
	return v, nil
}

func toJson(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseJSON(s string) (any, error) {
	if s == "" {
		return map[string]any{}, nil
	}

	var data any
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func toYAML(v any) (string, error) {
	result, err := yaml.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "yaml")
	}
	return string(bytes.TrimSpace(result)), nil
}

func parseYAML(s string) (any, error) {
	if s == "" {
		return map[string]any{}, nil
	}

	var data any
	if err := yaml.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func add(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Int()) + fb, nil
		default:
			return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Uint()) + fb, nil
		default:
			return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
			}
			return av.Float() + fb, nil
		default:
			return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("add: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fa + float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fa + float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return fa + bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
			}
			return fa + fb, nil
		default:
			return nil, fmt.Errorf("add: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("add: type of %q (%T) is not calculable", av, a)
	}
}

func subtract(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() - int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Int()) - fb, nil
		default:
			return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) - bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Uint()) - fb, nil
		default:
			return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() - float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
			}
			return av.Float() - fb, nil
		default:
			return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fa - float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fa - float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return fa - bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
			}
			return fa - fb, nil
		default:
			return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("subtract: type of %q (%T) is not calculable", av, a)
	}
}

func multiply(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() * int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Int()) * fb, nil
		default:
			return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) * bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Uint()) * fb, nil
		default:
			return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() * float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
			}
			return av.Float() * fb, nil
		default:
			return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fa * float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fa * float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return fa * bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
			}
			return fa * fb, nil
		default:
			return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("multiply: type of %q (%T) is not calculable", av, a)
	}
}

func divide(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() / int64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Int()) / fb, nil
		default:
			return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) / bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
			}
			return float64(av.Uint()) / fb, nil
		default:
			return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() / float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() / float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() / bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
			}
			return av.Float() / fb, nil
		default:
			return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fa / float64(bv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fa / float64(bv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return fa / bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
			}
			return fa / fb, nil
		default:
			return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("divide: type of %q (%T) is not calculable", av, a)
	}
}

// modulo returns the modulo of b from a.
func modulo(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() % int64(bv.Uint()), nil
		case reflect.String:
			ib, err := strconv.ParseInt(bv.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
			}
			return av.Int() % ib, nil
		default:
			return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() % bv.Uint(), nil
		case reflect.String:
			ib, err := strconv.ParseInt(bv.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
			}
			return int64(av.Uint()) % ib, nil
		default:
			return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		ia, err := strconv.ParseInt(av.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return ia % bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return ia % int64(bv.Uint()), nil
		case reflect.String:
			ib, err := strconv.ParseInt(bv.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
			}
			return ia % ib, nil
		default:
			return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("modulo: type of %q (%T) is not calculable", av, a)
	}
}

// minimum returns the minimum between a and b.
func minimum(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Int() < bv.Int() {
				return av.Int(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Int() < int64(bv.Uint()) {
				return av.Int(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if float64(av.Int()) < bv.Float() {
				return av.Int(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
			}
			if float64(av.Int()) < fb {
				return av.Int(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if int64(av.Uint()) < bv.Int() {
				return av.Uint(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Uint() < bv.Uint() {
				return av.Uint(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if float64(av.Uint()) < bv.Float() {
				return av.Uint(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
			}
			if float64(av.Uint()) < fb {
				return av.Uint(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Float() < float64(bv.Int()) {
				return av.Float(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Float() < float64(bv.Uint()) {
				return av.Float(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if av.Float() < bv.Float() {
				return av.Float(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
			}
			if av.Float() < fb {
				return av.Float(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fa < float64(bv.Int()) {
				return fa, nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fa < float64(bv.Uint()) {
				return fa, nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if fa < bv.Float() {
				return fa, nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
			}
			if fa < fb {
				return fa, nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("minimum: type of %q (%T) is not calculable", av, a)
	}
}

// maximum returns the maximum between a and b.
func maximum(b, a any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Int() > bv.Int() {
				return av.Int(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Int() > int64(bv.Uint()) {
				return av.Int(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if float64(av.Int()) > bv.Float() {
				return av.Int(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
			}
			if float64(av.Int()) > fb {
				return av.Int(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if int64(av.Uint()) > bv.Int() {
				return av.Uint(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Uint() > bv.Uint() {
				return av.Uint(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if float64(av.Uint()) > bv.Float() {
				return av.Uint(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
			}
			if float64(av.Uint()) > fb {
				return av.Uint(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Float() > float64(bv.Int()) {
				return av.Float(), nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if av.Float() > float64(bv.Uint()) {
				return av.Float(), nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if av.Float() > bv.Float() {
				return av.Float(), nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
			}
			if av.Float() > fb {
				return av.Float(), nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
		}
	case reflect.String:
		fa, err := strconv.ParseFloat(av.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", av, a)
		}
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fa > float64(bv.Int()) {
				return fa, nil
			}
			return bv.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fa > float64(bv.Uint()) {
				return fa, nil
			}
			return bv.Uint(), nil
		case reflect.Float32, reflect.Float64:
			if fa > bv.Float() {
				return fa, nil
			}
			return bv.Float(), nil
		case reflect.String:
			fb, err := strconv.ParseFloat(bv.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
			}
			if fa > fb {
				return fa, nil
			}
			return fb, nil
		default:
			return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", bv, b)
		}
	default:
		return nil, fmt.Errorf("maximum: type of %q (%T) is not calculable", av, a)
	}
}

func before(t2, t1 time.Time) (bool, error) {
	return t1.Before(t2), nil
}

func after(t2, t1 time.Time) (bool, error) {
	return t1.After(t2), nil
}

func timeAdd(dur any, t time.Time) (time.Time, error) {
	d := reflect.ValueOf(dur)

	switch d.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		duration := time.Duration(d.Int()) * time.Second
		return t.Add(duration), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		duration := time.Duration(d.Uint()) * time.Second
		return t.Add(duration), nil
	case reflect.Float32, reflect.Float64:
		duration := time.Duration(d.Float()) * time.Second
		return t.Add(duration), nil
	case reflect.String:
		duration, err := time.ParseDuration(d.String())
		if err != nil {
			return nullTime, errors.Wrap(err, "timeAdd")
		}
		return t.Add(duration), nil
	default:
		return nullTime, fmt.Errorf("timeAdd: type of %q (%T) is not convertable", d, dur)
	}
}

func timeSub(dur any, t time.Time) (time.Time, error) {
	d := reflect.ValueOf(dur)

	switch d.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		duration := time.Duration(d.Int()) * time.Second
		return t.Add(-duration), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		duration := time.Duration(d.Uint()) * time.Second
		return t.Add(-duration), nil
	case reflect.Float32, reflect.Float64:
		duration := time.Duration(d.Float()) * time.Second
		return t.Add(-duration), nil
	case reflect.String:
		duration, err := time.ParseDuration(d.String())
		if err != nil {
			return nullTime, errors.Wrap(err, "timeSub")
		}
		return t.Add(-duration), nil
	default:
		return nullTime, fmt.Errorf("timeSub: type of %q (%T) is not convertable", d, dur)
	}
}

func olderThan(v2, v1 *semver.Version) (bool, error) {
	return v1.LessThan(v2), nil
}

func newerThan(v2, v1 *semver.Version) (bool, error) {
	return v1.GreaterThan(v2), nil
}

func fetch(uri string) (string, error) {
	ch := make(chan string, 1)
	cli := httputil.NewClient()
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrapf(err, "load")
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return "", fmt.Errorf("load: %s should begin with \"http://\" or \"https://\"", uri)
	}

	if err := util.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
		defer cancel()

		do, e := cli.New(ctx, u.String())
		if e != nil {
			return e
		}
		return do(func(resp *http.Response) error {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("%s %s: %d %v", resp.Request.Method, resp.Request.URL, resp.StatusCode, err)
			}
			if resp.StatusCode >= http.StatusBadRequest {
				if len(b) > 0 {
					return fmt.Errorf("%s %s: %d %s", resp.Request.Method, resp.Request.URL, resp.StatusCode, string(b))
				}
				return fmt.Errorf("%s %s: %d %s", resp.Request.Method, resp.Request.URL, resp.StatusCode, http.StatusText(resp.StatusCode))
			}
			ch <- string(b)
			return nil
		})
	}, 3, 500*time.Millisecond); err != nil {
		return "", errors.Wrap(err, "load")
	}

	return <-ch, nil
}

func withImport(t *template.Template) func(tplName, s string) (string, error) {
	return func(tplName, s string) (string, error) {
		_, err := t.New(tplName).Parse(s)
		if err != nil {
			return "", errors.Wrap(err, "import")
		}
		return "", nil
	}
}
