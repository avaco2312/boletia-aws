package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"fmt"

	"github.com/avaco/clientes/contratos"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/segmentio/ksuid"
)

var svc *dynamodb.Client

func insertReserva(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var reserva contratos.Reserva
	err := json.Unmarshal([]byte(request.Body), &reserva)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "JSON no valido",
		}, nil
	}
	reserva.Estado = "A"
	reserva.SK = ksuid.New().String()
	itreserva, err := attributevalue.MarshalMap(&reserva)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error marshal DynamoDB",
		}, nil
	}
	itreserva["PK"] = &types.AttributeValueMemberS{Value: "I"}
	_, err = svc.TransactWriteItems(context.TODO(),
		&dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Put: &types.Put{
						Item:      itreserva,
						TableName: aws.String("inventario"),
					},
				},
				{
					Update: &types.Update{
						ConditionExpression: aws.String("#estado = :activa AND #capacidad >= :boletos"),
						ExpressionAttributeNames: map[string]string{
							"#capacidad": "Capacidad",
							"#estado":    "Estado",
						},
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":boletos": &types.AttributeValueMemberN{Value: fmt.Sprint(reserva.Cantidad)},
							":activa":  &types.AttributeValueMemberS{Value: "A"},
						},
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "E"},
							"SK": &types.AttributeValueMemberS{Value: reserva.Evento},
						},
						TableName:        aws.String("inventario"),
						UpdateExpression: aws.String("SET #capacidad = #capacidad - :boletos"),
					},
				},
			},
		},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error insertando reserva en DynamoDB",
		}, nil
	}
	response, err := json.MarshalIndent(&reserva, "", "  ")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error Unmarshal JSON",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(response),
	}, nil
}

func main() {
	lambda.Start(insertReserva)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
