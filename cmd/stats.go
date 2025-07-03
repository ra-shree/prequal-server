package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/ra-shree/prequal-server/common"
)

func main() {
	config := common.LoadConfig("config.yaml")
	resp, err := http.Get(fmt.Sprintf("http://localhost:%v/%v", config.Server.Port, config.Server.StatRoute))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var stats []common.ReplicaStatisticsParameters
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"Name", "Successful", "Failed", "In-Flight", "Latency", "Last 10 Latency", "Status",
	})

	for _, stat := range stats {
		lastTen := ""
		for i, val := range stat.LastTenLatency {
			if i > 0 {
				lastTen += ", "
			}
			lastTen += strconv.FormatUint(val, 10)
		}

		table.Append([]string{
			stat.Name,
			strconv.Itoa(stat.SuccessfulRequests),
			strconv.Itoa(stat.FailedRequests),
			strconv.FormatUint(stat.RequestsInFlight, 10),
			strconv.FormatUint(stat.Latency, 10),
			lastTen,
			stat.Status,
		})
	}

	table.Render()
}
