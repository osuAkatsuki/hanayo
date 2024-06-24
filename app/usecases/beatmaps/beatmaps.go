package beatmaps

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
)

type Beatmap struct {
	ID               int `json:"BeatmapID"`
	ParentSetID      int
	DiffName         string
	FileMD5          string
	Mode             int
	BPM              float64
	AR               float32
	OD               float32
	CS               float32
	HP               float32
	TotalLength      int
	HitLength        int
	Playcount        int
	Passcount        int
	MaxCombo         int
	DifficultyRating float64
}
type BeatmapSet struct {
	ID               int `json:"SetID"`
	ChildrenBeatmaps []Beatmap
	RankedStatus     int
	ApprovedDate     time.Time
	LastUpdate       time.Time
	LastChecked      time.Time
	Artist           string
	Title            string
	Creator          string
	Source           string
	Tags             string
	HasVideo         bool
	Genre            int
	Language         int
	Favourites       int
}

func GetBeatmapData(b string) (beatmap Beatmap, err error) {
	settings := settingsState.GetSettings()
	resp, err := http.Get(settings.INTERNAL_BEATMAPS_SERVICE_BASE_URL + "/api/b/" + b)
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

func GetBeatmapSetData(beatmap Beatmap) (bset BeatmapSet, err error) {
	settings := settingsState.GetSettings()
	resp, err := http.Get(settings.INTERNAL_BEATMAP_MIRROR_API_URL + "/s/" + strconv.Itoa(beatmap.ParentSetID))
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
