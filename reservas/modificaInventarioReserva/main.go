package main

import (
	"context"
	"fmt"
	"log"

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
	record := e.Records[0]
	fmt.Println("ID " + record.EventName)
	if record.EventName == "REMOVE" {
		return
	}
	estado := record.Change.NewImage["Estado"].String()
	fmt.Println("Estado " + estado)
	if estado != "C" {
		return
	}
	fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
	evento := record.Change.NewImage["PK"].String()
	var eSK map[string]types.AttributeValue
	for ok := true; ok; {
		qo, err := svc.Query(context.TODO(),
			&dynamodb.QueryInput{
				IndexName:              aws.String("evento-email"),
				TableName:              aws.String("inventario"),
				ExclusiveStartKey:      eSK,
				ProjectionExpression:   aws.String("SK"),
				KeyConditionExpression: aws.String("#evento = :evento"),
				FilterExpression:       aws.String("#estado = :activa"),
				ExpressionAttributeNames: map[string]string{
					"#evento": "Evento",
					"#estado": "Estado",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":evento": &types.AttributeValueMemberS{Value: evento},
					":activa": &types.AttributeValueMemberS{Value: "A"},
				},
			})
		if err != nil {
			fmt.Print("Error queryng DynamoDB " + err.Error())
			log.Fatal()
		}
		if qo.Count != 0 {
			var id string
			for _, item := range qo.Items {
				err = attributevalue.Unmarshal(item["SK"], &id)
				fmt.Println(id)
				if err != nil {
					fmt.Print("Error unmarshal DynamoDB")
					log.Fatal()
				}
				_, _ = svc.UpdateItem(context.TODO(),
					&dynamodb.UpdateItemInput{
						ConditionExpression: aws.String("#estado = :activa"),
						ExpressionAttributeNames: map[string]string{
							"#estado": "Estado",
						},
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":activa":    &types.AttributeValueMemberS{Value: "A"},
							":cancelada": &types.AttributeValueMemberS{Value: "C"},
						},
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "I"},
							"SK": &types.AttributeValueMemberS{Value: id},
						},
						TableName:        aws.String("inventario"),
						UpdateExpression: aws.String("SET #estado = :cancelada"),
					},
				)
			}
		}
		eSK = qo.LastEvaluatedKey
		ok = eSK != nil
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
