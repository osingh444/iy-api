package services

import (
	"github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/service/ses"

	"os"
	"fmt"

	"github.com/joho/godotenv"
)

var svc *ses.SES

const SENDER = "noreply@igreputation.com"
const CHARSET = "UTF-8"
const RESET_URL = "http://localhost:3000/reset?token="

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}

	sess, err := session.NewSession(&aws.Config{
		Region:aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(os.Getenv("awsID"), os.Getenv("secretKEY"), ""),
	})

	if err != nil {
		panic(err.Error())
	}

	svc = ses.New(sess)
}

func SendEmail(recipient string, sender string, title string, text string) error {
	input := &ses.SendEmailInput{
        Destination: &ses.Destination{
          CcAddresses: []*string{
          },
          ToAddresses: []*string{
            aws.String(recipient),
          },
        },
        Message: &ses.Message{
          Body: &ses.Body{
            Text: &ses.Content{
              Charset: aws.String(CHARSET),
              Data: aws.String(text),
            },
          },
          Subject: &ses.Content{
          	Charset: aws.String(CHARSET),
            Data: aws.String(title),
          },
        },
        Source: aws.String(sender),
			}

	_, err := svc.SendEmail(input)
	fmt.Println(err)
	return err
}

func SendEmailConfirmationEmail() error {
	return nil
}

func SendPasswordResetEmail(recipient string, token string) error {
	title := "Password Reset Requested"
	body := "<p> Password reset was requested.  Cick this link to reset your password. " + RESET_URL + token + "</p>"

	return SendEmail(recipient, SENDER, title, body)
}
