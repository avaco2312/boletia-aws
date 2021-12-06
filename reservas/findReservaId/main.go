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

func findReservaId(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, err := url.QueryUnescape(request.PathParameters["id"])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Id de reserva incorrecta",
		}, nil
	}
	res, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("inventario"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "I"},
			"SK": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error while fetching reserva from DynamoDB",
		}, nil
	}
	if len(res.Item) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Reserva Id " + id + " inexistente",
		}, nil
	}
	var reserva contratos.Reserva
	err = attributevalue.UnmarshalMap(res.Item, &reserva)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error while Unmarshal DynamoDB",
		}, nil
	}
	response, err := json.MarshalIndent(&reserva, "", "  ")
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
	lambda.Start(findReservaId)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
