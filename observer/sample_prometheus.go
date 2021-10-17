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

var (
	client  api.Client
	promApi v1.API
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

var (
	// Response Time (90%ile)
	queryResponseTime = string(
		`
		histogram_quantile(0.90,
			sum(rate(istio_request_duration_milliseconds_bucket{
				reporter="source",
				destination_service=~"frontend.default.svc.cluster.local"
				}[1m]
			)) by (le)
		) / 1000
		or
		histogram_quantile(0.90,
			sum(rate(istio_request_duration_seconds_bucket{
				reporter="source",
				destination_service=~"frontend.default.svc.cluster.local"
				}[1m]
			)) by (le)
		)
		`)
	// Success Rate (non-5xx)
	querySuccessRate = string(
		`
		sum(rate(istio_requests_total{
			reporter=~"source",
			destination_service=~"frontend.default.svc.cluster.local",
			response_code!~"5.*"
			}[1m]
		))
		/
		sum(rate(istio_requests_total{
			reporter=~"source",
			destination_service=~"frontend.default.svc.cluster.local"
			}[1m]
		))
		* 100
		`)
	// CPU Utilization
	queryCPUUtilization = string(
		`
		sum(rate(container_cpu_usage_seconds_total{
			pod=~"frontend.*"
			}[5m]))
		/
		sum(container_spec_cpu_quota{
			pod=~"frontend.*"
			} /
			container_spec_cpu_period{
			pod=~"frontend.*"
			})
		* 100
		`)
)

func execPromQL(ctx context.Context, queryString string) (float64, error) {
	now := time.Now()
	query_result, warning, err := promApi.Query(ctx, queryString, now)
	if err != nil {
		log.Printf("Error while querying prometheus: %v\n", err)
		return -1, err
	}
	if len(warning) > 0 {
		log.Printf("PromQL warning: %v\n", warning)
	}
	// response format might be "{} => xxx @[xxx.xxx]"
	rep := regexp.MustCompile(`\s`)
	result := rep.Split(query_result.String(), -1)
	if result[2] == "NaN" {
		log.Printf("One of the metrics got NaN. Is it loaded successfully?\n")
		return -1, nil
	}
	value, err := strconv.ParseFloat(result[2], 64)
	if err != nil {
		log.Printf("Error while converting string to float: %v\n", err)
		return -1, err
	}
	return value, nil
}

func checkMetrics(prometheusAddress string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var (
		err            error
		responseTime   float64
		successRate    float64
		cpuUtilization float64
		ok             = true
	)
	if responseTime, err = execPromQL(ctx, queryResponseTime); err != nil {
		log.Fatalln("execPromQL failed")
	} else if responseTime > 2.00 {
		log.Printf("Response time check failed: %8.2f sec\n", responseTime)
		ok = false
	}
	if successRate, err = execPromQL(ctx, querySuccessRate); err != nil {
		log.Fatalln("execPromQL failed")
	} else if successRate >= 0 && successRate < 99.9 {
		log.Printf("Success rate check failed: %8.2f %\n", successRate)
		ok = false
	}
	if cpuUtilization, err = execPromQL(ctx, queryCPUUtilization); err != nil {
		log.Fatalln("execPromQL failed")
	} else if cpuUtilization > 80 {
		log.Printf("CPU utilization check failed: %8.2f %\n", cpuUtilization)
		ok = false
	}
	log.Printf(`
Response time (90 percentile): %8.2f
Success rate (non 5xx errors): %8.2f
CPU utilization              : %8.2f
`, responseTime, successRate, cpuUtilization)
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

		userIncreaseStep, _ = strconv.Atoi(getEnv("USER_INCREASE_STEP", "10"))
		spawnDuration, _    = strconv.Atoi(getEnv("SPAWN_DURATION", "5"))
		spawnRate           = userIncreaseStep / spawnDuration
		loadDuration, _     = strconv.Atoi(getEnv("LOAD_DURATION", "30"))
	)
	log.SetFlags(log.Lmicroseconds)

	// Prepare for Prometheus API
	client, err := api.NewClient(api.Config{
		Address: prometheusAddress,
	})
	if err != nil {
		log.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	promApi = v1.NewAPI(client)

	ticker := time.NewTicker(time.Second * time.Duration(loadDuration))
	defer ticker.Stop()

	locustSwamEndpoint := fmt.Sprintf("%s/swarm", locustAddress)
	users := 0
	HttpPost(locustSwamEndpoint, 0, 1)
	log.Printf("Load test started. Users set to 0.\n")
	for {
		select {
		case <-ticker.C:
			users += userIncreaseStep
			if ret := checkMetrics(prometheusAddress); !ret {
				log.Printf("Metrics check failed. Stopping load test.\n")
				HttpPost(locustSwamEndpoint, 0, spawnRate)
				os.Exit(0)
			}
			log.Printf("increase users to %d\n", users)
			if err := HttpPost(locustSwamEndpoint, users, spawnRate); err != nil {
				log.Fatal(err)
			}
		}
		if users >= userIncreaseStep*1000 {
			HttpPost(locustSwamEndpoint, 0, 1)
			break
		}
	}
}
