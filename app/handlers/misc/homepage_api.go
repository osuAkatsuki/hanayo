package misc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/services"
)

// HomepageActivityHandler returns activity data for a specific mode
func HomepageActivityHandler(c *gin.Context) {
	rxStr := c.DefaultQuery("rx", "1") // Default to relax mode

	// Validate rx parameter (0=vanilla, 1=relax, 2=autopilot)
	if rxStr != "0" && rxStr != "1" && rxStr != "2" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rx parameter"})
		return
	}

	// Fetch first places from Redis
	firstPlacesKey := fmt.Sprintf("akatsuki:first_places:%s", rxStr)
	firstPlacesRaw := services.RD.Get(firstPlacesKey)
	var firstPlaces []map[string]interface{}
	if firstPlacesRaw != nil && firstPlacesRaw.Err() == nil {
		json.Unmarshal([]byte(firstPlacesRaw.Val()), &firstPlaces)
	}

	// Fetch high PP plays from Redis
	highPPKey := fmt.Sprintf("akatsuki:high_pp:%s", rxStr)
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
