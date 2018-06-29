package scripted

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func terraformify(input interface{}) map[string]interface{} {
	ret := map[string]interface{}{}
	var inner func(string, interface{})
	addPrefix := func(prefix string, value interface{}) string {
		add := fmt.Sprintf("%v", value)
		if prefix == "" {
			return add
		} else {
			return fmt.Sprintf("%s.%s", prefix, add)
		}
	}
	inner = func(prefix string, cur interface{}) {
		curValue := reflect.ValueOf(cur)
		kind := curValue.Kind()
		if kind == reflect.Map {
			for _, k := range curValue.MapKeys() {
				v := curValue.MapIndex(k).Interface()
				inner(addPrefix(prefix, k), v)
			}
			inner(addPrefix(prefix, "%"), curValue.Len())
		} else if kind == reflect.Slice {
			list := cur.([]interface{})
			for i, v := range list {
				inner(addPrefix(prefix, i), v)
			}
			inner(addPrefix(prefix, "#"), len(list))
		} else {
			value := fmt.Sprintf("%v", cur)
			ret[prefix] = value
		}
	}
	inner("", input)
	return ret
}

func deterraformify(input interface{}) interface{} {
	dict, ok := input.(map[string]interface{})

	if !ok {
		return input
	}

	dicts := map[string]interface{}{}
	for key, value := range dict {
		cur := dicts
		path := strings.Split(key, ".")
		for _, piece := range path[:len(path)-1] {
			nxt, ok := cur[piece]
			if !ok {
				nxt = map[string]interface{}{}
				cur[piece] = nxt
			}
			cur = nxt.(map[string]interface{})
		}
		cur[path[len(path)-1]] = value
	}
	var arrayify func(interface{}) interface{}
	arrayify = func(cur interface{}) interface{} {
		dict, ok := cur.(map[string]interface{})
		if !ok {
			return cur
		}
		indices := map[int]string{}
		length := -1
		isArray := true
		for key, value := range dict {
			if key == "#" {
				length, _ = strconv.Atoi(value.(string))
				delete(dict, key)
				continue
			}
			if key == "%" {
				delete(dict, key)
				continue
			}
			dict[key] = arrayify(value)
			if i, err := strconv.Atoi(key); err == nil {
				indices[i] = key
			} else {
				isArray = false
			}
		}
		if length < 0 {
			isArray = false
		}
		if isArray {
			for i := 0; i < length; i++ {
				if _, ok := indices[i]; !ok {
					isArray = false
					break
				}
			}
		}
		if !isArray {
			return dict
		}

		var ret []interface{}
		for i := 0; i < length; i++ {
			key := strconv.Itoa(i)
			if val, ok := dict[key]; ok {
				ret = append(ret, val)
			} else {
				ret = append(ret, nil)
			}
		}
		return ret
	}
	return arrayify(dicts)
}
