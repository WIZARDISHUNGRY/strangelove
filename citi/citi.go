package citi

// https://ride.citibikenyc.com/system-data
// 40.688265,-73.9184594,21z

import (
	"fmt"
	"math"
	"net/http"
	"time"

	spark "bitbucket.org/dtolpin/gosparkline"
	"github.com/Eraac/gbfs"
	gbfsspec "github.com/Eraac/gbfs/spec/v2.0"
	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"golang.org/x/exp/slices"
)

const sparklineLen = 10

func Citi(lat, lon float64) <-chan []string {

	c, err := gbfs.NewHTTPClient(
		gbfs.HTTPOptionClient(http.Client{Timeout: 10 * time.Second}),
		gbfs.HTTPOptionBaseURL("http://gbfs.citibikenyc.com/gbfs"),
		gbfs.HTTPOptionLanguage("en"),
	)
	if err != nil {
		panic(err)
	}

	out := make(chan []string)

	var si gbfsspec.FeedStationInformation

	if err := c.Get(gbfsspec.FeedKeyStationInformation, &si); err != nil {
		panic(err)
	}

	geo1 := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)

	dist := func(s gbfsspec.StationInformation) float64 {
		distance, _ := geo1.To(lat, lon, s.Latitude, s.Longitude)
		return distance
	}

	slices.SortFunc(si.Data.Stations, func(a, b gbfsspec.StationInformation) bool {
		return dist(a) < dist(b)
	})

	myStations := si.Data.Stations[:5]

	sparklines := make(map[string][]float64)

	const pollInterval = 60 * time.Second

	go func() {
		for i := 0; true; i++ {
			var ss gbfsspec.FeedStationStatus
			if err := c.Get(gbfsspec.FeedKeyStationStatus, &ss); err != nil {
				panic(err)
			}
			nextUpdate := time.Now().Add(pollInterval)
			throbber := []float64{0}

			stationMap := make(map[string]gbfsspec.StationStatus)
			for _, s := range ss.Data.Stations {
				stationMap[s.StationID] = s
			}

			first := true
			for time.Now().Before(nextUpdate) {

				output := make([]string, 0, len(myStations))

				for i, s := range myStations {
					distance, bearing := geo1.To(lat, lon, s.Latitude, s.Longitude)

					statusStr := "?????"
					st, ok := stationMap[s.StationID]
					if ok {
						frac := float64(st.NumBikesAvailable)
						if first {
							sparklines[s.StationID] = append([]float64{frac}, sparklines[s.StationID]...)
						}
						if len(sparklines[s.StationID]) > sparklineLen {
							sparklines[s.StationID] = sparklines[s.StationID][:sparklineLen]
						}
						statusStr = fmt.Sprintf("%2.0d/%2.0d %-10s", st.NumBikesAvailable, s.Capacity, spark.Line(sparklines[s.StationID]))
					}

					throb := " "

					if i == len(myStations)-1 && len(throbber) > 0 {
						runes := ([]rune)(spark.Line(throbber))
						throb = string(runes[len(runes)-1:])
					}

					str := fmt.Sprintf("%s\n%4.0fm %2s %s %s", s.Name, distance, direction(bearing), statusStr, throb)
					output = append(output, str)
				}

				dur := time.Until(nextUpdate).Seconds()
				throbber = append(throbber, float64(dur))
				out <- output
				time.Sleep(time.Second)
				first = false
			}

			// fmt.Printf("Last updated: %s\n", si.LastUpdated.ToTime().String())
		}
	}()

	return out
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
