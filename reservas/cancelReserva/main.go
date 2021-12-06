package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var svc *dynamodb.Client

func cancelReserva(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
	gio, err := svc.GetItem(context.TODO(),
		&dynamodb.GetItemInput{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "I"},
				"SK": &types.AttributeValueMemberS{Value: id},
			},
			TableName:            aws.String("inventario"),
			ProjectionExpression: aws.String("Evento,Cantidad"),
		},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error buscando la reserva " + id,
		}, nil
	}
	var evento string
	err = attributevalue.Unmarshal(gio.Item["Evento"], &evento)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error buscando la reserva " + id,
		}, nil
	}
	var boletos int
	err = attributevalue.Unmarshal(gio.Item["Cantidad"], &boletos)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Error buscando la reserva " + id,
		}, nil
	}
	_, err = svc.TransactWriteItems(context.TODO(),
		&dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Update: &types.Update{
						ConditionExpression: aws.String("#estado = :activa"),
						ExpressionAttributeNames: map[string]string{
							"#estado": "Estado",
						},
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":activa":    &types.AttributeValueMemberS{Value: "A"},
							":cancelada": &types.AttributeValueMemberS{Value: "X"},
						},
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "I"},
							"SK": &types.AttributeValueMemberS{Value: id},
						},
						TableName:        aws.String("inventario"),
						UpdateExpression: aws.String("SET #estado = :cancelada"),
					},
				},
				{
					Update: &types.Update{
						ConditionExpression: aws.String("#estado = :activa"),
						ExpressionAttributeNames: map[string]string{
							"#capacidad": "Capacidad",
							"#estado":    "Estado",
						},
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":activa":  &types.AttributeValueMemberS{Value: "A"},
							":boletos": &types.AttributeValueMemberN{Value: fmt.Sprint(boletos)},
						},
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "E"},
							"SK": &types.AttributeValueMemberS{Value: evento},
						},
						TableName:        aws.String("inventario"),
						UpdateExpression: aws.String("SET #capacidad = #capacidad + :boletos"),
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
			Body: "Error cancelando reserva en DynamoDB",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: "Reserva Id " + id + " cancelada",
	}, nil
}

func main() {
	lambda.Start(cancelReserva)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
