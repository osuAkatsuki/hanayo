package beatmaps

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	"github.com/osuripple/cheesegull/models"
)

func GetBeatmapData(b string) (beatmap models.Beatmap, err error) {
	settings := settingsState.GetSettings()
	resp, err := http.Get(settings.BEATMAP_MIRROR_API_URL + "/b/" + b)
	if err != nil {
		return beatmap, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return beatmap, err
	}

	err = json.Unmarshal(body, &beatmap)
	if err != nil {
		return beatmap, err
	}

	return beatmap, nil
}

func GetBeatmapSetData(beatmap models.Beatmap) (bset models.Set, err error) {
	settings := settingsState.GetSettings()
	resp, err := http.Get(settings.BEATMAP_MIRROR_API_URL + "/s/" + strconv.Itoa(beatmap.ParentSetID))
	if err != nil {
		return bset, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bset, err
	}

	err = json.Unmarshal(body, &bset)
	if err != nil {
		return bset, err
	}

	return bset, nil
}
