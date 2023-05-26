/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/sashabaranov/go-openai"

	"github.com/fatih/color"
)

type OpenAIClient struct {
	client   *openai.Client
	language string
	model    string
}

func (c *OpenAIClient) Configure(config IAIConfig, language string) error {
	token := config.GetPassword()
	defaultConfig := openai.DefaultConfig(token)

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.language = language
	c.client = client
	//c.model = config.GetModel()
	c.model = openai.GPT4
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, prompt string, prompt_name ...string) (string, error) {

	// Create a completion request
	// var cont = fmt.Sprintf(default_prompt, c.language, prompt)
	// println(cont)
	pname := "default_prompt"
	if len(prompt_name) > 0 {
		pname = prompt_name[0]
	}
	var content string
	switch pname {
	case "node_resource_prompt":
		nodeCount := strings.Count(prompt, "Name:")
		content = fmt.Sprintf(node_usage_prompt, nodeCount, prompt)
	default:
		content = fmt.Sprintf(default_prompt, c.language, prompt)
	}
	fmt.Println("****** Prompt Content is **********")
	fmt.Println(content)
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
		Temperature: 0,
	})

	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (a *OpenAIClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, kind ...string) (string, error) {

	var kd, prompt_name string
	if len(kind) > 0 {
		kd = kind[0]
	}
	inputKey := strings.Join(prompt, " ")

	// Check for cached data
	cacheKey := util.GetCacheKey(a.GetName(), a.language, inputKey)

	if !cache.IsCacheDisabled() && cache.Exists(cacheKey) {
		response, err := cache.Load(cacheKey)
		if err != nil {
			return "", err
		}

		if response != "" {
			output, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				color.Red("error decoding cached data: %v", err)
				return "", nil
			}
			return string(output), nil
		}
	}
	switch kd {
	case "NodeStatus":
		prompt_name = "node_resource_prompt"
		break
	}
	response, err := a.GetCompletion(ctx, inputKey, prompt_name)
	if err != nil {
		return "", err
	}
	err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}
	return response, nil
}

func (a *OpenAIClient) GetName() string {
	return "openai"
}
