package triggers

type ProcessRunner interface {
	Run(name string, data *map[string]string) error
}
