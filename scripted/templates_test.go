package scripted

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func Test_StringifyJsonValues(t *testing.T) {
	jsonData := `
{
  "id": 1,
  "person": {
   "name": "John",
   "age": 30
  },
  "cars": [
    {"car1": "Ford"},
    {"car2": "BMW"},
    {"car3": "Fiat"}
  ]
}
`
	stringifiedData := `
{
  "id": "1",
  "person": {
   "name": "John",
   "age": "30"
  },
  "cars": [
    {"car1": "Ford"},
    {"car2": "BMW"},
    {"car3": "Fiat"}
  ]
}
`

	expected := fromJsonMust(stringifiedData).(map[string]interface{})
	input, _ := fromJson(jsonData)
	output := stringifyJsonValues(input).(map[string]interface{})
	equal := true
	keys := map[string]bool{}
	for key := range output {
		keys[key] = true
	}
	for key := range expected {
		keys[key] = true
	}
	for key := range keys {
		this := output[key]
		other := expected[key]
		if !reflect.DeepEqual(this, other) {
			os.Stderr.WriteString(fmt.Sprintf("[%#v] %#v != %#v\n", key, this, other))
			equal = false
		}
	}
	if Debug {
		os.Stderr.WriteString(fmt.Sprintf("input: %#v\n", input))
		os.Stderr.WriteString(fmt.Sprintf("output: %#v\n", output))
		os.Stderr.WriteString(fmt.Sprintf("expected: %#v\n", expected))
	}
	if !equal {
		os.Stderr.WriteString(fmt.Sprintf("output: %v\n", toPrettyJsonMust(output)))
		os.Stderr.WriteString(fmt.Sprintf("expected: %v\n", toPrettyJsonMust(output)))
		t.Fail()
	}
}
