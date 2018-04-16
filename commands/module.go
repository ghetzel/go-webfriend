package commands

type DocItem struct {
	Name        string   `json:"name"`
	Types       []string `json:"types"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
}
type CallDoc struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Argument    *DocItem    `json:"argument,omitempty"`
	Options     interface{} `json:"options"`
	Return      interface{} `json:"return,omitempty"`
}

type ModuleDoc struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Commands    []CallDoc `json:"commands"`
}

type Module interface {
	ExecuteCommand(name string, arg interface{}, objargs map[string]interface{}) (interface{}, error)
}
