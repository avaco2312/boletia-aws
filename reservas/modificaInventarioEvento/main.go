package main

import (
	"context"
	"fmt"
	"log"

	"github.com/avaco/clientes/contratos"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var svc *dynamodb.Client

func handleRequest(ctx context.Context, e events.DynamoDBEvent) {
	var evento contratos.Evento
	for _, record := range e.Records {
		if record.EventName == "REMOVE" {
			return
		}
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		evento.PK = record.Change.NewImage["PK"].String()
		evento.Estado = record.Change.NewImage["Estado"].String()
		evento.Categoria = record.Change.NewImage["Categoria"].String()
		capac, err := record.Change.NewImage["Capacidad"].Integer()
		if err != nil {
			fmt.Print("Error Unmarshal integer DynamoDB")
			log.Fatal()
		}
		evento.Capacidad = int(capac)
		itevento, err := attributevalue.MarshalMap(&evento)
		if err != nil {
			fmt.Print("Error MarshalMap DynamoDB")
			log.Fatal()
		}
		itevento["SK"] = itevento["PK"]
		itevento["PK"] = &types.AttributeValueMemberS{Value: "E"}
		_, err = svc.PutItem(context.TODO(),
			&dynamodb.PutItemInput{
				Item:      itevento,
				TableName: aws.String("inventario"),
			},
		)
		if err != nil {
			fmt.Print("Error putitem DynamoDB")
			log.Fatal()
		}
	}
}

func main() {
	lambda.Start(handleRequest)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
