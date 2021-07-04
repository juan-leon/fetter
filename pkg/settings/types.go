package settings

const (
	// RunModeAudit is the string used to configure audit mode
	RunModeAudit string = "audit"
	// RunModeScanner is the string used to configure scanner mode
	RunModeScanner string = "scanner"
)

// Logging holds the configuration options referred to logging
type Logging struct {
	File  string `config:"file"`
	Level string `config:"level"`
}

// Rule holds the configuration options referred to a single rule
type Rule struct {
	Paths   []string `config:"paths,required"`
	Action  string   `config:"action,required"`
	Group   string   `config:"group"`
	Trigger string   `config:"trigger"`
}

// Audit holds the configuration options referred to a audit mode
type Audit struct {
	Mode string `config:"mode"`
}

// Group holds the configuration options referred to a single process group
type Group struct {
	RAM    int64 `config:"ram"`
	CPU    int   `config:"cpu"`
	Pids   int64 `config:"pids"`
	Freeze bool  `group:"freeze"`
}

// Trigger holds the configuration options referred to a single trigger
type Trigger struct {
	Run  string   `config:"run"`
	Args []string `config:"args"`
	User string   `config:"user"`
}

// Settings holds the configuration options referred to the whole application
type Settings struct {
	Logging  Logging            `config:"logging,required"`
	Rules    map[string]Rule    `config:"rules,required"`
	Groups   map[string]Group   `config:"groups"`
	Triggers map[string]Trigger `config:"triggers"`
	Audit    Audit              `config:"audit"`
	Name     string             `config:"name,required"`
	Mode     string             `config:"mode,required"`
}

// GetGroup returns the name of a group configured for a rule
func (s *Settings) GetGroup(rule string) string {
	return s.Rules[rule].Group
}

// GetTrigger returns the name of a trigger configured for a rule
func (s *Settings) GetTrigger(rule string) string {
	return s.Rules[rule].Trigger
}
