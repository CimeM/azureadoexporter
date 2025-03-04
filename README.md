# Azure DevOps exporter
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/CimeM/azureadoexporter/build.yml?branch=main) ![License](https://img.shields.io/github/license/CimeM/azureadoexporter) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/CimeM/azureadoexporter) [![Go Report Card](https://goreportcard.com/badge/github.com/cimem/azureadoexporter)](https://goreportcard.com/report/github.com/cimem/azureadoexporter)

The Azure DevOps Pipeline Telemetry Exporter is a Prometheus exporter designed to monitor and expose metrics from Azure DevOps pipelines. It functions similarly to the Node Exporter, providing valuable insights into your pipeline performance and health.

## Features
- Collects metrics from Azure DevOps pipelines
- Exposes metrics in Prometheus format
- Supports multiple pipeline queues
- Configurable scrape interval
- Easy integration with existing Prometheus setups
- Detailed build metrics and agent utilization tracking

## Quick Start with Docker
The easiest way to use the Azure DevOps Pipeline Telemetry Exporter is through our Docker image:

1. Pull the latest image:
``` bash 
docker pull cimem/azureadoexporter:latest
```

2. Run the container:
``` bash 
docker run --rm -p 8080:8080 \
-e ADO_ORGANIZATION=your-organization \
-e ADO_PROJECT=your-project \
-e ADO_PERSONAL_ACCESS_TOKEN=your-personal-access-token \
-e ADO_URL=your-ado-url \
cimem/azureadoexporter:latest
```

3. Access metrics at `http://localhost:8080/metrics`

## Configuration
The exporter can be configured using environment variables:
- `ADO_ORGANIZATION`: Your Azure DevOps organization name
- `ADO_PROJECT`: Your Azure DevOps project name
- `ADO_PERSONAL_ACCESS_TOKEN`: Your Personal Access Token
- `ADO_URL`: URL to your Azure DevOps site
- `SCRAPE_INTERVAL`: Scrape interval in minutes (default: 5)

## Metrics and Queries
The exporter provides several metric collections:

### Pipeline Metrics
- `azure_devops_pipeline` - Basic pipeline information
- `azure_devops_pipeline_run` - Pipeline run information

### Build API Metrics
- `azure_devops_build` - Build information including definition name, repository, source branch, requester, and result
- `azure_devops_build_duration_seconds` - Duration of builds
- `azure_devops_build_queue_duration_seconds` - Time spent in queue before build starts
- `azure_devops_job_duration_seconds` - Duration of individual jobs within builds
- `azure_devops_task_duration_seconds` - Duration of individual tasks within jobs
- `azure_devops_build_worker` - Information on which workers are occupied by which builds

### Example PromQL Queries

Pipeline failure rate:
``` promql
( count( azure_devops_pipeline_run{result="failed"} ) / (count( azure_devops_pipeline_run{result="failed"}) + count(azure_devops_pipeline_run{result="succeeded"}) ) ) * 100
```

Number of all pipelines in the project:
``` promql
sum( azure_devops_pipeline )
```

Pipelines that ran least amount of times:
``` promql
count( azure_devops_pipeline_run ) by (pipelinename) <= 3
```

Histogram of pipeline runs (use Histogram as graph type):
``` promql
count(azure_devops_pipeline_run) by (pipelinename)
```

Build success rate by repository (bar chart):
``` promql
(count by (repository) (azure_devops_build{result="succeeded", status="completed"})) 
/ 
(count by (repository) (azure_devops_build{status="completed"})) * 100
```

Agent workload distribution (pie chart):
``` promql
sum by (workerName) (azure_devops_build_worker)
```

Build distribution by reason (manual, CI, pull request, scheduled):
``` promql
count by (reason) (azure_devops_build)
```

## Development
Prerequisites:
- Go compiler
- make
- Azure DevOps account with appropriate permissions
- ADO PersonalAccessToken

Building:
``` bash
docker build -t azureadoexporter:latest .
```

## Contributing
Contributions are welcome! Please feel free to submit an Issue or a Pull Request.

## License
This project is licensed under the MIT License.
