package scripted

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	ctx := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			ctx[k] = v
		}
	}
	return ctx
}

func castConfigListString(v interface{}) []string {
	var ret []string
	for _, v := range v.([]interface{}) {
		ret = append(ret, fmt.Sprintf("%v", v))
	}
	return ret
}

func castConfigMap(v interface{}) map[string]interface{} {
	ret := map[string]interface{}{}
	if v == nil {
		return ret
	}
	valueMap, ok := v.(map[string]interface{})
	if !ok || valueMap == nil {
		return ret
	}
	return valueMap
}

func castConfigChangeMap(o, n interface{}) *ChangeMap {
	return &ChangeMap{
		Old: castConfigMap(o),
		New: castConfigMap(n),
	}
}

func castEnvironmentMap(v interface{}) map[string]string {
	ret := map[string]string{}
	if v == nil {
		return ret
	}
	valueMap, ok := v.(map[string]interface{})
	if !ok || valueMap == nil {
		return ret
	}
	for k, v := range valueMap {
		ret[k] = fmt.Sprintf("%s", v)
	}
	return ret
}

func castEnvironmentChangeMap(o, n interface{}) *EnvironmentChangeMap {
	return &EnvironmentChangeMap{
		Old: castEnvironmentMap(o),
		New: castEnvironmentMap(n),
	}
}

func mapToEnv(env map[string]string) []string {
	var ret []string
	for key, value := range env {
		ret = append(ret, fmt.Sprintf("%s=%s", key, fmt.Sprintf("%s", value)))
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
		v := fmt.Sprintf("%#v", ctx[k])
		entries = append(entries, hash(hash(k)+hash(v)))
	}
	return entries
}

func isSet(val interface{}) bool {
	if str, ok := val.(string); ok {
		return str != EnvEmptyString
	}
	return val != nil
}

func isFilled(val interface{}) bool {
	if str, ok := val.(string); ok {
		return str != EnvEmptyString && str != ""
	}
	return val != nil
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

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
