package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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
	var streams string
	for _, symbol := range symbols {
		// Symbols need to be lowercase
		streams += fmt.Sprintf("%s@markPrice/", strings.ToLower(symbol))
	}
	streams = strings.TrimRight(streams, "/") // Remove the trailing slash

	wsURL := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streams)
	log.Println(wsURL)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var response StreamResponse
		if err := json.Unmarshal(message, &response); err != nil {
			log.Println("json unmarshal:", err)
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
	// Creating a slice for sorting
	var sortedSymbols []string
	for symbol := range symbolDataMap {
		sortedSymbols = append(sortedSymbols, symbol)
	}

	// Sort symbols based on CurrentRateChange
	sort.Slice(sortedSymbols, func(i, j int) bool {
		return symbolDataMap[sortedSymbols[i]].CurrentRateChange > symbolDataMap[sortedSymbols[j]].CurrentRateChange
	})

	// Display top 63 symbols
	fmt.Println("Top Active Symbols:")
	fmt.Println("Symbol\t\tAvgRateChange (12h)\tCurrentRateChange\t\tHourly Change (Last 6h)")
	for i := 0; i < 63 && i < len(sortedSymbols); i++ {
		symbol := sortedSymbols[i]
		data := symbolDataMap[symbol]
		fmt.Printf("%-20s %-20.4f %-20.4f\t", symbol, data.AvgRateChange, data.CurrentRateChange)
		for _, hourlyChange := range data.HourlyAvgRateChange {
			fmt.Printf("%-10.4f ", hourlyChange)
		}
		fmt.Println()
	}
}

func init() {
	// Initialize any necessary data structures or goroutines
	go DisplayActiveSymbols()
}
