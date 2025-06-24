package rules

// RuleExecutionResult armazena o resultado da execução de uma única regra.
type RuleExecutionResultBackup struct {
	Rule    string
	Passed  bool
	Details string // Ex: "$.idade (20) >= 18 (true)" ou mensagem de erro
}

// Policy e Context permanecem iguais ao código anterior
type Policy struct {
	Name  string   `yaml:"name"`
	Rules []string `yaml:"rules"`
}

type Context struct {
	Data map[string]interface{}
	Vars map[string]interface{}
}
