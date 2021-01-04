package main

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"

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

	devHeaders := make(map[string]string)
	devHeaders["Access-Control-Allow-Origin"] = "*"
	devHeaders["Access-Control-Allow-Methods"] = "DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT"
	devHeaders["Access-Control-Allow-Headers"] = "*"
	devHeaders["X-Test-End-To-End-CICD"] = "True"

	if strings.ToUpper(request.HTTPMethod) == "OPTIONS" {
		log.Debug("RESPONDED TO OPTIONS")
		return events.APIGatewayProxyResponse{
			Body:       "Responded to OPTIONS",
			Headers:    devHeaders,
			StatusCode: 200,
		}, nil
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
	parsedURL, parseErr := url.ParseRequestURI(rr.Target)

	if parseErr != nil {
		log.WithFields(logrus.Fields{
			"Error": parseErr,
		}).Debug("Error parsing target URL")

		return events.APIGatewayProxyResponse{
			Body:       formatErrorMessage(""),
			StatusCode: 422,
		}, nil
	}

	if !parsedURL.IsAbs() {
		return events.APIGatewayProxyResponse{
			Body: formatErrorMessage("Relative URLs such as /example.html are not supported. Please supply a full qualified URL such as https://www.example.com"),

			StatusCode: 422,
		}, nil
	}

	// User-supplied URL passed sanity checks, so begin processing it
	rr.ParsedURL = parsedURL

	// Create 3 channels for each request:
	// 1. chUrls handles links as they are found in the given page
	// 2. chHosts handles link hostnames
	// 3. chRipFinished serves to indicate when processing of all links is complete
	chLinks := make(chan string)
	chHosts := make(chan string)
	chRipFinished := make(chan bool)

	go rip(parsedURL, chLinks, chHosts, chRipFinished)

	foundLinks := []string{}
	foundHosts := []string{}

	// Listen on channels for links and done status
	for {
		select {
		case link := <-chLinks:
			foundLinks = append(foundLinks, link)
		case host := <-chHosts:
			foundHosts = append(foundHosts, host)
		case <-chRipFinished:
			log.WithFields(logrus.Fields{
				"Target URL": rr.ParsedURL,
			}).Debug("Rip finished")

			r := ripResponse{
				Links: foundLinks,
				Hosts: tallyCounts(foundHosts),
			}
			j, marshalErr := json.Marshal(&r)
			if marshalErr != nil {
				log.WithFields(logrus.Fields{
					"Error": marshalErr,
				}).Debug("Error marshaling response to JSON")
			}

			// Store metrics on app usage in DynamoDB
			// incrementCountOfPagesRipped(rr, r)

			// Processing successful - return response with links and counts
			return events.APIGatewayProxyResponse{
				Headers:    devHeaders,
				Body:       string(j),
				StatusCode: 200,
			}, nil
		}
	}
}

func main() {
	lambda.Start(handler)
}
