package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// StreamResponse represents the response from the Binance WebSocket stream
type StreamResponse struct {
	Stream string    `json:"stream"`
	Data   MarkPrice `json:"data"`
}

// MarkPrice represents the market price information in the stream
type MarkPrice struct {
	EventType string  `json:"e"`
	Time      int64   `json:"E"`
	Symbol    string  `json:"s"`
	Price     float64 `json:"p,string"` // Note the use of 'string' in the tag to parse stringified float
}

type SymbolData struct {
	Prices              []float64  // Store recent prices for rate of change calculation
	AvgRateChange       float64    // Average rate of change over 12 hours
	HourlyAvgRateChange [6]float64 // Average rate of change for each of the last 6 hours
	CurrentRateChange   float64    // Current rate of change between streams
	LastPrice           float64    // Last received price
	LastUpdated         time.Time  // Time when the last update was received
}

var symbolDataMap map[string]*SymbolData = make(map[string]*SymbolData)

func ConnectToWebSocket(symbols []string) {
	var c *websocket.Conn
	var err error
	reconnectInterval := 5 * time.Second

	for {
		if c == nil {
			var streams string
			for _, symbol := range symbols {
				streams += fmt.Sprintf("%s@markPrice/", strings.ToLower(symbol))
			}
			streams = strings.TrimRight(streams, "/")

			wsURL := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streams)
			c, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				log.Printf("WebSocket Dial Error: %v, retrying in %v", err, reconnectInterval)
				time.Sleep(reconnectInterval)
				continue
			}
			log.Println("Connected to WebSocket")
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket Read Error: %v, reconnecting...", err)
			c.Close()
			c = nil
			continue
		}

		var response StreamResponse
		if err := json.Unmarshal(message, &response); err != nil {
			log.Println("JSON unmarshal error:", err)
			continue
		}

		updateSymbolData(response.Data.Symbol, response.Data.Price)
	}
}

const windowSize = 12 * 60 * 60 / 3 // 12 hours window, assuming data every 3 seconds
const pointsPerHour = 60 * 60 / 3   // Number of data points in an hour

func updateSymbolData(symbol string, price float64) {
	data, exists := symbolDataMap[symbol]
	if !exists {
		data = &SymbolData{LastUpdated: time.Now()}
		symbolDataMap[symbol] = data
	}

	// Calculate current rate of change
	if data.LastPrice != 0 {
		data.CurrentRateChange = math.Abs(price - data.LastPrice)
	}
	data.LastPrice = price

	// Add the new price to the slice
	data.Prices = append(data.Prices, price)

	// If the window size is exceeded, remove the oldest price
	if len(data.Prices) > windowSize {
		data.Prices = data.Prices[1:]
	}

	// Calculate the average rate of change
	if len(data.Prices) > 1 {
		var totalChange float64
		for i := 1; i < len(data.Prices); i++ {
			totalChange += math.Abs(data.Prices[i] - data.Prices[i-1])
		}
		data.AvgRateChange = totalChange / float64(len(data.Prices)-1)
	}

	// Calculate hourly averages
	for i := 0; i < 6; i++ {
		endIndex := len(data.Prices) - i*pointsPerHour
		if endIndex < 0 {
			// If endIndex is negative, no data is available for this hour.
			data.HourlyAvgRateChange[i] = 0
			continue
		}
		startIndex := max(0, endIndex-pointsPerHour)
		if startIndex >= endIndex {
			// Not enough data points in this window
			data.HourlyAvgRateChange[i] = 0
			continue
		}
		data.HourlyAvgRateChange[i] = calculateAverageRateChange(data.Prices[startIndex:endIndex])
	}

	// Update last updated time
	data.LastUpdated = time.Now()
}

