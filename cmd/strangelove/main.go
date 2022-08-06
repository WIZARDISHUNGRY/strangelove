package main

// https://ride.citibikenyc.com/system-data
// 40.688265,-73.9184594,21z

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/Eraac/gbfs"
	gbfsspec "github.com/Eraac/gbfs/spec/v2.0"
	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"golang.org/x/exp/slices"
)

func main() {
	c, err := gbfs.NewHTTPClient(
		gbfs.HTTPOptionClient(http.Client{Timeout: 10 * time.Second}),
		gbfs.HTTPOptionBaseURL("http://gbfs.citibikenyc.com/gbfs"),
		gbfs.HTTPOptionLanguage("en"),
	)
	if err != nil {
		panic(err)
	}

	var si gbfsspec.FeedStationInformation

	if err := c.Get(gbfsspec.FeedKeyStationInformation, &si); err != nil {
		panic(err)
	}
	var ss gbfsspec.FeedStationStatus
	if err := c.Get(gbfsspec.FeedKeyStationStatus, &ss); err != nil {
		panic(err)
	}
	stationMap := make(map[string]gbfsspec.StationStatus)
	for _, s := range ss.Data.Stations {
		stationMap[s.StationID] = s
	}

	geo1 := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
	lat, lon := 40.688265, -73.9184594

	dist := func(s gbfsspec.StationInformation) float64 {
		distance, _ := geo1.To(lat, lon, s.Latitude, s.Longitude)
		return distance
	}

	slices.SortFunc(si.Data.Stations, func(a, b gbfsspec.StationInformation) bool {
		return dist(a) < dist(b)
	})

	myStations := si.Data.Stations[:5]

	for _, s := range myStations {
		distance, bearing := geo1.To(lat, lon, s.Latitude, s.Longitude)

		statusStr := "?????"
		st, ok := stationMap[s.StationID]
		if ok {
			statusStr = fmt.Sprintf("%d/%d", st.NumBikesAvailable, s.Capacity)
		}

		fmt.Printf("%s (%.0fm %s): %s\n", s.Name, distance, direction(bearing), statusStr)
		direction(bearing)
	}
	// fmt.Printf("Last updated: %s\n", si.LastUpdated.ToTime().String())

}

func direction(bearing float64) string {
	const degrees = 360
	if bearing < 0 {
		bearing += degrees
	}
	dirs := []string{
		"N",
		"NE",
		"E",
		"SE",
		"S",
		"SW",
		"W",
		"NW",
	}
	dirSize := degrees / len(dirs)
	bearing -= float64(dirSize) / 2
	if bearing < 0 {
		bearing += degrees
	}
	idx := int(math.Round(bearing/float64(dirSize))) % len(dirs)
	return dirs[idx]
}
