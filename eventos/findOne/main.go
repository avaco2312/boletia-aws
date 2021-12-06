package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

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

func findOne(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	nombre, err := url.QueryUnescape(request.PathParameters["evento"])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Nombre de evento incorrecto",
		}, nil
	}
	res, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("eventos"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: nombre},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error while fetching evento from DynamoDB",
		}, nil
	}
	if len(res.Item) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Evento " + nombre + " inexistente",
		}, nil
	}
	var evento contratos.Evento
	err = attributevalue.UnmarshalMap(res.Item, &evento)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error while Unmarshal DynamoDB",
		}, nil
	}
	response, err := json.MarshalIndent(&evento, "", "  ")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error while Unmarshal JSON",
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
	lambda.Start(findOne)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
