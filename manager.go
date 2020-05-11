package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/guregu/dynamo"
	"github.com/zono-dev/stplib"
)

var s3svc *(s3.S3)  // s3 session object
var db *(dynamo.DB) // DynamoDB's db object

// CreateSession returns dynamo.Table object.
// dtable is the name of DynamoDB table.
func CreateSession(dtable string) dynamo.Table {
	db = dynamo.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
	table := db.Table(dtable)
	return table
}

// GetItems returns all items in table.
// dtable is the name of DynamoDB table.
func GetItems(dtable string) ([]stplib.ImgInfo, error) {
	table := CreateSession(dtable)
	var res []stplib.ImgInfo
	err := table.Scan().All(&res)
	return res, err
}

// SearchItem returns a ImgInfo matches FileName with fname.
func SearchItem(items []stplib.ImgInfo, fname string) *stplib.ImgInfo {
	for i, v := range items {
		if v.FileName == fname {
			return &(items[i])
		}
	}
	return nil
}

// Note: This function is just example for a DynamoDB table which uses Index Key and Range Key.
//func DeleteItem(dtable string, fname string, createdat string) error {
//	table := CreateSession(dtable)
//	//err := table.Delete("FileName", fname).Run()
//	dels := table.Delete("FileName", fname).Range("CreatedAt", createdat)
//	fmt.Printf("%#v\n", dels)
//	err := dels.Run()
//	if err != nil {
//		fmt.Println(err)
//	}
//	return err
//}

// DeleteItem deletes item has FileName matches fname in table.
func DeleteItem(dtable string, fname string) error {
	table := CreateSession(dtable)
	err := table.Delete("FileName", fname).Run()
	return err
}

// PutItem puts item in table.
func PutItem(dtable string, dat stplib.ImgInfo) error {
	table := CreateSession(dtable)
	err := table.Put(dat).Run()
	return err
}

// PutItem puts item in table.
func PutItems(dtable string, dat []stplib.ImgInfo) {
	table := CreateSession(dtable)
    for _, v:= range dat {
	    err := table.Put(v).Run()
        if err != nil {
            fmt.Println(err)
        }
    }
}

// NewS3Sess creates session and returns s3 session.
func NewS3Sess(region string) *s3.S3 {
	conf := aws.Config{
		Region: aws.String(region),
	}
	sess := session.Must(session.NewSession())
	s3svc = s3.New(sess, &conf)
	return s3svc
}

// SetObjs sets file(s) for s3.ObjectIdentifier.
func SetObjs(files []string) []*s3.ObjectIdentifier {
	objs := []*s3.ObjectIdentifier{}
	for _, v := range files {
		objs = append(objs, &s3.ObjectIdentifier{
			Key: aws.String(v),
		})
	}
	return objs
}

// DelObjS3 delete file(s) in a bucket.
func DelObjS3(bucket string, files []string) error {
	objs := SetObjs(files)
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: objs,
		},
	}

	res, err := s3svc.DeleteObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}
	fmt.Println("Files have been deleted.")
	fmt.Println(res)
	return nil
}
