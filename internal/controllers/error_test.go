package controllers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/geolocation-go/internal/chain"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want *ErrorResponse
	}{
		{
			name: "Non empty message",
			args: args{"this is an error"},
			want: &ErrorResponse{Status: "error", Msg: "this is an error"},
		},
		{
			name: "Empty message",
			args: args{""},
			want: &ErrorResponse{Status: "error", Msg: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewErrorResponse(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseHandlerNotFoundHandler(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []byte
		code int
	}{
		{
			name: "/nonexistent",
			path: "/nonexistent",
			want: []byte(`{"status":"error","msg":"404 page not found"}`),
			code: http.StatusNotFound,
		},
	}

	r := httprouter.New()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	a := &GeoAPIMock{}
	c := chain.New(&logger)
	c.Add("in-memory", mdb)
	c.Add("redis", rdb)

	// route registration
	h := NewBaseHandler(c, a, &logger)
	r.Handler("GET", "/rest/v1/:ip", http.HandlerFunc(http.HandlerFunc(h.GetGeoIP)))
	r.NotFound = http.HandlerFunc(h.NotFoundHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			resp := recorder.Result()
			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, data)
			assert.Equal(t, tt.code, resp.StatusCode)
		})
	}
}

func TestBaseHandlerMethodNotAllowedHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   []byte
		code   int
	}{
		{
			name:   "invalid method",
			method: "POST",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid method",
			method: "PUT",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "valid method",
			method: "GET",
			want:   []byte(`{"ip":"1.1.1.1","country_code":"AU","country_name":"Australia","city":"South Brisbane","latitude":-27.4766,"longitude":153.0166}`),
			code:   http.StatusOK,
		},
	}

	r := httprouter.New()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	a := &GeoAPIMock{}
	c := chain.New(&logger)
	c.Add("in-memory", mdb)
	c.Add("redis", rdb)

	// route registration
	h := NewBaseHandler(c, a, &logger)
	r.Handler("GET", "/rest/v1/:ip", http.HandlerFunc(http.HandlerFunc(h.GetGeoIP)))
	r.MethodNotAllowed = http.HandlerFunc(h.MethodNotAllowedHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/rest/v1/1.1.1.1", nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			resp := recorder.Result()
			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, data)
			assert.Equal(t, tt.code, resp.StatusCode)
		})
	}
}
