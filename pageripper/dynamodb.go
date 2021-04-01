package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	tableName   = "pageripper"
	dynamoDBSvc *dynamodb.DynamoDB
)

func init() {

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	dynamoDBSvc = dynamodb.New(sess)
}

type RipCountKey struct {
	URL string `json:"url"`
}

type RipCountInc struct {
	Increment int `json:":c"`
}

func updateRipCount() error {

	key, err := dynamodbattribute.MarshalMap(RipCountKey{
		URL: "system",
	})

	if err != nil {
		return err
	}

	increment, err := dynamodbattribute.MarshalMap(RipCountInc{
		Increment: 1,
	})

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       key,
		ReturnValues:              aws.String("UPDATED_NEW"),
		ExpressionAttributeValues: increment,
		UpdateExpression:          aws.String("set c = c + :c"),
	}

	fmt.Printf("input: %+v\n", input)

	_, err = dynamoDBSvc.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

func readRipCount() int {

	type Item struct {
		C int
	}

	result, err := dynamoDBSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"url": {
				S: aws.String("system"),
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling GetItem: %s", err)
	}

	if result.Item == nil {
		log.Fatalf("Could not find item")
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		log.Fatalf("Could not unmarshal DynamoDB item map to item struct")
	}

	fmt.Printf("Found item: %+v\n", item.C)

	return item.C
}
