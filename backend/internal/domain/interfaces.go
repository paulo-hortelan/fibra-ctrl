package domain

type OLTDetails struct {
	ID       string
	IP       string
	Type     string
	Model    string
	Protocol string
}

type FunctionDefinition struct {
	Name        string   // Function name (camelCase)
	Description string   // Human-readable description
	Commands    []string // List of command templates to execute in order
	Parameters  []string // List of required parameter names
}

type CommandInfo struct {
	CommandStr string
	Order      int // Execution order if multiple commands are needed
}
