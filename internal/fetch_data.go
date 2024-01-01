package internal

import (
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

var UsdtSymbols = []string{
	"1000BONKUSDT", "1000FLOKIUSDT", "1000LUNCUSDT", "1000PEPEUSDT", "1000RATSUSDT", "1000SATSUSDT", "1000SHIBUSDT", "1000XECUSDT", "1INCHUSDT",
	"AAVEUSDT", "ACEUSDT", "ACHUSDT", "ADAUSDT", "AGIXUSDT", "AGLDUSDT", "ALGOUSDT", "ALICEUSDT", "ALPHAUSDT", "AMBUSDT", "ANKRUSDT", "ANTUSDT",
	"APEUSDT", "API3USDT", "APTUSDT", "ARBUSDT", "ARKMUSDT", "ARKUSDT", "ARPAUSDT", "ARUSDT", "ASTRUSDT", "ATAUSDT", "ATOMUSDT", "AUCTIONUSDT",
	"AUDIOUSDT", "AVAXUSDT", "AXSUSDT", "BADGERUSDT", "BAKEUSDT", "BALUSDT", "BANDUSDT", "BATUSDT", "BCHUSDT", "BEAMXUSDT", "BELUSDT", "BICOUSDT",
	"BIGTIMEUSDT", "BLUEBIRDUSDT", "BLURUSDT", "BLZUSDT", "BNBUSDT", "BNTUSDT", "BNXUSDT", "BONDUSDT", "BSVUSDT", "BTCDOMUSDT", "BTCSTUSDT",
	"BTCUSDT", "BTSUSDT", "C98USDT", "CAKEUSDT", "CELOUSDT", "CELRUSDT", "CFXUSDT", "CHRUSDT", "CHZUSDT", "CKBUSDT", "COCOSUSDT", "COMBOUSDT",
	"COMPUSDT", "COTIUSDT", "CRVUSDT", "CTKUSDT", "CTSIUSDT", "CVCUSDT", "CVXUSDT", "CYBERUSDT", "DARUSDT", "DASHUSDT", "DEFIUSDT", "DENTUSDT",
	"DGBUSDT", "DODOXUSDT", "DOGEUSDT", "DOTUSDT", "DUSKUSDT", "DYDXUSDT", "EDUUSDT", "EGLDUSDT", "ENJUSDT", "ENSUSDT", "EOSUSDT", "ETCUSDT",
	"ETHUSDT", "ETHWUSDT", "FETUSDT", "FILUSDT", "FLMUSDT", "FLOWUSDT", "FOOTBALLUSDT", "FRONTUSDT", "FTMUSDT", "FTTUSDT", "FXSUSDT", "GALAUSDT",
	"GALUSDT", "GASUSDT", "GLMRUSDT", "GMTUSDT", "GMXUSDT", "GRTUSDT", "GTCUSDT", "HBARUSDT", "HFTUSDT", "HIFIUSDT", "HIGHUSDT", "HNTUSDT",
	"HOOKUSDT", "HOTUSDT", "ICPUSDT", "ICXUSDT", "IDEXUSDT", "IDUSDT", "ILVUSDT", "IMXUSDT", "INJUSDT", "IOSTUSDT", "IOTAUSDT", "IOTXUSDT",
	"JASMYUSDT", "JOEUSDT", "JTOUSDT", "KASUSDT", "KAVAUSDT", "KEYUSDT", "KLAYUSDT", "KNCUSDT", "KSMUSDT", "LDOUSDT", "LEVERUSDT", "LINAUSDT",
	"LINKUSDT", "LITUSDT", "LOOMUSDT", "LPTUSDT", "LQTYUSDT", "LRCUSDT", "LTCUSDT", "LUNA2USDT", "MAGICUSDT", "MANAUSDT", "MASKUSDT", "MATICUSDT",
	"MAVUSDT", "MBLUSDT", "MDTUSDT", "MEMEUSDT", "MINAUSDT", "MKRUSDT", "MOVRUSDT", "MTLUSDT", "NEARUSDT", "NEOUSDT", "NFPUSDT", "NKNUSDT", "NMRUSDT",
	"NTRNUSDT", "OCEANUSDT", "OGNUSDT", "OMGUSDT", "ONEUSDT", "ONGUSDT", "ONTUSDT", "OPUSDT", "ORBSUSDT", "ORDIUSDT", "OXTUSDT", "PENDLEUSDT",
	"PEOPLEUSDT", "PERPUSDT", "PHBUSDT", "POLYXUSDT", "POWRUSDT", "PYTHUSDT", "QNTUSDT", "QTUMUSDT", "RADUSDT", "RAYUSDT", "RDNTUSDT", "REEFUSDT",
	"RENUSDT", "RIFUSDT", "RLCUSDT", "RNDRUSDT", "ROSEUSDT", "RSRUSDT", "RUNEUSDT", "RVNUSDT", "SANDUSDT", "SCUSDT", "SEIUSDT", "SFPUSDT", "SKLUSDT",
	"SLPUSDT", "SNTUSDT", "SNXUSDT", "SOLUSDT", "SPELLUSDT", "SRMUSDT", "SSVUSDT", "STEEMUSDT", "STGUSDT", "STMXUSDT", "STORJUSDT", "STPTUSDT",
	"STRAXUSDT", "STXUSDT", "SUIUSDT", "SUPERUSDT", "SUSHIUSDT", "SXPUSDT", "THETAUSDT", "TIAUSDT", "TLMUSDT", "TOKENUSDT", "TOMOUSDT", "TRBUSDT",
	"TRUUSDT", "TRXUSDT", "TUSDT", "TWTUSDT", "UMAUSDT", "UNFIUSDT", "UNIUSDT", "USDCUSDT", "USTCUSDT", "VETUSDT", "WAVESUSDT", "WAXPUSDT", "WLDUSDT",
	"WOOUSDT", "XEMUSDT", "XLMUSDT", "XMRUSDT", "XRPUSDT", "XTZUSDT", "XVGUSDT", "XVSUSDT", "YFIUSDT", "YGGUSDT", "ZECUSDT", "ZENUSDT", "ZILUSDT", "ZRXUSDT",
}

// Fetch and save futures data for all USDT symbols
func FetchAndSaveFuturesData() {
	influxToken := os.Getenv("INFLUXDB_TRENDING_TOKEN")
	influxClient := influxdb2.NewClient(influxDBURL, influxToken)
	writeAPI := influxClient.WriteAPI(orgName, bucketName)
	defer influxClient.Close()

	for _, symbol := range UsdtSymbols {
		log.Println("Fetching data for", symbol)
		fetchDataForSymbol(writeAPI, symbol)
	}
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
