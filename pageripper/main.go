package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func formatErrorMessage(optionalMessage string) string {
	var targetMessage string
	var genericErrorMessage = "Please submit a valid, absolute URL such as https://example.com"

	if optionalMessage != "" {
		targetMessage = optionalMessage
	} else {
		targetMessage = genericErrorMessage
	}

	errorResponse := ripErrorResponse{Message: targetMessage}
	errJson, marshalErr := json.Marshal(errorResponse)
	if marshalErr != nil {
		log.WithFields(logrus.Fields{
			"Error": marshalErr,
		}).Debug("Unable to marshal error message to JSON")

		return targetMessage
	}
	return string(errJson)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	respBody := ripResponse{
		Links: []string{"https://wakka.com", "https://wikka.com", "https://tambourine.com"},
	}

	var rr ripRequest

	// Ensure the request contained a valid URL target
	unmarshalErr := json.Unmarshal([]byte(request.Body), &rr)

	if unmarshalErr != nil {
		log.WithFields(logrus.Fields{
			"Error": unmarshalErr,
		}).Debug("Invalid request - must be JSON with field named target containing a valid URL")
	}

	if rr.Target == "" {
		// Request did not include a target URL
		return events.APIGatewayProxyResponse{
			Body:       formatErrorMessage(""),
			StatusCode: 422,
		}, nil
	}

	// Ensure the supplied target is a valid URL
	parsedUrl, parseErr := url.ParseRequestURI(rr.Target)

	if parseErr != nil {
		log.WithFields(logrus.Fields{
			"Error": parseErr,
		}).Debug("Error parsing target URL")

		return events.APIGatewayProxyResponse{
			Body:       formatErrorMessage(""),
			StatusCode: 422,
		}, nil
	}

	if !parsedUrl.IsAbs() {
		return events.APIGatewayProxyResponse{
			Body: formatErrorMessage("Relative URLs such as /example.html are not supported. Please supply a full qualified URL such as https://www.example.com"),

			StatusCode: 422,
		}, nil
	}

	// Dummy success response
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("links:%s", respBody.Links),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
