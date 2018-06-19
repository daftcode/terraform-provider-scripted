# Template functions listing

## Scripted functions

| Name | Type | Overrides sprig? |
|:--- | --- | --- |
| `fromJson` | `func(string) (interface {}, error)` |  |
| `fromYaml` | `func(string) (interface {}, error)` |  |
| `is` | `func(interface {}, interface {}) bool` |  |
| `isFilled` | `func(string) bool` |  |
| `isSet` | `func(string) bool` |  |
| `toJson` | `func(interface {}) (string, error)` | yes |
| `toPrettyJson` | `func(interface {}) (string, error)` | yes |
| `toYaml` | `func(interface {}) (string, error)` |  |

## Sprig functions

Sprig docs are available at [http://masterminds.github.io/sprig/](http://masterminds.github.io/sprig/)

| Name | Type | Is overriden? | 
|:--- | --- | --- | --- |
| `abbrev` | `func(int, string) string` |  |
| `abbrevboth` | `func(int, int, string) string` |  |
| `add` | `func(...interface {}) int64` |  |
| `add1` | `func(interface {}) int64` |  |
| `ago` | `func(interface {}) string` |  |
| `append` | `func(interface {}, interface {}) []interface {}` |  |
| `atoi` | `func(string) int` |  |
| `b32dec` | `func(string) string` |  |
| `b32enc` | `func(string) string` |  |
| `b64dec` | `func(string) string` |  |
| `b64enc` | `func(string) string` |  |
| `base` | `func(string) string` |  |
| `biggest` | `func(interface {}, ...interface {}) int64` |  |
| `buildCustomCert` | `func(string, string) (sprig.certificate, error)` |  |
| `camelcase` | `func(string) string` |  |
| `cat` | `func(...interface {}) string` |  |
| `ceil` | `func(interface {}) float64` |  |
| `clean` | `func(string) string` |  |
| `coalesce` | `func(...interface {}) interface {}` |  |
| `compact` | `func(interface {}) []interface {}` |  |
| `contains` | `func(string, string) bool` |  |
| `date` | `func(string, interface {}) string` |  |
| `dateInZone` | `func(string, interface {}, string) string` |  |
| `dateModify` | `func(string, time.Time) time.Time` |  |
| `date_in_zone` | `func(string, interface {}, string) string` |  |
| `date_modify` | `func(string, time.Time) time.Time` |  |
| `default` | `func(interface {}, ...interface {}) interface {}` |  |
| `derivePassword` | `func(uint32, string, string, string, string) string` |  |
| `dict` | `func(...interface {}) map[string]interface {}` |  |
| `dir` | `func(string) string` |  |
| `div` | `func(interface {}, interface {}) int64` |  |
| `empty` | `func(interface {}) bool` |  |
| `ext` | `func(string) string` |  |
| `fail` | `func(string) (string, error)` |  |
| `first` | `func(interface {}) interface {}` |  |
| `float64` | `func(interface {}) float64` |  |
| `floor` | `func(interface {}) float64` |  |
| `genCA` | `func(string, int) (sprig.certificate, error)` |  |
| `genPrivateKey` | `func(string) string` |  |
| `genSelfSignedCert` | `func(string, []interface {}, []interface {}, int) (sprig.certificate, error)` |  |
| `genSignedCert` | `func(string, []interface {}, []interface {}, int, sprig.certificate) (sprig.certificate, error)` |  |
| `has` | `func(interface {}, interface {}) bool` |  |
| `hasKey` | `func(map[string]interface {}, string) bool` |  |
| `hasPrefix` | `func(string, string) bool` |  |
| `hasSuffix` | `func(string, string) bool` |  |
| `hello` | `func() string` |  |
| `htmlDate` | `func(interface {}) string` |  |
| `htmlDateInZone` | `func(interface {}, string) string` |  |
| `indent` | `func(int, string) string` |  |
| `initial` | `func(interface {}) []interface {}` |  |
| `initials` | `func(string) string` |  |
| `int` | `func(interface {}) int` |  |
| `int64` | `func(interface {}) int64` |  |
| `isAbs` | `func(string) bool` |  |
| `join` | `func(string, interface {}) string` |  |
| `keys` | `func(...map[string]interface {}) []string` |  |
| `kindIs` | `func(string, interface {}) bool` |  |
| `kindOf` | `func(interface {}) string` |  |
| `last` | `func(interface {}) interface {}` |  |
| `list` | `func(...interface {}) []interface {}` |  |
| `lower` | `func(string) string` |  |
| `max` | `func(interface {}, ...interface {}) int64` |  |
| `merge` | `func(map[string]interface {}, ...map[string]interface {}) interface {}` |  |
| `min` | `func(interface {}, ...interface {}) int64` |  |
| `mod` | `func(interface {}, interface {}) int64` |  |
| `mul` | `func(interface {}, ...interface {}) int64` |  |
| `nindent` | `func(int, string) string` |  |
| `nospace` | `func(string) string` |  |
| `now` | `func() time.Time` |  |
| `omit` | `func(map[string]interface {}, ...string) map[string]interface {}` |  |
| `pick` | `func(map[string]interface {}, ...string) map[string]interface {}` |  |
| `pluck` | `func(string, ...map[string]interface {}) []interface {}` |  |
| `plural` | `func(string, string, int) string` |  |
| `prepend` | `func(interface {}, interface {}) []interface {}` |  |
| `push` | `func(interface {}, interface {}) []interface {}` |  |
| `quote` | `func(...interface {}) string` |  |
| `randAlpha` | `func(int) string` |  |
| `randAlphaNum` | `func(int) string` |  |
| `randAscii` | `func(int) string` |  |
| `randNumeric` | `func(int) string` |  |
| `regexFind` | `func(string, string) string` |  |
| `regexFindAll` | `func(string, string, int) []string` |  |
| `regexMatch` | `func(string, string) bool` |  |
| `regexReplaceAll` | `func(string, string, string) string` |  |
| `regexReplaceAllLiteral` | `func(string, string, string) string` |  |
| `regexSplit` | `func(string, string, int) []string` |  |
| `repeat` | `func(int, string) string` |  |
| `replace` | `func(string, string, string) string` |  |
| `rest` | `func(interface {}) []interface {}` |  |
| `reverse` | `func(interface {}) []interface {}` |  |
| `round` | `func(interface {}, int, ...float64) float64` |  |
| `semver` | `func(string) (*semver.Version, error)` |  |
| `semverCompare` | `func(string, string) (bool, error)` |  |
| `set` | `func(map[string]interface {}, string, interface {}) map[string]interface {}` |  |
| `sha1sum` | `func(string) string` |  |
| `sha256sum` | `func(string) string` |  |
| `shuffle` | `func(string) string` |  |
| `snakecase` | `func(string) string` |  |
| `sortAlpha` | `func(interface {}) []string` |  |
| `split` | `func(string, string) map[string]string` |  |
| `splitList` | `func(string, string) []string` |  |
| `squote` | `func(...interface {}) string` |  |
| `sub` | `func(interface {}, interface {}) int64` |  |
| `substr` | `func(int, int, string) string` |  |
| `swapcase` | `func(string) string` |  |
| `ternary` | `func(interface {}, interface {}, bool) interface {}` |  |
| `title` | `func(string) string` |  |
| `toDate` | `func(string, string) time.Time` |  |
| `toJson` | `func(interface {}) string` | yes |
| `toPrettyJson` | `func(interface {}) string` | yes |
| `toString` | `func(interface {}) string` |  |
| `toStrings` | `func(interface {}) []string` |  |
| `trim` | `func(string) string` |  |
| `trimAll` | `func(string, string) string` |  |
| `trimPrefix` | `func(string, string) string` |  |
| `trimSuffix` | `func(string, string) string` |  |
| `trimall` | `func(string, string) string` |  |
| `trunc` | `func(int, string) string` |  |
| `tuple` | `func(...interface {}) []interface {}` |  |
| `typeIs` | `func(string, interface {}) bool` |  |
| `typeIsLike` | `func(string, interface {}) bool` |  |
| `typeOf` | `func(interface {}) string` |  |
| `uniq` | `func(interface {}) []interface {}` |  |
| `unset` | `func(map[string]interface {}, string) map[string]interface {}` |  |
| `until` | `func(int) []int` |  |
| `untilStep` | `func(int, int, int) []int` |  |
| `untitle` | `func(string) string` |  |
| `upper` | `func(string) string` |  |
| `uuidv4` | `func() string` |  |
| `without` | `func(interface {}, ...interface {}) []interface {}` |  |
| `wrap` | `func(int, string) string` |  |
| `wrapWith` | `func(int, string, string) string` |  |

## Builtin functions

Available in the [official docs](https://golang.org/pkg/text/template/#hdr-Functions)
