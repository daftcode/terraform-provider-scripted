package scripted

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func mergeMaps(maps ...map[string]string) map[string]string {
	ctx := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			ctx[k] = v
		}
	}
	return ctx
}

func castConfigList(v interface{}) []string {
	var ret []string
	for _, v := range v.([]interface{}) {
		ret = append(ret, v.(string))
	}
	return ret
}

func castConfigMap(v interface{}) map[string]string {
	ret := map[string]string{}
	if v == nil {
		return ret
	}
	valueMap, ok := v.(map[string]interface{})
	if !ok || valueMap == nil {
		return ret
	}
	for k, v := range valueMap {
		ret[k] = v.(string)
	}
	return ret
}

func castConfigChangeMap(o, n interface{}) *ChangeMap {
	return &ChangeMap{
		Old: castConfigMap(o),
		New: castConfigMap(n),
	}
}

func mapToEnv(env map[string]string) []string {
	var ret []string
	for key, value := range env {
		ret = append(ret, fmt.Sprintf("%s=%s", key, value))
	}
	return ret
}

func is(b, other interface{}) bool {
	x := reflect.ValueOf(b)
	y := reflect.ValueOf(other)
	return x.Pointer() == y.Pointer()
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func getMapHash(data map[string]interface{}) []string {
	var keys []string
	var entries []string

	ctx := data
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		entries = append(entries, hash(hash(k)+hash(ctx[k].(string))))
	}
	return entries
}

func isSet(str string) bool {
	return str != EmptyString
}

func isFilled(str string) bool {
	return str != EmptyString && str != ""
}

func chToString(lines chan string) chan string {
	output := make(chan string)
	go func() {
		var builder strings.Builder
		first := true
		for line := range lines {
			if !first {
				builder.WriteString("\n")
				first = false
			}
			builder.WriteString(line)
		}
		output <- builder.String()
		close(output)
	}()
	return output
}

func ToString(val interface{}) string {
	return fmt.Sprintf("%s", val)
}
