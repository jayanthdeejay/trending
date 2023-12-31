# Binance Futures Tracker

## Project Overview

**Title**: Binance Futures Tracker

**Description**: 
This project involves developing a web application in Golang using the Gin framework to track and analyze Binance Futures USDT trading pairs. The application provides real-time data visualization and analysis of cryptocurrency market trends.

## Key Features

1. **Real-Time Data Streaming**:
   - Establish WebSocket connections to the Binance API for live updates of USDT trading pairs.
   - Efficiently handle and process streaming data for real-time display.

2. **Data Analysis and Storage**:
   - Calculate and store the mean height of 15-minute candles for each trading pair over the last 12 hours. The height refers to the price difference between the opening and closing prices.
   - Compute and continuously update the standard deviation of the live price from the stored average heights.
   - Utilize InfluxDB for efficient storage and retrieval of time-series data.

3. **Web Interface**:
   - Develop a user-friendly web interface using Tailwind CSS.
   - Display the top 20-30 cryptocurrencies, sorted by standard deviation in descending order.
   - Each entry includes the ticker, heights of the last 10 candles, average height for the last 12 hours, standard deviation from live prices, average volume over the last 12 hours, and current volume.
   - Ensure the interface provides a dynamic and responsive user experience.

## Technical Stack

- **Programming Language**: GoLang
- **Web Framework**: Gin
- **Database**: InfluxDB (for efficient handling of time-series data)
- **Frontend**: Tailwind CSS (for styling the web interface)
- **API**: Binance API for live market data

## Getting Started

1. **Installation**:
   - Instructions on setting up the project locally.
   - Details on configuring InfluxDB and the Binance API connection.

2. **Running the Application**:
   - Steps to start the web server and access the web application.

3. **Usage**:
   - Guide on how to navigate and use the features of the web application.

## Contribution

- Guidelines for contributing to the project, including coding standards, pull request process, etc.

## License

- Information about the project's license.