package services

import (
	"github.com/go-resty/resty/v2"
)

const (
	SDBASEURL = "http://192.168.50.107:7860/sdapi/v1/txt2img"
	sdengine  = "sd"
)

type SDImageGenerationResponseBody struct {
	Images     []string `json:"images"`
	Parameters struct {
	} `json:"parameters"`
	Info string `json:"info"`
}

type SDT2ImageGenerationRequestBody struct {
	Prompt string `json:"prompt"`
	Steps  int    `json:"steps"`
}

func TrySD(prompt string) (string, error) {
	// Create a Resty Client
	client := resty.New()

	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDImageGenerationResponseBody{}
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDT2ImageGenerationRequestBody{Prompt: prompt, Steps: 5}).
		SetResult(r). // or SetResult(AuthSuccess{}).
		Post(SDBASEURL)
	return r.Images[0], err
}
