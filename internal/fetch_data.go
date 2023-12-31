package internal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/tidwall/gjson"
)

const (
	binanceFuturesAPIBaseURL = "https://fapi.binance.com/fapi/v1/klines"
	influxDBURL              = "http://localhost:8086" // Change to your InfluxDB URL
	orgName                  = "trending"
	bucketName               = "futures"
)

// Fetch and save futures data for all USDT symbols
func FetchAndSaveFuturesData() {
	influxToken := os.Getenv("INFLUXDB_TRENDING_TOKEN")
	influxClient := influxdb2.NewClient(influxDBURL, influxToken)
	writeAPI := influxClient.WriteAPI(orgName, bucketName)
	defer influxClient.Close()

	symbols := readSymbols("../binance/usdt_symbols.txt")
	for _, symbol := range symbols {
		log.Println("Fetching data for", symbol)
		fetchDataForSymbol(writeAPI, symbol)
	}
}

// Read symbols from the usdt_symbols.txt file
func readSymbols(filePath string) []string {
	log.Println("Reading symbols from", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var symbols []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		symbols = append(symbols, scanner.Text())
	}
	return symbols
}

// Fetch data for a specific symbol and write to InfluxDB
func fetchDataForSymbol(writeAPI api.WriteAPI, symbol string) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	interval := "15m"
	limit := 100 // Adjust as needed, max 1500

	url := fmt.Sprintf("%s?symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=%d",
		binanceFuturesAPIBaseURL, symbol, interval, startTime.UnixMilli(), endTime.UnixMilli(), limit)

	log.Println("Fetching data from", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	result := gjson.ParseBytes(body)
	for _, k := range result.Array() {
		openTime := k.Array()[0].Int()
		open := k.Array()[1].Float()
		high := k.Array()[2].Float()
		low := k.Array()[3].Float()
		close := k.Array()[4].Float()
		volume := k.Array()[5].Float()

		// Create a point and add to batch
		p := influxdb2.NewPoint(
			"futures",
			map[string]string{"symbol": symbol},
			map[string]interface{}{
				"open":   open,
				"high":   high,
				"low":    low,
				"close":  close,
				"volume": volume,
			},
			time.Unix(0, openTime*int64(time.Millisecond)),
		)
		writeAPI.WritePoint(p)
	}

	// Ensures background processes finishes
	writeAPI.Flush()
}
