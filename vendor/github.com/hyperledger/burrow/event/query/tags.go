package query

import (
	"reflect"
	"strings"
)

const defaultMaxTagDepth = 10

type Tagged interface {
	Get(key string) (value interface{}, ok bool)
}

type TagMap map[string]interface{}

func (ts TagMap) Get(key string) (value interface{}, ok bool) {
	var vint interface{}
	vint, ok = ts[key]
	if !ok {
		return "", false
	}
	return vint, true
}

type CombinedTags []interface{}

func TagsFor(vs ...interface{}) CombinedTags {
	return vs
}

func (ct CombinedTags) Get(key string) (interface{}, bool) {
	for _, t := range ct {
		tagged, ok := t.(Tagged)
		if ok {
			v, ok := tagged.Get(key)
			if ok {
				return v, true
			}
		}
		v, ok := GetReflect(reflect.ValueOf(t), key)
		if ok {
			return v, true
		}
	}
	return nil, false
}

func GetReflect(rv reflect.Value, key string) (interface{}, bool) {
	return GetReflectDepth(rv, key, defaultMaxTagDepth)
}

var zeroValue = reflect.Value{}

// Pull out values in a nested struct by following path
func GetReflectDepth(rv reflect.Value, key string, maxDepth int) (interface{}, bool) {
	if maxDepth < 0 {
		return nil, false
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	keys := strings.SplitN(key, ".", 2)
	field := rv.FieldByName(keys[0])
	if field == zeroValue {
		return nil, false
	}
	// If there there are unconsumed segments in the keys then descend
	if len(keys) > 1 && (field.Kind() == reflect.Struct ||
		field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
		return GetReflectDepth(field, keys[1], maxDepth-1)

	}
	return field.Interface(), true
}

type taggedPrefix struct {
	tagged Tagged
	prefix string
}

func TaggedPrefix(prefix string, tagged Tagged) *taggedPrefix {
	return &taggedPrefix{
		prefix: prefix,
		tagged: tagged,
	}
}

func (t *taggedPrefix) Get(key string) (value interface{}, ok bool) {
	return t.tagged.Get(strings.TrimPrefix(key, t.prefix))
}
