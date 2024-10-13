package ispec

const (
	TagKeyArguments = "ukarg"
	TagKeyFlag      = "ukflag"
	TagKeyInline    = "ukinline"
)

func ConsumableSet(valid ...string) func(string) bool {
	set := make(map[string]struct{})
	for _, item := range valid {
		set[item] = struct{}{}
	}

	return func(s string) (ok bool) { _, ok = set[s]; return }
}
