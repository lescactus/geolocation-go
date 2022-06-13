package ipbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	DefaultBaseURL = "https://api.ipbase.com/v2/info?ip="
)

// Prometheus metrics
var (
	ipBaseSuccessRequestSend = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ipbase_http_requests_success_total",
		Help: "Total number of successful http requests sent to the ipbase API",
	})
	ipBaseFailedRequestSend = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ipbase_http_requests_failed_total",
		Help: "Total number of failed http requests sent to the ipbase API",
	})
)

// IPBaseClient is an http client for the https://api.ipbase.com/ API.
type IPBaseClient struct {
	BaseURL string
	apiKey  string
	Client  *http.Client
	Logger  *zerolog.Logger
}

// IPBaseResponse represents the json response of the ipbase.com API
// Documentation can be found at https://ipbase.com/docs/info
type IPBaseResponse struct {
	Data IPBaseResponseData `json:"data"`
}

type IPBaseResponseData struct {
	Timezone   IPBaseResponseDataTimeZone   `json:"timezone"`
	IP         string                       `json:"ip"`
	Type       string                       `json:"type"`
	Connection IPBaseResponseDataConnection `json:"connection"`
	Location   IPBaseResponseDataLocation   `json:"location"`
}

type IPBaseResponseDataTimeZone struct {
	ID               string    `json:"id"`
	CurrentTime      time.Time `json:"current_time"`
	Code             string    `json:"code"`
	IsDaylightSaving bool      `json:"is_daylight_saving"`
	GmtOffset        int       `json:"gmt_offset"`
}

type IPBaseResponseDataConnection struct {
	Asn          int    `json:"asn"`
	Organization string `json:"organization"`
	Isp          string `json:"isp"`
}

type IPBaseResponseDataLocation struct {
	GeonamesID int                                 `json:"geonames_id"`
	Latitude   float64                             `json:"latitude"`
	Longitude  float64                             `json:"longitude"`
	Zip        string                              `json:"zip"`
	Continent  IPBaseResponseDataLocationContinent `json:"continent"`
	Country    IPBaseResponseDataLocationCountry   `json:"country"`
	City       IPBaseResponseDataLocationCity      `json:"city"`
	Region     IPBaseResponseDataLocationRegion    `json:"region"`
}

type IPBaseResponseDataLocationContinent struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
}

type IPBaseResponseDataLocationCity struct {
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
}

type IPBaseResponseDataLocationRegion struct {
	Fips           string `json:"fips"`
	Alpha2         string `json:"alpha2"`
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
}

type IPBaseResponseDataLocationCountry struct {
	Alpha2       string   `json:"alpha2"`
	Alpha3       string   `json:"alpha3"`
	CallingCodes []string `json:"calling_codes"`
	Currencies   []struct {
		Symbol        string `json:"symbol"`
		Name          string `json:"name"`
		SymbolNative  string `json:"symbol_native"`
		DecimalDigits int    `json:"decimal_digits"`
		Rounding      int    `json:"rounding"`
		Code          string `json:"code"`
		NamePlural    string `json:"name_plural"`
	} `json:"currencies"`
	Emoji     string `json:"emoji"`
	Ioc       string `json:"ioc"`
	Languages []struct {
		Name       string `json:"name"`
		NameNative string `json:"name_native"`
	} `json:"languages"`
	Name              string   `json:"name"`
	NameTranslated    string   `json:"name_translated"`
	Timezones         []string `json:"timezones"`
	IsInEuropeanUnion bool     `json:"is_in_european_union"`
}

func NewIPBaseClient(baseURL, apikey string, client *http.Client, logger *zerolog.Logger) *IPBaseClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &IPBaseClient{BaseURL: baseURL, apiKey: apikey, Client: client, Logger: logger}
}

func (c *IPBaseClient) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	// Get request id for logging purposes
	req_id, _ := hlog.IDFromCtx(ctx)

	// Building url
	url := fmt.Sprintf("%s%s&apikey=%s&language=en", c.BaseURL, ip, c.apiKey)
	urlRedacted := fmt.Sprintf("%s%s&apikey=%s&language=en", c.BaseURL, ip, "xxxxxx")

	// Building http request
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("building http request to " + urlRedacted + ip)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while building http request to %s: %s", urlRedacted, err.Error()))
		return nil, fmt.Errorf("error: error while building http request to %s: %w", urlRedacted, err)
	}

	// Send http request
	c.Logger.Debug().Str("req_id", req_id.String()).Msg("sending http request to " + urlRedacted + ip)
	resp, err := c.Client.Do(req)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while sending http request to %s: %s", urlRedacted, err.Error()))
		// Increment Prometheus counter
		ipBaseFailedRequestSend.Inc()
		return nil, fmt.Errorf("error: error while sending http request to %s: %w", urlRedacted, err)
	}

	// Increment Prometheus counter
	ipBaseSuccessRequestSend.Inc()

	// Ensure the response code is 200 OK
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("http request to " + urlRedacted + " sent")
	if resp.StatusCode != 200 {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("http response code is not 200 for http request %s: %d", urlRedacted, resp.StatusCode))
		return nil, fmt.Errorf("error: http response code is not 200 for http request %s: %d", urlRedacted, resp.StatusCode)
	}

	// Read http response
	c.Logger.Trace().Str("req_id", req_id.String()).Msg("reading http response from " + urlRedacted)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while reading http response to %s: %s", urlRedacted, err.Error()))
		return nil, fmt.Errorf("error: error while reading http response to %s: %w", urlRedacted, err)
	}

	// Unmarshal http response into a IPAPIResponse
	var r IPBaseResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		c.Logger.Error().Str("req_id", req_id.String()).
			Msg(fmt.Sprintf("error while unmarshalling http request to %s: %s", urlRedacted, err))
		return nil, fmt.Errorf("error: error while unmarshalling http request to %s: %w", urlRedacted, err)
	}

	// Map the IPAPIResponse into a models.GeoIP
	g := &models.GeoIP{
		IP:          ip,
		CountryCode: r.Data.Location.Country.Alpha2,
		CountryName: r.Data.Location.Country.Name,
		City:        r.Data.Location.City.Name,
		Latitude:    r.Data.Location.Latitude,
		Longitude:   r.Data.Location.Longitude,
	}

	return g, nil
}
