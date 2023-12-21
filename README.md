**Project Overview**
- **Title**: Binance Futures Tracker
- **Description**: Develop a command-line application in Golang that tracks and analyzes Binance Futures USDT trading pairs.

**Key Features**
1. **API Integration**: Connect to the Binance API to retrieve live prices of specified USDT trading pairs.
2. **Data Analysis**:
   - Calculate the mean height of 15-minute candles for each trading pair over the last 12 hours. The height is defined as the price difference between the opening and closing prices.
   - Compute the standard deviation of the live price from the stored average heights.
3. **Display**:
   - Present the top 20-30 currencies in a table format in the command line.
   - Each row should include the ticker, the heights of the last 10 candles, the average height for the last 12 hours, the standard deviation based on live prices (updated every minute), the average volume over the last 12 hours, and the current volume.
   - Sort the list by the standard deviation in descending order.

**Technical Stack**
- **Programming Language**: GoLang
- **Database**: Redis (for storing and retrieving candle data)
