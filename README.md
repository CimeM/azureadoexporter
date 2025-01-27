# Azure DevOps exporter


The Azure DevOps Pipeline Telemetry Exporter is a Prometheus exporter designed to monitor and expose metrics from Azure DevOps pipelines. It functions similarly to the Node Exporter, providing valuable insights into your pipeline performance and health.

## Features

- Collects metrics from Azure DevOps pipelines
- Exposes metrics in Prometheus format
- Supports multiple pipeline queues
- Configurable scrape interval
- Easy integration with existing Prometheus setups

## Quick Start with Docker

The easiest way to use the Azure DevOps Pipeline Telemetry Exporter is through our Docker image:

1. Pull the latest image:

``` bash 
docker pull cimartindev/azureadoexporter:latest
```

2. Run the container:
``` bash 
docker run --rm -p 8080:8080 \
-e ADO_ORGANIZATION=your-organization \
-e ADO_PROJECT=your-project \
-e ADO_PERSONAL_ACCESS_TOKEN=your-personal-access-token \
-e ADO_URL=your-ado-url \
cimartindev/azureadoexporter:latest
```

3. Access metrics at `http://localhost:8080/metrics`

## Configuration

The exporter can be configured using environment variables:

- `ADO_ORGANIZATION`: Your Azure DevOps organization name
- `ADO_PROJECT`: Your Azure DevOps project name
- `ADO_PERSONAL_ACCESS_TOKEN`: Your Personal Access Token
- `ADO_URL`: URL to your Azure DevOps site
- `SCRAPE_INTERVAL`: Scrape interval in minutes (default: 5)

## Development

Prerequisites:
- Go compiler
- Azure DevOps account with appropriate permissions
- ADO PersonalAccessToken

Building:

``` bash
docker build -t azureadoexporter:latest .
```

## Contributing

Contributions are welcome! Please feel free to submit Issue or a Pull Request.

## License

This project is licensed under the MIT License.