package services

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	SDBASEURL = "http://192.168.50.168:7861/sdapi/v1/%s"
	// SDBASEURL = "http://192.168.50.107:7860/sdapi/v1/%s"

	sdengine = "sd"
)

var SDT2IBASEURL = fmt.Sprintf(SDBASEURL, "txt2img")
var SDI2IBASEURL = fmt.Sprintf(SDBASEURL, "img2img")
var SDCLIPBASEURL = fmt.Sprintf(SDBASEURL, "interrogate")

type SDImageGenerationResponseBody struct {
	Images     []string `json:"images"`
	Parameters struct {
	} `json:"parameters"`

	Info string `json:"info"`
}

type SDT2ImageGenerationRequestBody struct {
	Prompt         string `json:"prompt"`
	Steps          int    `json:"steps"`
	NegativePrompt string `json:"negative_prompt"`
	SamplerIndex   string `json:"sampler_index"`
}

type SDI2ImageGenerationRequestBody struct {
	InitImages     []string `json:"init_images"`
	Prompt         string   `json:"prompt"`
	Steps          int      `json:"steps"`
	NegativePrompt string   `json:"negative_prompt"`
	SamplerIndex   string   `json:"sampler_index"`
}

type SDPNGINFORequestBody struct {
	Image string `json:"image"`
	Model string `json:"model"`
}

type SDClipINFOResptBody struct {
	Caption string `json:"caption"`
}

func TrySDT2I(prompt string) (string, error) {
	// Create a Resty Client
	client := resty.New()

	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDImageGenerationResponseBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDT2ImageGenerationRequestBody{Prompt: prompt + "masterpiece, best quality, ultra-detailed", NegativePrompt: "EasyNegative", SamplerIndex: "DDIM", Steps: 50}).
		SetResult(r).
		Post(SDT2IBASEURL)
	if err != nil {
		return "", err
	} else {
		if len(r.Images) == 0 {
			return "", errors.New("sd resp: " + resp.String())
		}
		fmt.Println("bs64 str", r.Images[0][0:20])
		return r.Images[0], nil
	}
}

func TrySDI2I(bs64, prompt string) (string, error) {
	// Create a Resty Client
	client := resty.New()
	imgs := []string{bs64}
	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDImageGenerationResponseBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDI2ImageGenerationRequestBody{InitImages: imgs, Prompt: prompt + "masterpiece, best quality, ultra-detailed", NegativePrompt: "EasyNegative", SamplerIndex: "DDIM", Steps: 50}).
		SetResult(r).
		Post(SDI2IBASEURL)
	if err != nil {
		return "", err
	} else {
		if len(r.Images) == 0 {
			return "", errors.New("sd resp: " + resp.String())
		}
		fmt.Println("bs64 str", r.Images[0][0:20])
		return r.Images[0], nil
	}

}

func TryCLIPINFO(bs64 string) (string, error) {
	// Create a Resty Client
	client := resty.New()
	// POST JSON string
	// No need to set content type, if you have client level setting
	r := &SDClipINFOResptBody{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(SDPNGINFORequestBody{Image: bs64, Model: "clip"}).
		SetResult(r).
		Post(SDCLIPBASEURL)
	if err != nil {
		return resp.String(), err
	} else {
		if r.Caption != "" {
			return r.Caption, nil
		} else {
			return resp.String(), err
		}

	}

}
