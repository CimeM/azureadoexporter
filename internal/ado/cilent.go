package ado

import (
	"os"


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
	return azureadocomms.Call(c.creds)
}
