package scripted

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"os"
	"strconv"
	"strings"
)

func defaultEmptyString() (interface{}, error) {
	return EnvEmptyString, nil
}

func defaultMsg(msg, defVal string) string {
	defVal = strings.Replace(defVal, "$", "$"+EnvPrefix, -1)
	return fmt.Sprintf("%s Defaults to: %s", msg, defVal)
}

func envKey(key string) (ret string) {
	if strings.HasPrefix(key, EnvPrefix) {
		ret = key
	} else {
		ret = EnvPrefix + key
	}
	if debugLogging {
		_, _ = Stderr.WriteString(fmt.Sprintf(`envKey("%s") -> ("%s")`+"\n", key, ret))
	}
	return ret
}

func envDefaultFunc(key, defVal string) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		ret, _ := getEnv(key, defVal)
		return ret, nil
	}
}

func getEnvMust(key, defValue string) string {
	val, _ := getEnv(key, defValue)
	return val
}

func getEnv(key, defValue string) (value string, ok bool) {
	return envDefaultOk(envKey(key), defValue)
}

func getEnvList(key string, defValue []string) (value []string, ok bool, err error) {
	val, ok := envDefaultOk(envKey(key), EnvEmptyString)
	if !ok {
		return defValue, ok, nil
	}
	json, err := fromJson(val)
	if err != nil {
		return nil, false, err
	}
	value = castConfigListString(json)
	return value, ok, err
}

func getEnvBoolOk(key string, defVal bool) (value, ok bool) {
	str, ok := getEnv(key, EnvEmptyString)
	if str == EnvEmptyString {
		return defVal, false
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		ok = false
	}
	return value, ok
}

func getEnvBool(key string, defVal bool) (value bool) {
	value, _ = getEnvBoolOk(key, defVal)
	return value
}

func getEnvBoolFalse(key string) bool {
	return getEnvBool(key, false)
}

//noinspection GoUnusedFunction
func getEnvBoolTrue(key string) bool {
	return getEnvBool(key, true)
}

func envDefault(key, defValue string) string {
	ret, _ := envDefaultOk(key, defValue)
	return ret
}

func envDefaultOk(key, defValue string) (value string, ok bool) {
	value, ok = os.LookupEnv(key)
	if !ok {
		value = defValue
	}
	if debugLogging {
		_, _ = Stderr.WriteString(fmt.Sprintf(`envDefaultOk("%s", "%s") -> ("%s", %v)`+"\n", key, defValue, value, ok))
	}
	return value, ok
}

func stringDefaultSchemaEmpty(schema *schema.Schema, key, description string) *schema.Schema {
	return stringDefaultSchemaMsgVal(schema, key, description, "not set")
}

func stringDefaultSchemaEmptyMsgVal(s *schema.Schema, key, description, msgVal string) *schema.Schema {
	return stringDefaultSchemaBaseOr(s, key, description, EnvEmptyString, msgVal)
}
func stringDefaultSchema(s *schema.Schema, key, description, defVal string) *schema.Schema {
	return stringDefaultSchemaBaseOr(s, key, description, defVal, fmt.Sprintf("`%s`", defVal))
}
func stringDefaultSchemaBaseOr(s *schema.Schema, key, description, defVal, msgVal string) *schema.Schema {
	if msgVal != "" {
		msgVal = " or " + msgVal
	}
	return stringDefaultSchemaBase(s, key, description, defVal, msgVal)
}
func stringDefaultSchemaBase(s *schema.Schema, key, description, defVal, msgVal string) *schema.Schema {
	if s == nil {
		s = &schema.Schema{}
	}
	key = strings.ToUpper(key)
	s.Type = schema.TypeString
	s.Optional = true
	s.DefaultFunc = envDefaultFunc(key, defVal)
	msg := fmt.Sprintf("`$%s`%s", key, msgVal)
	s.Description = defaultMsg(description, msg)
	return s
}

func stringDefaultSchemaMsgVal(s *schema.Schema, key, description, msgVal string) *schema.Schema {
	return stringDefaultSchemaBaseOr(s, key, description, EnvEmptyString, msgVal)
}

func boolDefaultSchema(s *schema.Schema, key, description string, defVal bool) *schema.Schema {
	key = strings.ToUpper(key)
	prefix := "="
	if defVal {
		prefix = "!"
	}
	s = stringDefaultSchemaBase(s, key, description, EnvEmptyString, fmt.Sprintf(" %s= `\"\"`", prefix))
	s.DefaultFunc = func() (interface{}, error) {
		value, ok := getEnvBoolOk(key, defVal)
		if !ok {
			value = defVal
		}
		return value, nil
	}
	s.Type = schema.TypeBool
	return s
}

func floatDefaultSchema(s *schema.Schema, key, description string, defVal float64) *schema.Schema {
	key = strings.ToUpper(key)
	s = stringDefaultSchemaMsgVal(s, key, description, "")
	s.DefaultFunc = func() (interface{}, error) {
		value, ok := getEnv(key, "0")
		if ok {
			i, err := strconv.ParseFloat(value, 64)
			if err == nil {
				return i, nil
			}
		}
		return defVal, nil
	}
	s.Type = schema.TypeFloat
	return s
}

//noinspection GoUnusedFunction
func intDefaultSchema(s *schema.Schema, key, description string, defVal int) *schema.Schema {
	key = strings.ToUpper(key)
	s = stringDefaultSchemaMsgVal(s, key, description, "")
	s.DefaultFunc = func() (interface{}, error) {
		value, ok := getEnv(key, "0")
		if ok {
			i, err := strconv.Atoi(value)
			if err == nil {
				return i, nil
			}
		}
		return defVal, nil
	}
	s.Type = schema.TypeInt
	return s
}
