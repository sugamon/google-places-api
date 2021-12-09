package places

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	key string
}

func NewClient(key string) PlaceAPI {
	return &Client{
		key: key,
	}
}

type PlaceAPI interface {
	TextSearch(ctx context.Context, query string, params OptionalParams) (*PlaceResult, error)
}

type PlaceResult struct {
	HtmlAttributions []string `json:"html_attributions"`
	Status           string   `json:"status"`
	Results          []Place  `json:"results"`
}

type Place struct {
	FormattedAddress    string   `json:"formatted_address"`
	Icon                string   `json:"icon"`
	IconBackgroundColor string   `json:"icon_background_color"`
	IconMaskBaseUri     string   `json:"icon_mask_base_uri"`
	Name                string   `json:"name"`
	PlaceID             string   `json:"place_id"`
	PlusCode            PlusCode `json:"plus_code"`
	PriceLevel          int64    `json:"price_level"`
	Rating              float64  `json:"rating"`
	Reference           string   `json:"reference"`
	Types               []string `json:"types"`
	UserRatingsTotal    int64    `json:"user_ratings_total"`
	BusinessStatus      string   `json:"business_status"`
	Geometry            Geometry
	Photos              []Photo
	OpeningHours        OpeningHours
}

type PlusCode struct {
	CompoundCode string `json:"compound_code"`
	GlobalCode   string `json:"global_code"`
}

type Photo struct {
	Height           int64    `json:"height"`
	HtmlAttributions []string `json:"html_attributions"`
	PhotoReference   string   `json:"photo_reference"`
	Width            int64    `json:"width"`
}

type OpeningHours struct {
	OpenNow bool
}

type Geometry struct {
	Location Location
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type OptionalParams struct {
	Language string

	// The point around which to retrieve place information.
	// This must be specified as latitude,longitude.
	Location string

	// Restricts results to only those places within the specified range.
	// Valid values range between 0 (most affordable) to 4 (most expensive), inclusive.
	// The exact amount indicated by a specific value will vary from region to region.
	Maxprice string

	// Restricts results to only those places within the specified range.
	// Valid values range between 0 (most affordable) to 4 (most expensive),
	// inclusive. The exact amount indicated by a specific value will vary from region to region.
	Minprice string

	// Returns only those places that are open for business at the time the query is sent.
	// Places that do not specify opening hours in the Google Places database will not be returned if you include this parameter in your query.
	Opennow bool

	// Returns up to 20 results from a previously run search.
	// Setting a pagetoken parameter will execute a search with the same parameters used previously — all parameters other than pagetoken will be ignored.
	Pagetoken string

	// Defines the distance (in meters) within which to return place results.
	// You may bias results to a specified circle by passing a location and a radius parameter.
	// Doing so instructs the Places service to prefer showing results within that circle; results outside of the defined area may still be displayed.
	Radius string

	// The region code, specified as a ccTLD ("top-level domain") two-character value.
	// Most ccTLD codes are identical to ISO 3166-1 codes, with some notable exceptions.
	// For example, the United Kingdom's ccTLD is "uk" (.co.uk) while its ISO 3166-1 code is "gb" (technically for the entity of "The United Kingdom of Great Britain and Northern Ireland").
	Region string

	// Restricts the results to places matching the specified type.
	// Only one type may be specified. If more than one type is provided, all types following the first entry are ignored.
	Type string
}

const textSearchURL = "https://maps.googleapis.com/maps/api/place/textsearch/json"

// TextSearch
// https://developers.google.com/maps/documentation/places/web-service/search-text
func (c *Client) TextSearch(ctx context.Context, query string, params OptionalParams) (*PlaceResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required. %v", query)
	}

	u, err := c.createURL(textSearchURL, params)
	if err != nil {
		return nil, fmt.Errorf("faild to create url. %v", err)
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("parse query error. %v", err)
	}

	q.Add("query", query)
	u.RawQuery = q.Encode()

	var body PlaceResult

	resp, err := request(u.String(), &body)
	if err != nil {
		return nil, err
	}

	placeResult := resp.(*PlaceResult)

	return placeResult, nil
}

func (c *Client) createURL(baseURL string, params OptionalParams) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = c.createParams(u, params)
	return u, nil
}

func (c *Client) createParams(u *url.URL, params OptionalParams) string {
	q := u.Query()
	q.Set("key", c.key)

	if params.Pagetoken != "" {
		q.Set("pagetoken", params.Pagetoken)
		return q.Encode()
	}

	if params.Language != "" {
		q.Set("language", params.Language)
	}

	if params.Region != "" {
		q.Set("region", params.Region)
	}

	if params.Location != "" {
		loc := strings.Split(params.Location, ",")
		if len(loc) == 2 {
			q.Set("location", params.Location)
		}
	}

	if params.Maxprice != "" {
		maxp, err := strconv.Atoi(params.Maxprice)
		if err == nil && (0 <= maxp && maxp <= 4) {
			q.Set("maxprice", params.Maxprice)
		}
	}

	if params.Minprice != "" {
		minp, err := strconv.Atoi(params.Minprice)
		if err == nil && (0 <= minp && minp <= 4) {
			q.Set("minprice", params.Minprice)
		}
	}

	if params.Minprice != "" {
		minp, err := strconv.Atoi(params.Minprice)
		if err == nil && (0 <= minp && minp <= 4) {
			q.Set("minprice", params.Minprice)
		}
	}

	if params.Opennow {
		q.Set("opennow", "true")
	}

	if params.Radius != "" {
		_, err := strconv.Atoi(params.Radius)
		if err == nil {
			q.Set("radius", params.Radius)
		}
	}

	if params.Type != "" {
		q.Set("type", params.Type)
	}

	return q.Encode()
}

func request(url string, response *PlaceResult) (interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error. status is %v", resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bytes := []byte(body)
	json.Unmarshal(bytes, &response)

	return response, nil
}
