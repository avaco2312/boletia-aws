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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var svc *dynamodb.Client

func findAll(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var eSK map[string]types.AttributeValue
	var eventos []contratos.Evento
	for ok := true; ok; {
		so, err := svc.Scan(context.TODO(),
			&dynamodb.ScanInput{
				TableName:         aws.String("eventos"),
				ExclusiveStartKey: eSK,
			},
		)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: "Error while scanning DynamoDB",
			}, nil
		}
		var teventos []contratos.Evento
		err = attributevalue.UnmarshalListOfMaps(so.Items, &teventos)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: "Error while Unmarshal DynamoDB",
			}, nil
		}
		eventos = append(eventos, teventos...)
		eSK = so.LastEvaluatedKey
		ok = eSK != nil
	}
	response, err := json.MarshalIndent(&eventos, "", "  ")
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
	lambda.Start(findAll)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
