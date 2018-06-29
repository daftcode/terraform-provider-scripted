package scripted

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestTerraformify(t *testing.T) {
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
	expected := map[string]interface{}{
		"%":           "3",
		"id":          "1",
		"person.%":    "2",
		"person.age":  "30",
		"person.name": "John",
		"cars.#":      "3",
		"cars.0.%":    "1",
		"cars.0.car1": "Ford",
		"cars.1.%":    "1",
		"cars.1.car2": "BMW",
		"cars.2.%":    "1",
		"cars.2.car3": "Fiat",
	}
	input, _ := fromJson(jsonData)
	output := terraformify(input)
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
		if this != other {
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

	bw := deterraformify(output)
	backwards := bw.(map[string]interface{})

	if Debug {
		os.Stderr.WriteString(fmt.Sprintf("backwards: %#v\n", backwards))
	}

	inputMap := input.(map[string]interface{})
	inputMap["id"] = fmt.Sprintf("%v", inputMap["id"])
	person := inputMap["person"].(map[string]interface{})
	person["age"] = fmt.Sprintf("%v", person["age"])

	keys = map[string]bool{}
	for key := range backwards {
		keys[key] = true
	}
	for key := range inputMap {
		keys[key] = true
	}
	for key := range keys {
		this := backwards[key]
		other := inputMap[key]
		if !reflect.DeepEqual(this, other) {
			t, _ := toJson(this)
			o, _ := toJson(other)
			os.Stderr.WriteString(fmt.Sprintf("%#v %v != %v\n", key, t, o))
			equal = false
		}
	}
	if !equal {
		os.Stderr.WriteString(fmt.Sprintf("output: %v\n", toPrettyJsonMust(backwards)))
		os.Stderr.WriteString(fmt.Sprintf("expected: %v\n", toPrettyJsonMust(inputMap)))
		t.Fail()
	}
}
