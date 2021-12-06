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

func findReservasEventoEmail(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var eSK map[string]types.AttributeValue
	evento, err := url.QueryUnescape(request.PathParameters["id"])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Nombre de evento incorrecto",
		}, nil
	}
	email, err := url.QueryUnescape(request.PathParameters["email"])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "Email incorrecto",
		}, nil
	}
	var reservas []contratos.Reserva
	for ok := true; ok; {
		qo, err := svc.Query(context.TODO(),
			&dynamodb.QueryInput{
				IndexName:                aws.String("evento-email"),
				TableName:                aws.String("inventario"),
				ExclusiveStartKey:        eSK,
				KeyConditionExpression:   aws.String("#evento = :evento AND #email = :email"),
				ExpressionAttributeNames: map[string]string{"#evento": "Evento", "#email": "Email"},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":evento": &types.AttributeValueMemberS{Value: evento},
					":email":  &types.AttributeValueMemberS{Value: email},
				},
			})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: "Error while queryng DynamoDB",
			}, nil
		}
		var treservas []contratos.Reserva
		err = attributevalue.UnmarshalListOfMaps(qo.Items, &treservas)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: "Error while Unmarshal DynamoDB",
			}, nil
		}
		reservas = append(reservas, treservas...)
		eSK = qo.LastEvaluatedKey
		ok = eSK != nil
	}
	response, err := json.MarshalIndent(&reservas, "", "  ")
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
	lambda.Start(findReservasEventoEmail)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
