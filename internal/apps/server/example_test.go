package server_test

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	host = "http://localhost:8080"
)

func ExamplePingDB() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.
		R().
		Get("/ping")

	print(response, err)
}

func ExampleUpdatesByJSON() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"metrics":[{"type":"gauge","name":"cpu_usage","value":0.5}]}`).
		Post("/updates/")

	print(response, err)
}

func ExampleUpdateByJSON() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"type":"gauge","name":"cpu_usage","value":0.5}`).
		Post("/update/")

	print(response, err)
}

func ExampleUpdateByParams() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.R().
		Post("/update/gauge/cpu_usage/0.5")

	print(response, err)
}

func ExampleGetByParam() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.
		R().
		Get("/value/gauge/cpu_usage")

	print(response, err)
}

func ExampleGetByJSON() {
	client := resty.New()
	client.SetBaseURL(host)
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"type":"cpu","name":"usage"}`).
		Post("/value/")

	print(response, err)
}

func print(response *resty.Response, err error) {
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("status Code: %d\n", response.StatusCode())
	fmt.Printf("body: %s\n", response.Body())
}
