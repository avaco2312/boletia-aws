package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/avaco/clientes/contratos"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var svc *dynamodb.Client

func insert(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var evento contratos.Evento
	err := json.Unmarshal([]byte(request.Body), &evento)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "JSON no valido",
		}, nil
	}
	evento.Estado = "A"
	itevento, err := attributevalue.MarshalMap(&evento)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error marshal DynamoDB",
		}, nil
	}
	_, err = svc.PutItem(context.TODO(),
		&dynamodb.PutItemInput{
			Item:                itevento,
			TableName:           aws.String("eventos"),
			ConditionExpression: aws.String("attribute_not_exists(#pk)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "PK",
			},
		},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error insertando evento en DynamoDB",
		}, nil
	}
	response, err := json.MarshalIndent(&evento, "", "  ")
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
	lambda.Start(insert)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
