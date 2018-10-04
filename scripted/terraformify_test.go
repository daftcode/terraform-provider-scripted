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
  "true": true,
  "nil": null,
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
	jsonDataTerraformified := `
{
  "id": "1",
  "true": "1",
  "nil": "",
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
	expected := map[string]interface{}{
		"%":           "5",
		"id":          "1",
		"true":        "1",
		"nil":         "",
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
			_, _ = os.Stderr.WriteString(fmt.Sprintf("[%#v] %#v != %#v\n", key, this, other))
			equal = false
		}
	}
	if Debug {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("input: %#v\n", input))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("output: %#v\n", output))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("expected: %#v\n", expected))
	}
	if !equal {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("output: %v\n", toPrettyJsonMust(output)))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("expected: %v\n", toPrettyJsonMust(output)))
		t.Fail()
	}

	bw := deterraformify(output)
	backwards := bw.(map[string]interface{})

	if Debug {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("backwards: %#v\n", backwards))
	}

	terraformified := fromJsonMust(jsonDataTerraformified).(map[string]interface{})

	keys = map[string]bool{}
	for key := range backwards {
		keys[key] = true
	}
	for key := range terraformified {
		keys[key] = true
	}
	for key := range keys {
		this := backwards[key]
		other := terraformified[key]
		if !reflect.DeepEqual(this, other) {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("[%#v] %#v != %#v\n", key, this, other))
			equal = false
		}
	}
	if !equal {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("output: %v\n", toPrettyJsonMust(backwards)))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("expected: %v\n", toPrettyJsonMust(terraformified)))
		t.Fail()
	}
}

func Test_TerraformifyPrimitive(t *testing.T) {
	expects := []interface{}{
		0, "0",
		1, "1",
		false, "0",
		true, "1",
		nil, "",
		1.1, "1.1",
		[3]interface{}{1, 2, 3}, "[1 2 3]",
		map[string]int{"1": 1}, "map[1:1]",
	}
	for i := 0; i < len(expects); i += 2 {
		k := expects[i]
		v := expects[i+1]
		output := terraformifyPrimitive(k)
		if output != v {
			t.Errorf("%#v does not equal %#v", output, v)
		}
	}
}

func Test_TerraformifyPrimitives(t *testing.T) {
	jsonData := `
{
  "id": 1,
  "true": true,
  "float": 1.123,
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
  "true": "1",
  "float": "1.123",
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
	input := fromJsonMust(jsonData)
	output := terraformifyPrimitives(input).(map[string]interface{})
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
			_, _ = os.Stderr.WriteString(fmt.Sprintf("[%#v] %#v != %#v\n", key, this, other))
			equal = false
		}
	}
	if Debug {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("input: %#v\n", input))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("output: %#v\n", output))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("expected: %#v\n", expected))
	}
	if !equal {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("output: %v\n", toPrettyJsonMust(output)))
		_, _ = os.Stderr.WriteString(fmt.Sprintf("expected: %v\n", toPrettyJsonMust(output)))
		t.Fail()
	}
}
