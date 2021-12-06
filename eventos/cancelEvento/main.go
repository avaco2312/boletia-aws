package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var svc *dynamodb.Client

func cancel(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
	_, err = svc.UpdateItem(context.TODO(),
		&dynamodb.UpdateItemInput{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: nombre},
			},
			TableName:           aws.String("eventos"),
			ConditionExpression: aws.String("#estado = :activa"),
			ExpressionAttributeNames: map[string]string{
				"#estado": "Estado",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cancelada": &types.AttributeValueMemberS{Value: "C"},
				":activa":    &types.AttributeValueMemberS{Value: "A"},
			},

			UpdateExpression: aws.String("SET #estado = :cancelada"),
		},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error cancelando evento en DynamoDB",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: string("Evento " + nombre + " cancelado"),
	}, nil
}

func main() {
	lambda.Start(cancel)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
