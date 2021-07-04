package settings

const (
	RUN_MODE_AUDIT   string = "audit"
	RUN_MODE_SCANNER string = "scanner"
)

type Logging struct {
	File  string `config:"file"`
	Level string `config:"level"`
}

type Rule struct {
	Paths   []string `config:"paths,required"`
	Action  string   `config:"action,required"`
	Group   string   `config:"group"`
	Trigger string   `config:"trigger"`
}

type Audit struct {
	Mode string `config:"mode"`
}

type Group struct {
	RAM    int64 `config:"ram"`
	CPU    int   `config:"cpu"`
	Pids   int64 `config:"pids"`
	Freeze bool  `group:"freeze"`
}

type Trigger struct {
	Run  string   `config:"run"`
	Args []string `config:"args"`
	User string   `config:"user"`
}

type Settings struct {
	Logging  Logging            `config:"logging,required"`
	Rules    map[string]Rule    `config:"rules,required"`
	Groups   map[string]Group   `config:"groups"`
	Triggers map[string]Trigger `config:"triggers"`
	Audit    Audit              `config:"audit"`
	Name     string             `config:"name,required"`
	Mode     string             `config:"mode,required"`
}

func (s *Settings) GetGroup(rule string) string {
	return s.Rules[rule].Group
}

func (s *Settings) GetTrigger(rule string) string {
	return s.Rules[rule].Trigger
}
