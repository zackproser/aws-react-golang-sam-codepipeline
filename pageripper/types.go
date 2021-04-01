package main

import "net/url"

// Represents a processing request from the client
type ripRequest struct {
	Target    string `json:"target"`
	ParsedURL *url.URL
}

// Represents a json error response
type ripErrorResponse struct {
	Message string `json:"message"`
}

// A scraped links response from the backend
type ripResponse struct {
	Links    []string       `json:"links"`
	Hosts    map[string]int `json:"hostnames"`
	RipCount int            `json:"ripcount"`
}

type countResponse struct {
	Count int `json:"count"`
}

type PagesRippedStats struct {
	Count int64
}
