package azureadocomms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
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

// TimelineResponse represents the Azure DevOps build timeline API response
type TimelineResponse struct {
	Records []TimelineRecord `json:"records"`
}

// TimelineRecord represents a record in the build timeline
type TimelineRecord struct {
    ID           string     `json:"id"`
    ParentID     string     `json:"parentId"`
    Type         string     `json:"type"`
    Name         string     `json:"name"`
    StartTime    string     `json:"startTime"`
    FinishTime   string     `json:"finishTime"`
    Result       string     `json:"result"`
    WorkerName   string     `json:"workerName"`
    LogID        LogInfo    `json:"log"`  
    Order        int        `json:"order"`
    State        string     `json:"state"`
    ErrorCount   int        `json:"errorCount"`
    WarningCount int        `json:"warningCount"`
}

// LogInfo represents the log information in a timeline record
type LogInfo struct {
    ID   int    `json:"id"`
    Type string `json:"type"`
    URL  string `json:"url"`
}
// FetchBuildTimeline retrieves the timeline for a specific build
func FetchBuildTimeline(ado_cred ADOCredentials, buildID int) (TimelineResponse, error) {
	var timeline TimelineResponse
	
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/timeline?api-version=7.0",
		ado_cred.URL,
		ado_cred.Organization,
		ado_cred.Project,
		buildID,
	)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return timeline, err
	}
	
	req.SetBasicAuth("", ado_cred.PAT)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return timeline, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return timeline, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return timeline, err
	}
	
	err = json.Unmarshal(body, &timeline)
	if err != nil {
		return timeline, err
	}
	
	return timeline, nil
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

	// Fetch timeline data for completed builds and generate enhanced metrics
	buildMetrics, err := generateMetrics(buildResp.Value)
	if err != nil {
		return nil, err
	}
	
	timelineMetrics, err := generateTimelineMetrics(buildResp.Value, ado_cred, 100) // Limit to 100 most recent builds
	if err != nil {
		log.Printf("Warning: error generating timeline metrics: %v", err)
		// Continue with build metrics even if timeline metrics fail
		return buildMetrics, nil
	}
	
	// Combine both metrics
	allMetrics := append(buildMetrics, timelineMetrics...)
	return allMetrics, nil
}

// generateMetrics processes builds and returns Prometheus-formatted metrics
func generateMetrics(builds []BuildDetail) ([]string, error) {
	metricsData := []string{}
	
	// Process each build
	for _, build := range builds {
		// Get Time spent in Queue
		queueDuration, queueErr := calculateDuration(build.QueueTime, build.StartTime)
		if queueErr != nil {
		    log.Printf("Warning: error calculating queue duration: %v", queueErr)
		    queueDuration = 0
		}
		// Get Time spent building
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

		// Main build metric with simplified labels (removed instance, job, buildNumber, definition, priority)
		metric := fmt.Sprintf(
			"azure_devops_build{id=\"%d\",status=\"%s\",result=\"%s\",definitionName=\"%s\",project=\"%s\",repository=\"%s\",repoType=\"%s\",requestedBy=\"%s\",sourceBranch=\"%s\",pool=\"%s\",reason=\"%s\",triggerSourceBranch=\"%s\"} 1",
			build.ID,
			build.Status,
			build.Result,
			build.Definition.Name,
			build.Project.Name,
			build.Repository.Name,
			build.Repository.Type,
			build.RequestedBy.DisplayName,
			build.SourceBranch,
			build.Queue.Name,
			build.Reason,
			triggerSourceBranch,
		)
		metricsData = append(metricsData, metric)

		// Queue duration metric (only if valid)
		if queueDuration > 0 {
			queueMetric := fmt.Sprintf(
				"azure_devops_build_queue_duration_seconds{id=\"%d\",definitionName=\"%s\",project=\"%s\"} %.2f",
				build.ID,
				build.Definition.Name,
				build.Project.Name,
				queueDuration.Seconds(),
			)
			metricsData = append(metricsData, queueMetric)
		}

		// Build duration metric (only if valid and status is completed)
		if buildDuration > 0 && build.Status == "completed" {
			buildMetric := fmt.Sprintf(
				"azure_devops_build_duration_seconds{id=\"%d\",definitionName=\"%s\",project=\"%s\",result=\"%s\"} %.2f",
				build.ID,
				build.Definition.Name,
				build.Project.Name,
				build.Result,
				buildDuration.Seconds(),
			)
			metricsData = append(metricsData, buildMetric)
		}
	}

	return metricsData, nil
}

