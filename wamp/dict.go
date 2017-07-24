package wamp

import (
	"errors"
	"reflect"
	"strings"
)

// NormalizeDict takes a dict and creates a new normalized dict where all
// map[string]xxx are converted to map[string]interface{}.  Values that cannot
// be converted, or are already the correct map type, remain the same.
//
// This is used for initial conversion of hello details.  The original dict is
// not mutated.
func NormalizeDict(v interface{}) map[string]interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Map {
		return nil
	}
	dict := map[string]interface{}{}
	for _, key := range val.MapKeys() {
		if key.Kind() == reflect.Interface {
			key = key.Elem()
		}
		if key.Kind() != reflect.String {
			continue
		}
		cv := val.MapIndex(key)
		newVal := NormalizeDict(cv.Interface())
		if newVal == nil {
			dict[key.String()] = cv.Interface()
			continue
		}
		dict[key.String()] = newVal
	}
	return dict
}

// Return the child dictionary for the given key, or nil if not present.
//
// If the child is not a map[string]interface{}, an attempt is made to convert
// it.  The dict is not modified since features may be looked up cuncurrently
// for the same session.
func DictChild(dict map[string]interface{}, key string) map[string]interface{} {
	iface, ok := dict[key]
	if !ok || iface == nil {
		// Map does not have the specified key or value is nil.
		return nil
	}
	child, ok := iface.(map[string]interface{})
	if !ok {
		// value is not in expected form; try to convert
		// Session details are normalized whensession is attached, so this
		// should not be necessary normally.
		child = NormalizeDict(iface)
		if child == nil {
			// could not convert
			return nil
		}
	}
	return child
}

// DictValue returns the value specified by the slice of path elements.
//
// To specify the path using a dot-separated string, call like this:
//     DictValue(dict, strings.Split(path, "."))
//
// For example, the path []string{"roles","callee","features","call_timeout"}
// returns  the value of the call_timeout feature as interface{}.  An error
// is returned if the value is not present.
func DictValue(dict map[string]interface{}, path []string) (interface{}, error) {
	for i := range path[:len(path)-1] {
		dict = DictChild(dict, path[i])
		if dict == nil {
			return nil, errors.New(
				"cannot find: " + strings.Join(path[:i+1], "."))
		}
	}
	v, ok := dict[path[len(path)-1]]
	if !ok {
		return nil, errors.New("cannot find: " + strings.Join(path, "."))
	}
	return v, nil
}

// DictFlag returns the bool specified by the dot-separated path.
//
// To specify the path using a dot-separated string, call like this:
//     DictFlag(dict, strings.Split(path, "."))
//
// For example: "roles.subscriber.features.publisher_identification" returns
// the value of the publisher_identification feature.  An error is returned if
// the value is not present or is not a boolean type.
func DictFlag(dict map[string]interface{}, path []string) (bool, error) {
	v, err := DictValue(dict, path)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, errors.New(
			strings.Join(path, ".") + " is not a boolean type")
	}
	return b, nil
}

// OptionString returns the string value of the option with the specified name.
// If the option is not present, an empty string is returned.
func OptionString(opts map[string]interface{}, optionName string) string {
	var opt string
	if _opt, ok := opts[optionName]; ok && _opt != nil {
		opt, _ = _opt.(string)
	}
	return opt
}

// OptionInt64 returns the int64 value of the option with the specified name.
// If the option is not present, a value of 0 is returned.
func OptionInt64(opts map[string]interface{}, optionName string) int64 {
	if opt, ok := opts[optionName]; ok && opt != nil {
		switch opt := opt.(type) {
		case int:
			return int64(opt)
		case int32:
			return int64(opt)
		case float64:
			return int64(opt)
		case uint:
			return int64(opt)
		case uint32:
			return int64(opt)
		case uint64:
			return int64(opt)
		}
	}
	return 0
}

// OptionString returns the boolean value of the option with the specified
// name.  If the option is not present, false is returned.
func OptionFlag(opts map[string]interface{}, optionName string) bool {
	var opt bool
	if _opt, ok := opts[optionName]; ok && _opt != nil {
		opt, _ = _opt.(bool)
	}
	return opt
}

// SetOption sets a single option name-value pair in message options dict.
func SetOption(dict map[string]interface{}, name string, value interface{}) map[string]interface{} {
	if dict == nil {
		dict = map[string]interface{}{}
	}
	dict[name] = value
	return dict
}