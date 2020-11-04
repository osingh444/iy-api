package utils

import (
	"os"
	"fmt"
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

var bucket string
var region string
var bucketURL string

var svc *s3.S3

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}

	bucket = os.Getenv("bucket")
	region = os.Getenv("region")
	bucketURL = fmt.Sprintf("%s.s3-%s.amazonaws.com", bucket, region)
	svc = generateNewSession()
}

func generateNewSession() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(os.Getenv("awsID"), os.Getenv("secretKEY"), ""),
	})

	if err != nil {
		panic(err)
	}
	return s3.New(sess)
}

func PutToS3(finalName, filetype string, fileReader *bytes.Reader, fileSize int64) error {
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &finalName,
		Body:          fileReader,
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(filetype),
	})
	if err != nil {
		fmt.Println(err, "puterr")
		return err
	}
	return nil
}

func DeleteFromS3(objectKey string) (error) {
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
		if err != nil {
			return err
		}

	return nil
}

func GenerateS3URL(str string) string {
	return fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", bucket, region, str)
}
