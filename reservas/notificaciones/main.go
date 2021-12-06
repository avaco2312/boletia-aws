package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var svc *ses.SES

func handleRequest(ctx context.Context, e events.DynamoDBEvent) {
	for _, record := range e.Records {
		if record.EventName != "REMOVE" && record.Change.NewImage["PK"].String() == "I" {

			fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
			id := record.Change.NewImage["SK"].String()
			evento := record.Change.NewImage["Evento"].String()
			email := record.Change.NewImage["Email"].String()
			estado := record.Change.NewImage["Estado"].String()
			can, err := record.Change.NewImage["Cantidad"].Integer()
			if err != nil {
				fmt.Print("Error cantidad del evento")
				log.Fatal()
			}
			err = enviaEmail(id, evento, estado, email, int(can))
			if err != nil {
				fmt.Print("Error enviando email")
				log.Fatal()
			}
		}
	}
}

var Sender string

const CharSet = "UTF-8"

type mensaje struct {
	subject  string
	textbody string
}

var mensajes = [...]mensaje{
	{
		subject:  "Confirmación de reserva",
		textbody: "Su reserva %s de %d boletos para el evento %s está confirmada",
	},
	{
		subject:  "Cancelación de reserva",
		textbody: "Su reserva %s de %d boletos para el evento %s fue cancelada, el evento fue suspendido por los organizadores",
	},
	{
		subject:  "Cancelación de reserva",
		textbody: "Su reserva %s de %d boletos para el evento %s fue cancelada a petición suya",
	},
}

func enviaEmail(id, evento, estado, email string, cantidad int) error {
	tipo := strings.Index("ACX", estado)
	if tipo == -1 {
		return errors.New("estado de la reserva no valido")
	}
	msg := fmt.Sprintf(mensajes[tipo].textbody, id, cantidad, evento)
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(msg),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(mensajes[tipo].subject),
			},
		},
		Source: aws.String(Sender),
	}
	_, err := svc.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
			return nil
		} else {
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}

func init() {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	svc = ses.New(sess)
	Sender = os.Getenv("SENDER_EMAIL")
}
