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

ENV ADO_ORGANIZATION=fabrikam
ENV ADO_PERSONAL_ACCESS_TOKEN=mypat
ENV ADO_PROJECT=fabrikam-fiber-tfvc
ENV ADO_URL=localhost

EXPOSE 8080

# run
CMD ["/azureadoexporter"]