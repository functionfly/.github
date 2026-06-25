package cityranking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type NominatimLatLon struct {
	Value float64
}

func (n *NominatimLatLon) UnmarshalJSON(data []byte) error {
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		n.Value = num
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		n.Value = val
		return nil
	}
	return fmt.Errorf("cannot parse %s as float64", string(data))
}

// NominatimResult represents a geocoding response from OpenStreetMap Nominatim.
type NominatimResult struct {
	Lat        NominatimLatLon `json:"lat"`
	Lon        NominatimLatLon `json:"lon"`
	DisplayName string         `json:"display_name"`
	Type       string          `json:"type"`
	Address    struct {
		City        string `json:"city"`
		State       string `json:"state"`
		StateCode   string `json:"state_code"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
}

// GeoResolver handles geocoding for unknown cities using Nominatim (OpenStreetMap).
type GeoResolver struct {
	client *http.Client
	baseURL string
	log     *logrus.Logger
}

// NewGeoResolver creates a GeoResolver with sensible defaults.
func NewGeoResolver(log *logrus.Logger) *GeoResolver {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &GeoResolver{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: "https://nominatim.openstreetmap.org",
		log:     log,
	}
}

// Geocode looks up coordinates for a location string using Nominatim.
// Returns (lat, lon, displayName, error).
// Rate limit: 1 request/second — caller is responsible for throttling.
func (g *GeoResolver) Geocode(ctx context.Context, location string) (*NominatimResult, error) {
	if location == "" {
		return nil, fmt.Errorf("empty location")
	}

	params := url.Values{}
	params.Set("q", location)
	params.Set("format", "json")
	params.Set("limit", "1")
	params.Set("addressdetails", "1")

	// User-Agent required by Nominatim ToS — identify your application.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/search?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "FunctionFly-CityRankings/1.0 (city-rankings@functionfly.com)")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim returned %d", resp.StatusCode)
	}

	var results []NominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // not found
	}

	return &results[0], nil
}

// StateCentroid returns approximate lat/lon for a US state by its code.
// Used as fallback when geocoding fails but we need something to seed the city.
// Data from USGS and Census Bureau.
var USStateCentroids = map[string]struct{ Lat, Lon float64 }{
	"AL": {32.806671, -86.791130},
	"AK": {61.370716, -152.404419},
	"AZ": {33.729759, -111.431221},
	"AR": {34.969704, -92.373123},
	"CA": {36.116203, -119.681564},
	"CO": {39.059811, -105.311104},
	"CT": {41.597782, -72.755371},
	"DE": {39.318523, -75.507141},
	"FL": {27.766279, -81.686783},
	"GA": {33.040619, -83.643074},
	"HI": {21.094318, -157.498337},
	"ID": {44.240459, -114.478828},
	"IL": {40.349457, -88.986137},
	"IN": {39.849426, -86.258278},
	"IA": {41.011905, -93.210526},
	"KS": {38.526600, -96.726486},
	"KY": {37.668140, -84.670067},
	"LA": {31.169546, -91.867805},
	"ME": {44.693947, -69.381927},
	"MD": {39.063946, -76.802101},
	"MA": {42.230171, -71.530106},
	"MI": {43.326618, -84.536095},
	"MN": {45.694454, -93.900192},
	"MS": {32.741646, -89.678696},
	"MO": {38.456085, -92.288368},
	"MT": {46.921925, -110.454353},
	"NE": {41.125370, -98.268082},
	"NV": {38.313515, -117.055374},
	"NH": {43.452492, -71.563896},
	"NJ": {40.298904, -74.521011},
	"NM": {34.840515, -106.248482},
	"NY": {42.165726, -74.948051},
	"NC": {35.630066, -79.806419},
	"ND": {47.528912, -99.784012},
	"OH": {40.388783, -82.764915},
	"OK": {35.565342, -96.928917},
	"OR": {44.572021, -122.070938},
	"PA": {40.590752, -77.209755},
	"RI": {41.680893, -71.511780},
	"SC": {34.344701, -80.958718},
	"SD": {44.299782, -99.438828},
	"TN": {35.747845, -86.692345},
	"TX": {31.054487, -97.563461},
	"UT": {40.150032, -111.862434},
	"VT": {43.046914, -72.710276},
	"VA": {37.769337, -78.169968},
	"WA": {47.400902, -121.490494},
	"WV": {38.491226, -80.954453},
	"WI": {44.268543, -89.616508},
	"WY": {42.755966, -107.302185},
}

// GetUSStateCentroid returns the approximate center of a US state by its 2-letter code.
func GetUSStateCentroid(stateCode string) (lat, lon float64, ok bool) {
	coord, ok := USStateCentroids[strings.ToUpper(stateCode)]
	return coord.Lat, coord.Lon, ok
}
