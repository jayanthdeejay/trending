package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func extractUSDTSymbols(jsonFilePath string) ([]string, error) {
	// Read the JSON file
	fileBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, err
	}

	// Decode the JSON data
	var data map[string]interface{}
	err = json.Unmarshal(fileBytes, &data)
	if err != nil {
		return nil, err
	}

	// Extract USDT symbols
	usdtSymbols := []string{}
	symbolsArray := data["symbols"].([]interface{})
	for _, symbolInfo := range symbolsArray {
		symbol := symbolInfo.(map[string]interface{})["symbol"].(string)
		if strings.HasSuffix(symbol, "USDT") {
			usdtSymbols = append(usdtSymbols, symbol)
		}
	}

	return usdtSymbols, nil
}

func main() {
	// Fetch exchange information from Binance Futures API (replace with your API key)
	resp, err := http.Get("https://fapi.binance.com/fapi/v1/exchangeInfo")
	if err != nil {
		fmt.Println("Error fetching exchange info:", err)
		return
	}
	defer resp.Body.Close()

	// Save the response to a file
	jsonBytes, err := io.ReadAll(resp.Body) // Use io.ReadAll for reading
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	err = os.WriteFile("assets.json", jsonBytes, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	// Extract USDT symbols from the JSON file
	usdtSymbols, err := extractUSDTSymbols("assets.json")
	if err != nil {
		fmt.Println("Error extracting USDT symbols:", err)
		return
	}

	// Write the symbols to a file
	file, err := os.Create("usdt_symbols.txt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer file.Close()

	for _, symbol := range usdtSymbols {
		_, err := file.WriteString(symbol + "\n")
		if err != nil {
			fmt.Println("Error writing symbol to file:", err)
			return
		}
	}

	fmt.Println("USDT symbols written to usdt_symbols.txt")
}
