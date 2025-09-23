package main

import (
	"context"
	"fmt"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/raywall/cloud-service-pack/go/config"
)

func main() {
	cfg, _ := awscfg.LoadDefaultConfig(context.TODO())
	ssmClient := ssm.NewFromConfig(cfg)
	builder := config.New(ssmClient)

	ctx := context.Background()

	// // Connectors como array
	// connectorsJSON, err := builder.BuildConnectorsConfig(ctx, "/teste/app")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(connectorsJSON))

	// // Schema com estrutura mista
	// schemaJSON, err := builder.BuildSchemaConfig(ctx, "/teste/app")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(schemaJSON))

	// // Configuração completa
	// fullConfig, err := builder.BuildCompleteConfig(ctx, "/teste/app")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(fullConfig))

	// Configuração
	fullConfig, err := builder.BuildConfig(ctx, "/teste/app/schema")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(fullConfig))
}