// generateTimelineMetrics fetches and processes timeline data for builds
func generateTimelineMetrics(builds []BuildDetail, ado_cred ADOCredentials, limit int) ([]string, error) {
	var metricsData []string
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Limit the number of builds to process
	buildCount := len(builds)
	if buildCount > limit {
		buildCount = limit
	}

	// Use a semaphore to limit concurrent API calls
	semaphore := make(chan struct{}, 5) // Maximum 5 concurrent requests

	// Process each build
	for i := 0; i < buildCount; i++ {
		build := builds[i]
		// Only process completed builds
		if build.Status == "completed" {
			wg.Add(1)
			go func(build BuildDetail) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Fetch timeline
				timeline, err := FetchBuildTimeline(ado_cred, build.ID)
				if err != nil {
					log.Printf("Error fetching timeline for build %d: %v", build.ID, err)
					return
				}

				// Maps to track data
				buildWorkers := make(map[string]bool)
				jobMap := make(map[string]TimelineRecord) // Map job IDs to job records
				
				// First pass: identify jobs and workers
				for _, record := range timeline.Records {
					if record.Type == "Job" {
						jobMap[record.ID] = record
					}
					
					if record.WorkerName != "" {
						buildWorkers[record.WorkerName] = true
					}
				}
				
				// Create worker-to-build association metrics
				for workerName := range buildWorkers {
					workerMetric := fmt.Sprintf(
						"azure_devops_build_worker{buildId=\"%d\",definitionName=\"%s\",project=\"%s\",workerName=\"%s\"} 1",
						build.ID,
						build.Definition.Name,
						build.Project.Name,
						workerName,
					)
					mutex.Lock()
					metricsData = append(metricsData, workerMetric)
					mutex.Unlock()
				}

				// Second pass: process all tasks and jobs
				for _, record := range timeline.Records {
					// Process jobs
					if record.Type == "Job" {
						jobDuration, err := calculateDuration(record.StartTime, record.FinishTime)
						if err == nil && jobDuration > 0 {
							jobMetric := fmt.Sprintf(
								"azure_devops_job_duration_seconds{buildId=\"%d\",definitionName=\"%s\",jobName=\"%s\",workerName=\"%s\",result=\"%s\"} %.2f",
								build.ID,
								build.Definition.Name,
								record.Name,
								record.WorkerName,
								record.Result,
								jobDuration.Seconds(),
							)
							mutex.Lock()
							metricsData = append(metricsData, jobMetric)
							mutex.Unlock()
						}
					}
					
					// Process tasks
					if record.Type == "Task" && record.WorkerName != "" {
						// Find the parent job if it exists
						var jobName string
						if job, exists := jobMap[record.ParentID]; exists {
							jobName = job.Name
						} else {
							jobName = "unknown"
						}
						
						taskDuration, err := calculateDuration(record.StartTime, record.FinishTime)
						if err == nil && taskDuration > 0 {
							taskMetric := fmt.Sprintf(
								"azure_devops_task_duration_seconds{buildId=\"%d\",definitionName=\"%s\",jobName=\"%s\",taskName=\"%s\",workerName=\"%s\",result=\"%s\"} %.2f",
								build.ID,
								build.Definition.Name,
								jobName,
								record.Name,
								record.WorkerName,
								record.Result,
								taskDuration.Seconds(),
							)
							mutex.Lock()
							metricsData = append(metricsData, taskMetric)
							mutex.Unlock()
						}
					}
				}
			}(build)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return metricsData, nil
}
