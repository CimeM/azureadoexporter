package azureadocomms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type BuildResponse struct {
	Count int           `json:"count"`
	Value []BuildDetail `json:"value"`
}

type BuildDetail struct {
	ID            int               `json:"id"`
	BuildNumber   string            `json:"buildNumber"`
	Status        string            `json:"status"`
	Result        string            `json:"result"`
	QueueTime     string            `json:"queueTime"`
	StartTime     string            `json:"startTime"`
	FinishTime    string            `json:"finishTime"`
	SourceBranch  string            `json:"sourceBranch"`
	SourceVersion string            `json:"sourceVersion"`
	TriggerInfo   map[string]string `json:"triggerInfo"`
	Reason        string            `json:"reason"`
	Priority      string            `json:"priority"`
	Definition    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"definition"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Repository struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"repository"`
	RequestedBy struct {
		DisplayName string `json:"displayName"`
	} `json:"requestedBy"`
	Queue struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"queue"`
}

// FetchBuilds retrieves the latest 1000 builds from Azure DevOps
func FetchBuilds(ado_cred ADOCredentials) ([]string, error) {
	// Request the latest 1000 builds, ordered by most recent first
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds?api-version=7.0&queryOrder=queueTimeDescending", 
		ado_cred.URL,
		ado_cred.Organization,
		ado_cred.Project,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for builds: %v", err)
		return nil, err
	}

	req.SetBasicAuth("", ado_cred.PAT)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending build request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response from API: %s", resp.Status)
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading build response: %v", err)
		return nil, err
	}

	var buildResp BuildResponse
	err = json.Unmarshal(body, &buildResp)
	if err != nil {
		log.Printf("Error parsing build JSON: %v", err)
		return nil, err
	}

	return generateMetrics(buildResp.Value), nil
}

// generateMetrics processes builds and returns Prometheus-formatted metrics
func generateMetrics(builds []BuildDetail) []string {
	metricsData := []string{}
	
	// Create maps to track unique pools, definitions, and projects for summary metrics
	buildPools := make(map[int]string)
	buildDefinitions := make(map[int]string)
	buildProjects := make(map[string]string)
	
	// Process each build
	for _, build := range builds {
		// Store info for summary metrics
		buildPools[build.Queue.ID] = build.Queue.Name
		buildDefinitions[build.Definition.ID] = build.Definition.Name
		buildProjects[build.Project.ID] = build.Project.Name
		
		queueDuration, queueErr := calculateDuration(build.QueueTime, build.StartTime)
		if queueErr != nil {
		    log.Printf("Warning: error calculating queue duration: %v", queueErr)
		    queueDuration = 0
		}

		buildDuration, buildErr := calculateDuration(build.StartTime, build.FinishTime)
		if buildErr != nil {
		    log.Printf("Warning: error calculating build duration: %v", buildErr)
		    buildDuration = 0
		}		
		// Get triggerSourceBranch if it exists
		triggerSourceBranch := ""
		if val, ok := build.TriggerInfo["ci.sourceBranch"]; ok {
			triggerSourceBranch = val
		}

		// Main build metric with labels
		metric := fmt.Sprintf(
			"azure_devops_build{id=\"%d\",buildNumber=\"%s\",status=\"%s\",result=\"%s\",definitionId=\"%d\",definitionName=\"%s\",project=\"%s\",repository=\"%s\",repoType=\"%s\",requestedBy=\"%s\",sourceBranch=\"%s\",priority=\"%s\",pool=\"%s\",reason=\"%s\",triggerSourceBranch=\"%s\"} 1",
			build.ID,
			build.BuildNumber,
			build.Status,
			build.Result,
			build.Definition.ID,
			build.Definition.Name,
			build.Project.Name,
			build.Repository.Name,
			build.Repository.Type,
			build.RequestedBy.DisplayName,
			build.SourceBranch,
			build.Priority,
			build.Queue.Name,
			build.Reason,
			triggerSourceBranch,
		)
		metricsData = append(metricsData, metric)

		// Queue duration metric (only if valid)
		if queueDuration > 0 {
			queueMetric := fmt.Sprintf(
				"azure_devops_build_queue_duration_seconds{id=\"%d\",buildNumber=\"%s\",definitionName=\"%s\",project=\"%s\"} %.2f",
				build.ID,
				build.BuildNumber,
				build.Definition.Name,
				build.Project.Name,
				queueDuration.Seconds(),
			)
			metricsData = append(metricsData, queueMetric)
		}

		// Build duration metric (only if valid and status is completed)
		if buildDuration > 0 && build.Status == "completed" {
			buildMetric := fmt.Sprintf(
				"azure_devops_build_duration_seconds{id=\"%d\",buildNumber=\"%s\",definitionName=\"%s\",project=\"%s\",result=\"%s\"} %.2f",
				build.ID,
				build.BuildNumber,
				build.Definition.Name,
				build.Project.Name,
				build.Result,
				buildDuration.Seconds(),
			)
			metricsData = append(metricsData, buildMetric)
		}
	}

	return metricsData
}

// countBuildsByStatus counts builds in each status category
func countBuildsByStatus(builds []BuildDetail) map[string]int {
	counts := make(map[string]int)
	for _, build := range builds {
		counts[build.Status]++
	}
	return counts
}
