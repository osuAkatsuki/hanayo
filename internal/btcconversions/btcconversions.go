package btcconversions

import (
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/exp/slog"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

var rates = struct {
	EUR float64
	USD float64
}{
	// 15m values fetched from blockchain ticker on 2022-07-27
	21003.14,
	21309.21,
}

// GetRates returns the bitcoin rates as JSON.
func GetRates(c *gin.Context) {
	c.JSON(200, rates)
}

func init() {
	go func() {
		for {
			updateRates()
			time.Sleep(time.Hour)
		}
	}()
}

type blockchainCurrency struct {
	// The X is necessary a) to make this exported b) because an identifier
	// can't start with a number.
	X15m float64 `json:"15m"`
}

func updateRates() {
	resp, err := http.Get("https://blockchain.info/ticker")
	if err != nil {
		slog.Error("Failed to update blockchain rates", "error", err.Error())
		return
	}

	m := make(map[string]blockchainCurrency)
	json.NewDecoder(resp.Body).Decode(&m)

	rates.EUR = m["EUR"].X15m
	rates.USD = m["USD"].X15m
}
