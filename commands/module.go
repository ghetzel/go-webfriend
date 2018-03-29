package commands

type Module interface {
	ExecuteCommand(name string, arg interface{}, objargs map[string]interface{}) (interface{}, error)
}
