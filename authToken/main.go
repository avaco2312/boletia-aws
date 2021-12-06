package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func authToken(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user User
	err := json.Unmarshal([]byte(request.Body), &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid payload",
		}, nil
	}

	cognitoClient := cognito.New(cognito.Options{
		Region: "us-west-2",
	})
	authTry := &cognito.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": user.Username,
			"PASSWORD": user.Password,
		},
		ClientId: aws.String("54qcmhjf62m3305453ls4ujq40"),
	}
	res, err := cognitoClient.InitiateAuth(context.TODO(), authTry)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{ "Authorization": "` + *res.AuthenticationResult.IdToken + `"}`,
	}, nil
}

func main() {
	lambda.Start(authToken)
}
