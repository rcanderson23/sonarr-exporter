sonarr-exporter
===================

Prometheus exporter for Sonarr.

Building the binary
===================

For running in a Docker image, run `make standalone`

Configuration
===================

Move `config.example.json` to `config.json` and supply your API key and the URL for sonarr.

Exporter listens on port `9175`

