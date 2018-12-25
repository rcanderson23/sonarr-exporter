GOCMD=go
GOBUILD=$(GOCMD) build
GOGET=$(GOCMD) get
BIN_NAME=sonarr_exporter

build:
	$(GOBUILD) -o $(BIN_NAME) -v

standalone:
	CGO_ENABLED=0 $(GOBUILD) -o $(BIN_NAME) -v
deps:
	$(GOGET) github.com/prometheus/client_golang/prometheus
	$(GOGET) github.com/prometheus/client_golang/prometheus/promhttp

