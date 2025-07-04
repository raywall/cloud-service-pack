package graph

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raywall/cloud-service-pack/go/graphql/graph/connectors"
	"github.com/raywall/cloud-service-pack/go/graphql/types"
)

type Resolver interface {
	ResolveDataSource(p graphql.ResolveParams) (interface{}, error)
	AddConfig(cfg *types.Config) error
}

type resolver struct {
	dataConnectors map[string]connectors.Connector
	config         *types.Config
	logger         *slog.Logger
	mock           *mockResolver
	// apiClient      *adapters.APIClient
}

type mockResolver struct {
	Status bool                   `json:"status"`
	Values map[string]interface{} `json:"values"`
}

func NewResolver(cfg *types.Config, connectorConfig string) (Resolver, error) {
	connectors, err := connectors.LoadConnectors(cfg, connectorConfig)
	if err != nil {
		return nil, err
	}

	return &resolver{
		dataConnectors: connectors,
		logger:         slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		mock: &mockResolver{
			Status: false,
			Values: nil,
		},
		// apiClient:      adapters.NewAPIClient(),
	}, nil
}

func (r *resolver) AddConfig(cfg *types.Config) error {
	r.config = cfg

	// if jsonMock, err := ctx.GetParameterValue(local.New().GetEnvOrDefault("SSM_MOCK_VALUE", "/graphql/dev/mock"), false); err != nil && jsonMock != nil {
	// 	if err := json.Unmarshal([]byte(jsonMock.(string)), r.mock); err != nil {
	// 		return fmt.Errorf("failed to deserialize mock value from ssm: \n\t%v", err)
	// 	}
	// }

	return nil
}

func (r *resolver) ResolveDataSource(p graphql.ResolveParams) (interface{}, error) {
	var (
		result          = make(map[string]interface{})
		requestedFields = getRequestedFields(p.Info)
		errChan         = make(chan error, len(requestedFields))
		wg              sync.WaitGroup
	)

	// Context for timeout/cancellation control
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	for _, field := range requestedFields {
		conn, exists := r.dataConnectors[field]
		if !exists {
			errChan <- fmt.Errorf("no connector found for field: \n\t%s", field)
			continue
		}

		// if r.mock.Status {
		// 	if values, ok := r.mock.Values[field]; ok {
		// 		result[field] = (values.(map[string]interface{}))[fmt.Sprintf("%d", codigo)]
		// 	}
		// 	continue
		// }

		wg.Add(1)
		go func(field string, conn connectors.Connector) {
			defer wg.Done()
			data, err := conn.GetData(p.Args)
			if err != nil {
				// Log the error instead of sending it to the error channel
				r.logger.Error(fmt.Sprintf("error fetching %s", field), "error", err)
				return

				// select {
				// case errChan <- fmt.Errorf("error fetching %s: \n\t%w", field, err):
				// case <-ctx.Done():
				// 	return
				// }
				// return
			}
			result[field] = data
		}(field, conn)
	}

	wg.Wait()
	close(errChan)

	var combinedErr error
	for err := range errChan {
		combinedErr = errors.Join(combinedErr, err)
	}
	if combinedErr != nil {
		return nil, fmt.Errorf("error fetching data: \n\t%w", combinedErr)
	}

	return result, nil
}

func getRequestedFields(info graphql.ResolveInfo) []string {
	fields := make([]string, 0)

	if len(info.FieldASTs) == 0 {
		return fields
	}

	if info.FieldASTs[0].SelectionSet == nil {
		return fields
	}

	for _, selection := range info.FieldASTs[0].SelectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			fields = append(fields, sel.Name.Value)
		}
	}

	return fields
}
