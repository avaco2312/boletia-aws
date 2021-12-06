package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err.Error())
	}
	svc := dynamodb.NewFromConfig(cfg)
		_, err = svc.CreateTable(context.TODO(),
		&dynamodb.CreateTableInput{
			TableName:   aws.String("eventos"),
			BillingMode: types.BillingModeProvisioned,
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("PK"),
					KeyType:       types.KeyTypeHash,
				},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("PK"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			StreamSpecification: &types.StreamSpecification{
				StreamEnabled:  aws.Bool(true),
				StreamViewType: types.StreamViewTypeNewImage,
			},
		},
	)
	if err != nil {
		log.Fatal("error creando tabla eventos")
	}
	_, err = svc.CreateTable(context.TODO(),
		&dynamodb.CreateTableInput{
			TableName:   aws.String("inventario"),
			BillingMode: types.BillingModeProvisioned,
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("PK"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: aws.String("SK"),
					KeyType:       types.KeyTypeRange,
				},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("PK"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("SK"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("Evento"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("Email"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			StreamSpecification: &types.StreamSpecification{
				StreamEnabled:  aws.Bool(true),
				StreamViewType: types.StreamViewTypeNewImage,
			},			
			GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("evento-email"),
					Projection: &types.Projection{
						ProjectionType: types.ProjectionTypeAll,
					},
					ProvisionedThroughput: &types.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(1),
						WriteCapacityUnits: aws.Int64(1),
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("Evento"),
							KeyType:       types.KeyTypeHash,
						},
						{
							AttributeName: aws.String("Email"),
							KeyType:       types.KeyTypeRange,
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.Fatal("error creando tabla inventario")
	}
}
