# syntax=docker/dockerfile:1
# Azure DevOps Pipeline exporter written in go language
FROM golang:1.23-bookworm

WORKDIR /app

# install the modules necessary to compile it
COPY src/ ./

WORKDIR /app/azureadoexporter
# install dependencies 
RUN go mod download
RUN go mod tidy

# compile the go application 
RUN CGO_ENABLED=0 GOOS=linux go build -o /azureadoexporter

ENV ADO_ORGANIZATION=${ADO_ORGANIZATION}
ENV ADO_PERSONAL_ACCESS_TOKEN=${PAT}
ENV ADO_PROJECT=${ADO_PROJECT}
ENV ADO_URL=${ADO_URL}
EXPOSE 8080

# run
CMD ["/azureadoexporter"]