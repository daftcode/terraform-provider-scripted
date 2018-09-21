# Template functions listing

Sprig docs are available at [http://masterminds.github.io/sprig/](http://masterminds.github.io/sprig/)

Builtin go template functions are available at the [official docs](https://golang.org/pkg/text/template/#hdr-Functions)

| Name | Source | Type |
|:--- | --- | --- |
| `abbrev` | `sprig` | `func(int, string) string` |
| `abbrevboth` | `sprig` | `func(int, int, string) string` |
| `add` | `sprig` | `func(...interface {}) int64` |
| `add1` | `sprig` | `func(interface {}) int64` |
| `ago` | `sprig` | `func(interface {}) string` |
| `append` | `sprig` | `func(interface {}, interface {}) []interface {}` |
| `atoi` | `sprig` | `func(string) int` |
| `b32dec` | `sprig` | `func(string) string` |
| `b32enc` | `sprig` | `func(string) string` |
| `b64dec` | `sprig` | `func(string) string` |
| `b64enc` | `sprig` | `func(string) string` |
| `base` | `sprig` | `func(string) string` |
| `biggest` | `sprig` | `func(interface {}, ...interface {}) int64` |
| `buildCustomCert` | `sprig` | `func(string, string) (sprig.certificate, error)` |
| `camelcase` | `sprig` | `func(string) string` |
| `cat` | `sprig` | `func(...interface {}) string` |
| `ceil` | `sprig` | `func(interface {}) float64` |
| `clean` | `sprig` | `func(string) string` |
| `coalesce` | `sprig` | `func(...interface {}) interface {}` |
| `compact` | `sprig` | `func(interface {}) []interface {}` |
| `contains` | `sprig` | `func(string, string) bool` |
| `date` | `sprig` | `func(string, interface {}) string` |
| `dateInZone` | `sprig` | `func(string, interface {}, string) string` |
| `dateModify` | `sprig` | `func(string, time.Time) time.Time` |
| `date_in_zone` | `sprig` | `func(string, interface {}, string) string` |
| `date_modify` | `sprig` | `func(string, time.Time) time.Time` |
| `default` | `sprig` | `func(interface {}, ...interface {}) interface {}` |
| `derivePassword` | `sprig` | `func(uint32, string, string, string, string) string` |
| `dict` | `sprig` | `func(...interface {}) map[string]interface {}` |
| `dir` | `sprig` | `func(string) string` |
| `div` | `sprig` | `func(interface {}, interface {}) int64` |
| `empty` | `sprig` | `func(interface {}) bool` |
| `ext` | `sprig` | `func(string) string` |
| `fail` | `sprig` | `func(string) (string, error)` |
| `first` | `sprig` | `func(interface {}) interface {}` |
| `float64` | `sprig` | `func(interface {}) float64` |
| `floor` | `sprig` | `func(interface {}) float64` |
| `fromJson` | `scripted` | `func(string) (interface {}, error)` |
| `fromYaml` | `scripted` | `func(string) (interface {}, error)` |
| `genCA` | `sprig` | `func(string, int) (sprig.certificate, error)` |
| `genPrivateKey` | `sprig` | `func(string) string` |
| `genSelfSignedCert` | `sprig` | `func(string, []interface {}, []interface {}, int) (sprig.certificate, error)` |
| `genSignedCert` | `sprig` | `func(string, []interface {}, []interface {}, int, sprig.certificate) (sprig.certificate, error)` |
| `has` | `sprig` | `func(interface {}, interface {}) bool` |
| `hasKey` | `sprig` | `func(map[string]interface {}, string) bool` |
| `hasPrefix` | `sprig` | `func(string, string) bool` |
| `hasSuffix` | `sprig` | `func(string, string) bool` |
| `hello` | `sprig` | `func() string` |
| `htmlDate` | `sprig` | `func(interface {}) string` |
| `htmlDateInZone` | `sprig` | `func(interface {}, string) string` |
| `include` | `scripted` | `func(string, interface {}) string` |
| `indent` | `sprig` | `func(int, string) string` |
| `initial` | `sprig` | `func(interface {}) []interface {}` |
| `initials` | `sprig` | `func(string) string` |
| `int` | `sprig` | `func(interface {}) int` |
| `int64` | `sprig` | `func(interface {}) int64` |
| `is` | `scripted` | `func(interface {}, interface {}) bool` |
| `isAbs` | `sprig` | `func(string) bool` |
| `isFilled` | `scripted` | `func(interface {}) bool` |
| `isSet` | `scripted` | `func(interface {}) bool` |
| `join` | `sprig` | `func(string, interface {}) string` |
| `keys` | `sprig` | `func(...map[string]interface {}) []string` |
| `kindIs` | `sprig` | `func(string, interface {}) bool` |
| `kindOf` | `sprig` | `func(interface {}) string` |
| `last` | `sprig` | `func(interface {}) interface {}` |
| `list` | `sprig` | `func(...interface {}) []interface {}` |
| `lower` | `sprig` | `func(string) string` |
| `max` | `sprig` | `func(interface {}, ...interface {}) int64` |
| `merge` | `sprig` | `func(map[string]interface {}, ...map[string]interface {}) interface {}` |
| `min` | `sprig` | `func(interface {}, ...interface {}) int64` |
| `mod` | `sprig` | `func(interface {}, interface {}) int64` |
| `mul` | `sprig` | `func(interface {}, ...interface {}) int64` |
| `nindent` | `sprig` | `func(int, string) string` |
| `nospace` | `sprig` | `func(string) string` |
| `now` | `sprig` | `func() time.Time` |
| `omit` | `sprig` | `func(map[string]interface {}, ...string) map[string]interface {}` |
| `pick` | `sprig` | `func(map[string]interface {}, ...string) map[string]interface {}` |
| `pluck` | `sprig` | `func(string, ...map[string]interface {}) []interface {}` |
| `plural` | `sprig` | `func(string, string, int) string` |
| `prepend` | `sprig` | `func(interface {}, interface {}) []interface {}` |
| `push` | `sprig` | `func(interface {}, interface {}) []interface {}` |
| `quote` | `sprig` | `func(...interface {}) string` |
| `randAlpha` | `sprig` | `func(int) string` |
| `randAlphaNum` | `sprig` | `func(int) string` |
| `randAscii` | `sprig` | `func(int) string` |
| `randNumeric` | `sprig` | `func(int) string` |
| `regexFind` | `sprig` | `func(string, string) string` |
| `regexFindAll` | `sprig` | `func(string, string, int) []string` |
| `regexMatch` | `sprig` | `func(string, string) bool` |
| `regexReplaceAll` | `sprig` | `func(string, string, string) string` |
| `regexReplaceAllLiteral` | `sprig` | `func(string, string, string) string` |
| `regexSplit` | `sprig` | `func(string, string, int) []string` |
| `repeat` | `sprig` | `func(int, string) string` |
| `replace` | `sprig` | `func(string, string, string) string` |
| `required` | `scripted` | `func(string, interface {}) interface {}` |
| `rest` | `sprig` | `func(interface {}) []interface {}` |
| `reverse` | `sprig` | `func(interface {}) []interface {}` |
| `round` | `sprig` | `func(interface {}, int, ...float64) float64` |
| `semver` | `sprig` | `func(string) (*semver.Version, error)` |
| `semverCompare` | `sprig` | `func(string, string) (bool, error)` |
| `set` | `sprig` | `func(map[string]interface {}, string, interface {}) map[string]interface {}` |
| `sha1sum` | `sprig` | `func(string) string` |
| `sha256sum` | `sprig` | `func(string) string` |
| `shuffle` | `sprig` | `func(string) string` |
| `snakecase` | `sprig` | `func(string) string` |
| `sortAlpha` | `sprig` | `func(interface {}) []string` |
| `split` | `sprig` | `func(string, string) map[string]string` |
| `splitList` | `sprig` | `func(string, string) []string` |
| `squote` | `sprig` | `func(...interface {}) string` |
| `sub` | `sprig` | `func(interface {}, interface {}) int64` |
| `substr` | `sprig` | `func(int, int, string) string` |
| `swapcase` | `sprig` | `func(string) string` |
| `ternary` | `sprig` | `func(interface {}, interface {}, bool) interface {}` |
| `terraformifyValues` | `scripted` | `func(interface {}) interface {}` |
| `title` | `sprig` | `func(string) string` |
| `toDate` | `sprig` | `func(string, string) time.Time` |
| `toJson` | `scripted` | `func(interface {}) (string, error)` |
| `toPrettyJson` | `scripted` | `func(interface {}) (string, error)` |
| `toString` | `sprig` | `func(interface {}) string` |
| `toStrings` | `sprig` | `func(interface {}) []string` |
| `toYaml` | `scripted` | `func(interface {}) (string, error)` |
| `trim` | `sprig` | `func(string) string` |
| `trimAll` | `sprig` | `func(string, string) string` |
| `trimPrefix` | `sprig` | `func(string, string) string` |
| `trimSuffix` | `sprig` | `func(string, string) string` |
| `trimall` | `sprig` | `func(string, string) string` |
| `trunc` | `sprig` | `func(int, string) string` |
| `tuple` | `sprig` | `func(...interface {}) []interface {}` |
| `typeIs` | `sprig` | `func(string, interface {}) bool` |
| `typeIsLike` | `sprig` | `func(string, interface {}) bool` |
| `typeOf` | `sprig` | `func(interface {}) string` |
| `uniq` | `sprig` | `func(interface {}) []interface {}` |
| `unset` | `sprig` | `func(map[string]interface {}, string) map[string]interface {}` |
| `until` | `sprig` | `func(int) []int` |
| `untilStep` | `sprig` | `func(int, int, int) []int` |
| `untitle` | `sprig` | `func(string) string` |
| `upper` | `sprig` | `func(string) string` |
| `uuidv4` | `sprig` | `func() string` |
| `without` | `sprig` | `func(interface {}, ...interface {}) []interface {}` |
| `wrap` | `sprig` | `func(int, string) string` |
| `wrapWith` | `sprig` | `func(int, string, string) string` |
