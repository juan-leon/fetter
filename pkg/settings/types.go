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
	Path   string `config:"path,required"`
	Action string `config:"action,required"`
	Group  string `config:"group,required"`
}

type Audit struct {
	Mode string `config:"mode"`
}

type Group struct {
	Name   string `config:"name,required"`
	RAM    int64  `config:"ram"`
	CPU    int    `config:"cpu"`
	Pids   int64  `config:"pids"`
	Freeze bool   `group:"freeze"`
}

type Settings struct {
	Logging Logging `config:"logging,required"`
	Rules   []Rule  `config:"rules,required"`
	Groups  []Group `config:"groups,required"`
	Audit   Audit   `config:"audit"`
	Name    string  `config:"name,required"`
	Mode    string  `config:"mode,required"`
}
