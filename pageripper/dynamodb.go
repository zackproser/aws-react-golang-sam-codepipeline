package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	tableName   = "pageripper"
	dynamoDBSvc *dynamodb.DynamoDB
)

// init is run the first time this package is instantiated, so it's a good place to do the DynamoDB service setup
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

// updateRipCount increments the special "system" entry for tracking the overall number of jobs the app has run
func updateRipCount() {

	key, err := dynamodbattribute.MarshalMap(RipCountKey{
		URL: "system",
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error marshalling rip count increment to map")
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

	_, err = dynamoDBSvc.UpdateItem(input)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error updating rip count in DynamoDB")
	}
}

// readRipCount retrieves the value of total jobs processed by the app, stored in the DynamoDB system key
func readRipCount(chRipCount chan<- int, chCountRetrieved chan<- bool) {

	defer func() {
		chCountRetrieved <- true
	}()

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
		log.Debugf("Got error calling GetItem: %s", err)
	}

	if result.Item == nil {
		log.Debug("Could not find item")
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		log.Fatalf("Could not unmarshal DynamoDB item map to item struct")
	}

	if item.C == 0 {
		log.Debug("Retrieved rip count from DynamoDB was 0")
		chCountRetrieved <- true
		return
	}

	log.WithFields(logrus.Fields{
		"Rip Count": item.C,
	}).Debug("Successfully retrieved ripCount from DynamoDB")

	chRipCount <- item.C
}
