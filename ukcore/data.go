package ukcore

type Input struct {
	Program   string
	Target    []string
	Arguments []string
	Flags     []Flag
}

type Flag struct{ Name, Value string }
