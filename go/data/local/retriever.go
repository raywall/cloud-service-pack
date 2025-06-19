package local

import "github.com/raywall/cloud-service-pack/go/data/types"

// parser é a interface genérica para extrair dados de configurações inline para a cloud AWS
type parser interface {
	Extract() (*types.Config, error)
}
