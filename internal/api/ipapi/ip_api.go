package ipapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	DefaultBaseURL = "http://ip-api.com/json/" // https isn't available for free usage
)

// Prometheus metrics
var (
	ipAPISuccessRequestSend = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_api_http_requests_success_total",
		Help: "Total number of successful http requests sent to the ip-api API",
	})
	ipAPIFailedRequestSend = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_api_http_requests_failed_total",
		Help: "Total number of failed http requests sent to the ip-api API",
	})
)

// IPAPIClient is an http client for the http://ip-api.com/ API.
type IPAPIClient struct {
	BaseURL string
	Client  *http.Client
	Logger  *zerolog.Logger
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

func NewIPAPIClient(baseURL string, client *http.Client, logger *zerolog.Logger) *IPAPIClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &IPAPIClient{BaseURL: baseURL, Client: client, Logger: logger}
}

func (c *IPAPIClient) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(ctx)

	// Building http request
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("building http request to " + c.BaseURL + ip)
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+ip, nil)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while building http request to %s: %s", c.BaseURL+ip, err.Error()))
		return nil, fmt.Errorf("error: error while building http request to %s: %w", c.BaseURL+ip, err)
	}

	// Send http request
	c.Logger.Debug().Str("req_id", req_id.String()).Msg("sending http request to " + c.BaseURL + ip)
	resp, err := c.Client.Do(req)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while sending http request to %s: %s", c.BaseURL+ip, err.Error()))
		// Increment Prometheus counter
		ipAPIFailedRequestSend.Inc()
		return nil, fmt.Errorf("error: error while sending http request to %s: %w", c.BaseURL+ip, err)
	}

	// Increment Prometheus counter
	ipAPISuccessRequestSend.Inc()

	// Ensure the response code is 200 OK
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("http request to " + c.BaseURL + ip + " sent")
	if resp.StatusCode != 200 {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("http response code is not 200 for http request %s: %d", c.BaseURL+ip, resp.StatusCode))
		return nil, fmt.Errorf("error: http response code is not 200 for http request %s: %d", c.BaseURL+ip, resp.StatusCode)
	}

	// Read http response
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("reading http response from " + c.BaseURL + ip)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while reading http response to %s: %s", c.BaseURL+ip, err.Error()))
		return nil, fmt.Errorf("error: error while reading http response to %s: %w", c.BaseURL+ip, err)
	}

	// Unmarshal http response into a IPAPIResponse
	var r IPAPIResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while unmarshalling http request to %s: %s", c.BaseURL+ip, err))
		return nil, fmt.Errorf("error: error while unmarshalling http request to %s: %w", c.BaseURL+ip, err)
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
