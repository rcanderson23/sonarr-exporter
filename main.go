package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	apiKey    string
	sonarrUrl string
)

func getJson(url string, apiKey string, target interface{}) error {
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

type SonarrCollector struct {
	systemStatus     *prometheus.Desc
	historyRecords   *prometheus.Desc
	wantedRecords    *prometheus.Desc
	queueRecords     *prometheus.Desc
	folderProperties *prometheus.Desc
}

func newSonarrCollector() *SonarrCollector {
	return &SonarrCollector{
		systemStatus:     prometheus.NewDesc("sonarr_status", "System Status", []string{"version", "appData", "branch"}, nil),
		historyRecords:   prometheus.NewDesc("sonarr_history_total_records", "Total records in Sonarr histor", nil, nil),
		wantedRecords:    prometheus.NewDesc("sonarr_missing_episodes", "Total missing episodes in Sonarr", nil, nil),
		queueRecords:     prometheus.NewDesc("sonarr_queue_total_records", "Total records in Sonarr queue", nil, nil),
		folderProperties: prometheus.NewDesc("sonarr_root_folder_space", "Root folder space in Sonarr", []string{"path"}, nil),
	}
}

func (c *SonarrCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.systemStatus
	ch <- c.historyRecords
	ch <- c.wantedRecords
	ch <- c.queueRecords
	ch <- c.folderProperties
}

func (c *SonarrCollector) Collect(ch chan<- prometheus.Metric) {

	status := SystemStatus{}
	sonarrStatus := 1.0
	getJson(sonarrUrl+"/system/status", apiKey, &status)
	if (SystemStatus{}) == status {
		sonarrStatus = 0.0
	}
	ch <- prometheus.MustNewConstMetric(c.systemStatus, prometheus.CounterValue, sonarrStatus, status.Version, status.AppData, status.Branch)

	history := History{}
	getJson(sonarrUrl+"/history", apiKey, &history)
	ch <- prometheus.MustNewConstMetric(c.historyRecords, prometheus.CounterValue, float64(history.TotalRecords))

	wanted := WantedMissing{}
	getJson(sonarrUrl+"/wanted/missing", apiKey, &wanted)
	ch <- prometheus.MustNewConstMetric(c.wantedRecords, prometheus.CounterValue, float64(wanted.TotalRecords))

	queue := Queue{}
	getJson(sonarrUrl+"/queue", apiKey, &queue)
	for _, titles := range queue {
		ch <- prometheus.MustNewConstMetric(c.queueRecords, prometheus.CounterValue, float64(titles.Size), titles.Title)
	}

	folders := RootFolder{}
	getJson(sonarrUrl+"/rootfolder", apiKey, &folders)
	for _, folder := range folders {
		ch <- prometheus.MustNewConstMetric(c.folderProperties, prometheus.CounterValue, float64(folder.FreeSpace), folder.Path)
	}
}

type RootFolder []struct {
	Path      string `json:"path"`
	FreeSpace int64  `json:"freeSpace"`
}

type SystemStatus struct {
	Version string `json:"version"`
	AppData string `json:"appData"`
	Branch  string `json:"branch"`
}

type Queue []struct {
	Title string `json:"title"`
	Size  int32  `json:"size"`
}

type History struct {
	TotalRecords int `json:"totalRecords"`
}

type WantedMissing struct {
	TotalRecords int `json:"totalRecords"`
}

type Configuration struct {
	APIKey    string `json:"apiKey"`
	SonarrURL string `json:"sonarrUrl"`
}

func main() {
	config := Configuration{}
	configFilePtr := flag.String("configFile", "config.json", "path to json config file")
	flag.Parse()
	file, err := os.Open(*configFilePtr)
	if err != nil {
		fmt.Println("Failed to open config file")
		os.Exit(3)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Failed to parse config file")
		os.Exit(3)
	}

	apiKey = config.APIKey
	sonarrUrl = config.SonarrURL
	Sonarr := newSonarrCollector()
	prometheus.MustRegister(Sonarr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Sonarr Exporter</title></head>
			<body>
			<h1>Sonarr Exporter</h1>
			<p><a href="` + "metrics" + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Exporter listening on :9715/metrics")
	log.Fatal(http.ListenAndServe(":9715", nil))
}
