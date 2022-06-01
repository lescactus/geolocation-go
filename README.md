# geolocation-go

This repository contains a simple geolocation api microservice, fast, reliable, Kubernetes friendly and ready written in go as a proof of concept.

## Motivations

Study the feasibility of having a geolocation REST API microservice running alongside our other microservices in Kubernetes to avoid relying on the `Cloudfront-Viewer-Country` http headers.
The requirements are:

* Reliability
* Fast
* Concurrent

## Design

`geolocation-go` is written in Go, a fery fast and performant garbaged collected and concurrent programming language.

It expose a simple `GET /rest/v1/{ip}` REST endpoint.

Parameter: 

* `/rest/v1/{ip}` (string) - IPv4 

Response:

* `{"ip":"88.74.7.1","country_code":"DE","country_name":"Germany","city":"Düsseldorf","latitude":51.2217,"longitude":6.77616}`

To retrieve the country code and country name of the given IP address, `geolocation-go` use the [ip-api.com](https://ip-api.com/) real-time Geolocation API, and then cache it in-memory and in Redis for later fast retrievals.

### Flow

```
                                                                                          
                                                                                (2)
                                                                         +--------------> In-memory cache lookup
                                                                         |                       ^ 
                                                                         |                       *
                                              +------------------------+ |                       *
+-------------+            (1)                |                        | |                       * Update in-memory cache
|             |   GET /geolocation/{ip}       |                        | |                       *
|             +------------------------------>|                        | |      (3)              *
|   Client    |                               | geolocation-go         | +--------------> Redis lookup (optional)
|             |          (5)                  |                        | |                       ^ 
|             |<------------------------------+                        | |                       *
+-------------+       200 - OK                |                        | |                       * Update Redis cache
                                              +------------------------+ |                       *
                                                                         |                       *
                                                                         |      (4)              *
                                                                         +--------------> https://api.ipstack.com/{ip} lookup (optional)
```

1) Client make an HTTP request to `/rest/v1/{ip}`

2) `geolocation-go` will lookup for in his in-memory datastore and send the response if cache HIT. In case of cache MISS, go to step 3)

3) `geolocation-go` will lookup in Redis, send the response if cache HIT and add the response in his in-memory datastore asynchronously. In case of cache MISS, go to step 4)

4) `geolocation-go` will make an HTTP call to the [ip-api.com](https://ip-api.com/docs/api:json) API, send back the response to the client and add the response to Redis and the in-memory datastore asynchronously.

## Configuration

`geolocation-go` is a 12-factor app using [Viper](https://github.com/spf13/viper) as a configuration manager. It can read configuration from environment variables or from .env files.

### Available variables

* `APP_ADDR`(default value: `:8080`). Define the TCP address for the server to listen on, in the form "host:port".

* `APP_CONFIG_NAME` (default value: `.env`). Name of the configuration file to read from.

* `APP_CONFIG_PATH` (default value: `.`). Directory containing the configuration file to read from.

* `REDIS_CONNECTION_STRING` (default value `redis://localhost:6379`). Connection string to connect to Redis. The format is the following: `"redis://<user>:<pass>@<host>:<port>/<db>"`.

* `GEOLOCATION_API` (default value `ip-api`). Define which geolocation API to use to retrieve geo IP information. Available options are:
       - [`ip-api`](https://ip-api.com/)

* `IP_API_BASE_URL` (default value: `http://ip-api.com/json/`). Base URL for the [`ip-api`](https://ip-api.com/) API. Note that https is not available with the free plan.

* `HTTP_CLIENT_TIMEOUT` (default value: `15s`). Timeout value for the http client.