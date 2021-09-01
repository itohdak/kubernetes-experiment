package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func HttpPost(target_url string, user_count, spawn_rate int) error {
	values := url.Values{}

	// curl "http://localhost:8089/swarm" -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "locust_count=200&hatch_rate=1"
	// values.Add("locust_count", strconv.Itoa(user_count))
	// values.Add("hatch_rate", strconv.Itoa(spawn_rate))

	// curl "http://localhost:8089/swarm" -X POST -H "Content-Type: application/x-www-form-urlencoded; charset=UTF-8" --data "user_count=200&spawn_rate=1"
	values.Set("user_count", strconv.Itoa(user_count))
	values.Set("spawn_rate", strconv.Itoa(spawn_rate))

	req, err := http.NewRequest(
		"POST",
		target_url,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

func getMetrics(prometheusAddress string) bool {
	client, err := api.NewClient(api.Config{
		Address: prometheusAddress,
	})
	if err != nil {
		log.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// // /api/v1/targets
	// // result, err := api.Targets(ctx)
	// // if err != nil {
	// //     log.Printf("Error get Targets: %v\n", err)
	// //     os.Exit(1)
	// // }
	// // log.Printf("Result:\n%v\n", result)

	// // /api/v1/target/metadata
	// result_metric_metadata, err := api.TargetsMetadata(ctx, "{app=\"frontend\"}", "istio_request_duration_milliseconds", "")
	// if err != nil {
	// 	log.Printf("Error get target metadata: %v\n", err)
	// 	os.Exit(1)
	// }
	// log.Printf("Result TargetsMetadata: \n%v\n", result_metric_metadata)

	// /api/v1/query_range
	// query := "istio_requests_total{destination_app=\"frontend\",response_code=\"200\",source_app=\"istio-ingressgateway\",reporter=\"source\"}"
	// query := "istio_request_duration_milliseconds_bucket{reporter=\"source\",destination_service=\"frontend.default.svc.cluster.local\",response_code=\"200\"}"
	query :=
		`(
		histogram_quantile(0.90,
			sum(irate(istio_request_duration_milliseconds_bucket{
				reporter="source",
				destination_service=~"frontend.default.svc.cluster.local"
				}[1m]
		)) by (le)) / 1000) or
		histogram_quantile(0.90,
			sum(irate(istio_request_duration_seconds_bucket{
				reporter="source",
				destination_service=~"frontend.default.svc.cluster.local"
				}[1m]
		)) by (le))`
	now := time.Now()
	// range_param := v1.Range{Start: now.Add(-5 * time.Minute), End: now, Step: 5 * time.Second}
	// result_query_range, warning, err := api.QueryRange(ctx, query, range_param)
	result_query, warning, err := api.Query(ctx, query, now)
	if err != nil {
		log.Printf("Error while querying prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warning) > 0 {
		log.Printf("Warning QueryRange: %v\n", warning)
	}
	// log.Printf("Result QueryRange: \n%v\n", result_query_range)
	// log.Printf("response time (90 percentile): %v\n", result_query)
	// {} => xxx @[xxx.xxx]
	rep := regexp.MustCompile(`\s`)
	result := rep.Split(result_query.String(), -1)
	if result[2] == "NaN" {
		log.Printf("One of the metrics got NaN. Is it loaded successfully?\n")
		return true
	}
	response_time, err := strconv.ParseFloat(result[2], 64)
	if err != nil {
		log.Printf("Error while converting string to float: %v\n", err)
		os.Exit(1)
	}
	log.Printf("response time (90 percentile): %f\n", response_time)
	ok := response_time <= 2.00
	return ok
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {

	var (
		locustHost    = getEnv("LOCUST_HOST", "localhost")
		locustPort    = getEnv("LOCUST_PORT", "8089")
		locustAddress = fmt.Sprintf("http://%s:%s", locustHost, locustPort)

		prometheusHost    = getEnv("PROMETHEUS_HOST", "localhost")
		prometheusPort    = getEnv("PROMETHEUS_PORT", "9090")
		prometheusAddress = fmt.Sprintf("http://%s:%s", prometheusHost, prometheusPort)
	)
	log.SetFlags(log.Lmicroseconds)
	ticker := time.NewTicker(time.Millisecond * 20000)
	defer ticker.Stop()

	locustSwamEndpoint := fmt.Sprintf("%s/swarm", locustAddress)
	users := 0
	HttpPost(locustSwamEndpoint, 0, 1)
	log.Printf("Load test started. Users set to 0.\n")
	for {
		select {
		case <-ticker.C:
			users += 10
			ret := getMetrics(prometheusAddress)
			if !ret {
				log.Printf("Metrics check failed. Stopping load test.\n")
				HttpPost(locustSwamEndpoint, 0, 1)
				os.Exit(0)
			}
			log.Printf("increase users to %d\n", users)
			err := HttpPost(locustSwamEndpoint, users, 2)
			if err != nil {
				log.Fatal(err)
			}
		}
		if users >= 100 {
			HttpPost(locustSwamEndpoint, 0, 1)
			break
		}
	}
}
