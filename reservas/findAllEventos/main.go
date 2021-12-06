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

func findAllEventos(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var eSK map[string]types.AttributeValue
	var eventos []contratos.REvento
	for ok := true; ok; {
		qo, err := svc.Query(context.TODO(),
			&dynamodb.QueryInput{
				TableName:                aws.String("inventario"),
				ExclusiveStartKey:        eSK,
				KeyConditionExpression:   aws.String("#pk = :evento"),
				ExpressionAttributeNames: map[string]string{"#pk": "PK"},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":evento": &types.AttributeValueMemberS{Value: "E"},
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
		var teventos []contratos.REvento
		err = attributevalue.UnmarshalListOfMaps(qo.Items, &teventos)
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
		eSK = qo.LastEvaluatedKey
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
	lambda.Start(findAllEventos)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error while retrieving AWS credentials")
	}
	svc = dynamodb.NewFromConfig(cfg)
}
