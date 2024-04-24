package beatmaps

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	bu "github.com/osuAkatsuki/hanayo/app/usecases/beatmaps"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func BeatmapPageHandler(c *gin.Context) {
	data := new(models.BeatmapPageData)
	defer tu.Resp(c, 200, "beatmap.html", data)

	b := c.Param("bid")
	if _, err := strconv.Atoi(b); err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
	} else {
		data.Beatmap, err = bu.GetBeatmapData(b)
		if err != nil {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}
		data.Beatmapset, err = bu.GetBeatmapSetData(data.Beatmap)
		if err != nil {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}

		sort.Slice(data.Beatmapset.ChildrenBeatmaps, func(i, j int) bool {
			if data.Beatmapset.ChildrenBeatmaps[i].Mode != data.Beatmapset.ChildrenBeatmaps[j].Mode {
				return data.Beatmapset.ChildrenBeatmaps[i].Mode < data.Beatmapset.ChildrenBeatmaps[j].Mode
			}
			return data.Beatmapset.ChildrenBeatmaps[i].DifficultyRating < data.Beatmapset.ChildrenBeatmaps[j].DifficultyRating
		})
	}

	if data.Beatmapset.ID == 0 {
		data.TitleBar = lu.T(c, "Beatmap not found.")
		data.Messages = append(data.Messages, msg.ErrorMessage{lu.T(c, "Beatmap could not be found.")})
		return
	}

	for i := range data.Beatmapset.ChildrenBeatmaps {
		err := services.DB.QueryRow("SELECT playcount, passcount FROM beatmaps WHERE beatmap_md5 = ?", data.Beatmapset.ChildrenBeatmaps[i].FileMD5).Scan(&data.Beatmapset.ChildrenBeatmaps[i].Playcount, &data.Beatmapset.ChildrenBeatmaps[i].Passcount)
		if err != nil {
			slog.Error("Beatmap not found", "error", err.Error())
			data.Beatmapset.ChildrenBeatmaps[i].Playcount = 0
			data.Beatmapset.ChildrenBeatmaps[i].Passcount = 0
		}
	}

	data.BannerContent = fmt.Sprintf("https://assets.ppy.sh/beatmaps/%d/covers/cover.jpg?%d", data.Beatmapset.ID, data.Beatmapset.LastUpdate.Unix())
	data.BannerAbsolute = true
	data.BannerType = 1

	setJSON, err := json.Marshal(data.Beatmapset)
	if err == nil {
		data.SetJSON = string(setJSON)
	} else {
		data.SetJSON = "[]"
	}

	data.TitleBar = lu.T(c, "%s - %s", data.Beatmapset.Artist, data.Beatmapset.Title)
	data.Scripts = append(data.Scripts, "/static/js/pages/beatmap.tablesort.min.js", "/static/js/pages/beatmap.min.js")
}
