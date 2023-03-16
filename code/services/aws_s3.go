package services

// import (
// 	"bytes"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3/s3manager"
// )

// var uploader *s3manager.Uploader

// func NewUploader() *s3manager.Uploader {
// 	s3Config := &aws.Config{
// 		Region: aws.String("us-west-2"),
// 		// Credentials: credentials.NewStaticCredentials("KeyID", "SecretKey", ""),
// 	}

// 	s3Session := session.New(s3Config)

// 	uploader := s3manager.NewUploader(s3Session)
// 	return uploader
// }

// func Upload(ImageBytes []byte) {
// 	NewUploader()
// 	upFile, err := os.Open("./dog.png")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	defer upFile.Close()

// 	upFileInfo, _ := upFile.Stat()
// 	var fileSize int64 = upFileInfo.Size()
// 	fileBuffer := make([]byte, fileSize)
// 	upFile.Read(fileBuffer)
// 	log.Println("uploading", len(ImageBytes))

// 	upInput := &s3manager.UploadInput{
// 		Bucket:      aws.String("sagemaker-us-west-2-887392381071"), // bucket's name
// 		Key:         aws.String("test.png"),                         // files destination location
// 		Body:        bytes.NewReader(fileBuffer),                    // content of the file
// 		ContentType: aws.String("image/png"),
// 		ACL:         aws.String("public-read"), // content type
// 	}
// 	res, err := uploader.Upload(upInput)
// 	log.Printf("res %+v\n", res)
// 	log.Printf("err %+v\n", err)
// }

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	// "github.com/aws/aws-sdk-go/aws/credentials"
)

const (
	AWS_S3_REGION = ""
	AWS_S3_BUCKET = ""
)

func UploadFile(ImageBytes []byte, key string) error {
	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	// Credentials: credentials.NewStaticCredentials("KeyID", "SecretKey", ""),

	outPut, err := s3.New(session).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(fmt.Sprintf("images/%s", key)),
		ACL:    aws.String("public-read"),
		Body:   bytes.NewReader(ImageBytes),
		// ContentLength:        aws.Int64(fileSize),
		// ContentType:          aws.String(http.DetectContentType(fileBuffer)),
		// ContentDisposition:   aws.String("attachment"),
		// ServerSideEncryption: aws.String("AES256"),
	})
	fmt.Println(outPut)
	return err
}