func calculateAverageRateChange(prices []float64) float64 {
	if len(prices) <= 1 {
		return 0
	}
	var totalChange float64
	for i := 1; i < len(prices); i++ {
		totalChange += math.Abs(prices[i] - prices[i-1])
	}
	return totalChange / float64(len(prices)-1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func DisplayActiveSymbols() {
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		displayTable()
	}
}

func displayTable() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	table := widgets.NewTable()
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = false

	updateTable := func() {
		table.Rows = [][]string{
			{"Symbol", "AvgRateChange (12h)", "CurrentRateChange", "1h", "2h", "3h", "4h", "5h", "6h"},
		}

		var sortedSymbols []string
		for symbol := range symbolDataMap {
			sortedSymbols = append(sortedSymbols, symbol)
		}

		sort.Slice(sortedSymbols, func(i, j int) bool {
			return symbolDataMap[sortedSymbols[i]].CurrentRateChange > symbolDataMap[sortedSymbols[j]].CurrentRateChange
		})

		for i := 0; i < 63 && i < len(sortedSymbols); i++ {
			symbol := sortedSymbols[i]
			data := symbolDataMap[symbol]
			row := []string{
				symbol,
				strconv.FormatFloat(data.AvgRateChange, 'f', 5, 64),
				strconv.FormatFloat(data.CurrentRateChange, 'f', 5, 64),
			}
			for j := 0; j < 6; j++ {
				if j < len(data.HourlyAvgRateChange) {
					row = append(row, strconv.FormatFloat(data.HourlyAvgRateChange[j], 'f', 5, 64))
				} else {
					row = append(row, "N/A")
				}
			}
			table.Rows = append(table.Rows, row)
		}

		grid.Set(
			ui.NewRow(1.0,
				ui.NewCol(1.0, table),
			),
		)
	}

	updateTable()
	ui.Render(grid)

	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = false

	// Update and render the table in a ticker loop
	ticker := time.NewTicker(3 * time.Second)
	uiEvents := ui.PollEvents()
	for {
		select {
		case <-ticker.C:
			updateTable()
			ui.Render(grid)
		case e := <-uiEvents:
			if e.ID == "q" || e.Type == ui.KeyboardEvent {
				return
			}
		}
	}
}

func writeToInfluxDB(writeAPI api.WriteAPI) {
	for symbol, data := range symbolDataMap {
		// Prepare fields for InfluxDB point
		fields := map[string]interface{}{
			"avg_rate_change":     data.AvgRateChange,
			"current_rate_change": data.CurrentRateChange,
		}

		// Add last 6 hours' average rate changes
		for i, rateChange := range data.HourlyAvgRateChange {
			fields[fmt.Sprintf("hourly_avg_change_%dh", i+1)] = rateChange
		}

		// Create a point and add to batch
		p := influxdb2.NewPoint(
			"symbol_rate_change",
			map[string]string{"symbol": symbol},
			fields,
			time.Now(),
		)
		writeAPI.WritePoint(p)
	}
	writeAPI.Flush()
}

func loadDataFromInfluxDB(queryAPI api.QueryAPI) {
	fluxQuery := `from(bucket:"futures")
								|> range(start: -12h)
								|> filter(fn: (r) => r._measurement == "symbol_rate_change")
								|> last()`

	result, err := queryAPI.Query(context.Background(), fluxQuery)
	if err != nil {
		log.Fatalf("error querying data: %v", err)
	}

	for result.Next() {
		if result.Record() != nil {
			symbol := result.Record().Values()["symbol"].(string)
			avgRateChange, ok := result.Record().Values()["avg_rate_change"].(float64)
			if !ok {
				avgRateChange = 0 // Default value if nil or not a float64
			}
			currentRateChange, ok := result.Record().Values()["current_rate_change"].(float64)
			if !ok {
				currentRateChange = 0 // Default value if nil or not a float64
			}

			data := &SymbolData{
				AvgRateChange:     avgRateChange,
				CurrentRateChange: currentRateChange,
				// populate other fields as needed
			}
			symbolDataMap[symbol] = data
		}
	}

	if err := result.Err(); err != nil {
		log.Fatalf("error iterating over query result: %v", err)
	}
}

func init() {
	// Initialize the InfluxDB client and query API
	influxToken := os.Getenv("INFLUXDB_TRENDING_TOKEN")
	influxClient := influxdb2.NewClient(influxDBURL, influxToken)
	queryAPI := influxClient.QueryAPI(orgName)
	writeAPI := influxClient.WriteAPI(orgName, bucketName)

	// Load data from InfluxDB
	loadDataFromInfluxDB(queryAPI)

	// Schedule writeToInfluxDB to run every minute
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		for range ticker.C {
			writeToInfluxDB(writeAPI)
		}
	}()

	// Start displaying symbols
	go DisplayActiveSymbols()
}
