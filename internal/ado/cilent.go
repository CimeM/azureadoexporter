package ado

import (
	"os"
	"log"

	"github.com/cimem/azureadoexporter/internal/azureadocomms"
)

type Client struct {
	creds azureadocomms.ADOCredentials
}

func NewClient() *Client {
	return &Client{
		creds: azureadocomms.ADOCredentials{
			URL:          os.Getenv("ADO_URL"),
			Project:      os.Getenv("ADO_PROJECT"),
			Organization: os.Getenv("ADO_ORGANIZATION"),
			PAT:          os.Getenv("ADO_PERSONAL_ACCESS_TOKEN"),
		},
	}
}

func (c *Client) FetchMetrics() ([]string, error) {
	// Fetch pipeline metrics
	pipelineMetrics, err := azureadocomms.Call(c.creds)
	if err != nil {
		log.Printf("Error fetching pipeline metrics: %v", err)
		return nil, err
	}

	// Fetch build metrics
	buildMetrics, err := azureadocomms.FetchBuilds(c.creds)
	if err != nil {
		log.Printf("Error fetching build metrics: %v", err)
		// Continue even if build metrics fail, so we still return pipeline metrics
	}

	// Combine both metrics
	allMetrics := append(pipelineMetrics, buildMetrics...)
	
	return allMetrics, nil
}

// FetchPipelineMetrics fetches only pipeline metrics
func (c *Client) FetchPipelineMetrics() ([]string, error) {
	return azureadocomms.Call(c.creds)
}

// FetchBuildMetrics fetches only build metrics
func (c *Client) FetchBuildMetrics() ([]string, error) {
	return azureadocomms.FetchBuilds(c.creds)
}
