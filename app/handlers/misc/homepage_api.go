package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/services"
)

// Valid mode combinations:
// Vanilla (rx=0): mode 0-3 -> combined 0-3
// Relax (rx=1): mode 0-2 -> combined 4-6
// Autopilot (rx=2): mode 0 -> combined 8

func getCombinedMode(mode, rx int) (int, bool) {
	switch rx {
	case 0: // Vanilla: all modes
		if mode >= 0 && mode <= 3 {
			return mode, true
		}
	case 1: // Relax: std, taiko, ctb (no mania)
		if mode >= 0 && mode <= 2 {
			return mode + 4, true
		}
	case 2: // Autopilot: std only
		if mode == 0 {
			return 8, true
		}
	}
	return 0, false
}

// HomepageActivityHandler returns activity data for a specific mode
func HomepageActivityHandler(c *gin.Context) {
	modeStr := c.DefaultQuery("mode", "0")
	rxStr := c.DefaultQuery("rx", "0")

	mode, err := strconv.Atoi(modeStr)
	if err != nil || mode < 0 || mode > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode parameter"})
		return
	}

	rx, err := strconv.Atoi(rxStr)
	if err != nil || rx < 0 || rx > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rx parameter"})
		return
	}

	combinedMode, valid := getCombinedMode(mode, rx)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode/rx combination"})
		return
	}

	// Fetch first places from Redis
	firstPlacesKey := fmt.Sprintf("akatsuki:first_places:%d", combinedMode)
	firstPlacesRaw := services.RD.Get(firstPlacesKey)
	var firstPlaces []map[string]interface{}
	if firstPlacesRaw != nil && firstPlacesRaw.Err() == nil {
		json.Unmarshal([]byte(firstPlacesRaw.Val()), &firstPlaces)
	}

	// Fetch high PP plays from Redis
	highPPKey := fmt.Sprintf("akatsuki:high_pp:%d", combinedMode)
	highPPRaw := services.RD.Get(highPPKey)
	var highPP []map[string]interface{}
	if highPPRaw != nil && highPPRaw.Err() == nil {
		json.Unmarshal([]byte(highPPRaw.Val()), &highPP)
	}

	c.JSON(http.StatusOK, gin.H{
		"first_places": firstPlaces,
		"high_pp":      highPP,
	})
}
