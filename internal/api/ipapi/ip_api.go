package ipapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/lescactus/geolocation-go/models"
)

const (
	DefaultBaseURL = "http://ip-api.com/json/" // https isn't available for free usage
)

// IPAPIClient is an http client for the http://ip-api.com/ API.
type IPAPIClient struct {
	BaseURL string
	Client  *http.Client
}

// IPAPIResponse represent the json response of the http://ip-api.com/ API.
// Documentation can be found at https://ip-api.com/docs/api:json
type IPAPIResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
}

func NewIPAPIClient(baseURL string, client *http.Client) *IPAPIClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &IPAPIClient{BaseURL: baseURL, Client: client}
}

func (c *IPAPIClient) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	// Building http request
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+ip, nil)
	if err != nil {
		return nil, fmt.Errorf("error: error while building http request to %s: %w", c.BaseURL+ip, err)
	}

	// Send http request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error: error while sending http request to %s: %w", c.BaseURL+ip, err)
	}

	// Ensure the response code is 200 OK
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: http response code is not 200 for http request %s: %d", c.BaseURL+ip, resp.StatusCode)
	}

	// Read http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error: error while reading http response to %s: %w", c.BaseURL+ip, err)
	}

	// Unmarshal http response into a IPAPIResponse
	var r IPAPIResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, fmt.Errorf("error: error while unmarsh http request to %s: %w", c.BaseURL+ip, err)
	}

	// Map the IPAPIResponse into a models.GeoIP
	g := &models.GeoIP{
		IP:          ip,
		CountryCode: r.CountryCode,
		CountryName: r.Country,
		City:        r.City,
		Latitude:    r.Lat,
		Longitude:   r.Lon,
	}

	return g, nil
}

// Status will retrieve the status of ip-api.com API.
// It will simply send a GET request to the base URL.
func (c *IPAPIClient) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()

	// Building http request
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL, nil)
	if err != nil {
		ch <- fmt.Errorf("error: error while building http request to %s: %w", c.BaseURL, err)
		return
	}

	// Send http request
	resp, err := c.Client.Do(req)
	if err != nil {
		ch <- fmt.Errorf("error: error while sending http request to %s: %w", c.BaseURL, err)
		return
	}

	// Ensure the response code is 200 OK
	if resp.StatusCode != 200 {
		ch <- fmt.Errorf("error: http response code is not 200 for http request %s: %d", c.BaseURL, resp.StatusCode)
		return
	}

	ch <- nil
}
