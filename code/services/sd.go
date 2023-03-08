package services

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	SDT2IBASEURL = "http://192.168.50.107:7860/sdapi/v1/txt2img"
	SDI2IBASEURL = "http://192.168.50.107:7860/sdapi/v1/img2img"
	sdengine     = "sd"
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

type SDI2ImageGenerationRequestBody struct {
	InitImages      []string `json:"init_images"`
	Prompt          string   `json:"prompt"`
	Steps           int      `json:"steps"`
	SeedResizeFromH int      `json:"seed_resize_from_h"`
	SeedResizeFromW int      `json:"seed_resize_from_w"`
}

func TrySDT2I(prompt string) (string, error) {
	// Create a Resty Client
	client := resty.New()

	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDImageGenerationResponseBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDT2ImageGenerationRequestBody{Prompt: prompt, Steps: 5}).
		SetResult(r).
		Post(SDT2IBASEURL)
	if err != nil {
		return "", err
	} else {
		if len(r.Images) == 0 {
			return "", errors.New("sd resp: " + resp.String())
		}
		return r.Images[0], nil
	}
}

func TrySDI2I(bs64, prompt string) (string, error) {
	// Create a Resty Client
	fmt.Println("send sd promt:", prompt)
	client := resty.New()
	imgs := []string{bs64}
	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDImageGenerationResponseBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDI2ImageGenerationRequestBody{InitImages: imgs, Prompt: prompt, Steps: 20, SeedResizeFromH: 512, SeedResizeFromW: 512}).
		SetResult(r).
		Post(SDI2IBASEURL)
	if err != nil {
		return "", err
	} else {
		if len(r.Images) == 0 {
			return "", errors.New("sd resp: " + resp.String())
		}
		return r.Images[0], nil
	}

}
