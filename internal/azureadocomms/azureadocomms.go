package azureadocomms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ADOCredentials struct {
	URL          string
	Project      string
	Organization string
	PAT          string
}

type PipelineResponse struct {
	Count int `json:"count"`
	Value []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Revision int    `json:"revision"`
	} `json:"value"`
}

type PipelineRunResponse struct {
	Count int `json:"count"`
	Value []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Url          string `json:"url"`
		State        string `json:"state"`
		Result       string `json:"result"`
		FinishedDate string `json:"finishedDate"`
		CreatedDate  string `json:"createdDate"`
		Pipeline     struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"pipeline"`
	} `json:"value"`
}

func calculateDuration(createdDate, finishedDate string) (time.Duration, error) {
	layout := "2006-01-02T15:04:05.9999999Z"

	t1, err := time.Parse(layout, createdDate)
	if err != nil {
		return 0, fmt.Errorf("error parsing createdDate: %v", err)
	}

	t2, err := time.Parse(layout, finishedDate)
	if err != nil {
		// -1 is a sign that its not finished
		return -1, nil //, fmt.Errorf("error parsing finishedDate: %v", err)
	}

	duration := t2.Sub(t1)
	return duration, nil
}

// exported name Call, calls the ADO using api
func Call(ado_cred ADOCredentials) ([]string, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines?api-version=7.0",
		ado_cred.URL,
		ado_cred.Organization,
		ado_cred.Project,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.SetBasicAuth("", ado_cred.PAT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return nil, err
	}

	var pipelineResp PipelineResponse
	err = json.Unmarshal(body, &pipelineResp)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return nil, err
	}

	metricsData := []string{}
	// Process the pipeline information for Prometheus
	for _, pipeline := range pipelineResp.Value {
		// describe every run of the pipeline
		runs, err := GetPipelineRuns(ado_cred, strconv.Itoa(pipeline.ID), pipeline.Name)
		if err != nil {
			log.Printf("Error parsing JSON: %s", err)
			return nil, err
		}
		// append runs metrics to the final output
		metricsData = append(metricsData, runs...)

		// insert pipeline data point
		metric := fmt.Sprintf(
			"azure_devops_pipeline{id=\"%d\",name=\"%s\",revision=\"%d\",runcount=\"%d\"} 1\n",
			pipeline.ID,
			pipeline.Name,
			pipeline.Revision,
			len(runs),
		)
		metricsData = append(metricsData, metric)
	}
	return metricsData, nil
}

func GetPipelineRuns(ado_cred ADOCredentials, pipelineID string, pipelineName string) ([]string, error) {
	url := fmt.Sprintf(
		"%s/%s/%s/_apis/pipelines/%s/runs?api-version=7.0",
		ado_cred.URL,
		ado_cred.Organization,
		ado_cred.Project,
		pipelineID,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.SetBasicAuth("", ado_cred.PAT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return nil, err
	}

	var pipelineRunResp PipelineRunResponse
	err = json.Unmarshal(body, &pipelineRunResp)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return nil, err
	}

	var metricsData []string
	// Process the pipeline information for Prometheus
	for _, pipelnerun := range pipelineRunResp.Value {

		// calculate duration
		duration, err := calculateDuration(pipelnerun.CreatedDate, pipelnerun.FinishedDate)
		if err != nil {
			log.Printf("Error: %v", err)
			duration = 0
		}

		metric := fmt.Sprintf(
			"azure_devops_pipeline_run{name=\"%s\",result=\"%s\",durationinseconds=\"%.2f\",state=\"%s\",finishedDate=\"%s\",id=\"%d\", pipelineid=\"%d\", pipelinename=\"%s\"} 1",
			pipelnerun.Name,
			pipelnerun.Result,
			duration.Seconds(),
			pipelnerun.State,
			pipelnerun.FinishedDate,
			pipelnerun.ID,
			pipelnerun.Pipeline.ID,
			pipelineName,
		)

		metricsData = append(metricsData, metric)
	}
	return metricsData, nil
}
