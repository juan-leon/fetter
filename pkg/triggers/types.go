package triggers

// ProcessRunner runs processes based on rule names.
type ProcessRunner interface {
	Run(name string, data *map[string]string) error
}
