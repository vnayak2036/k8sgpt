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

package analysis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

type Analysis struct {
	Context        context.Context
	Filters        []string
	Client         *kubernetes.Client
	MetricsClient  *kubernetes.MetricsClient
	AIClient       ai.IAI
	Results        []common.Result
	Errors         []string
	Namespace      string
	Cache          cache.ICache
	Explain        bool
	MaxConcurrency int
}

type AnalysisStatus string
type AnalysisErrors []string

const (
	StateOK              AnalysisStatus = "OK"
	StateProblemDetected AnalysisStatus = "ProblemDetected"
)

type JsonOutput struct {
	Errors   AnalysisErrors  `json:"errors"`
	Status   AnalysisStatus  `json:"status"`
	Problems int             `json:"problems"`
	Results  []common.Result `json:"results"`
}

func NewAnalysis(backend string, language string, filters []string, namespace string, noCache bool, explain bool, maxConcurrency int) (*Analysis, error) {
	var configAI ai.AIConfiguration
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if len(configAI.Providers) == 0 && explain {
		color.Red("Error: AI provider not specified in configuration. Please run k8sgpt auth")
		os.Exit(1)
	}

	var aiProvider ai.AIProvider
	for _, provider := range configAI.Providers {
		if backend == provider.Name {
			aiProvider = provider
			break
		}
	}

	if aiProvider.Name == "" {
		color.Red("Error: AI provider %s not specified in configuration. Please run k8sgpt auth", backend)
		return nil, errors.New("AI provider not specified in configuration")
	}

	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider, language); err != nil {
		color.Red("Error: %v", err)
		return nil, err
	}

	ctx := context.Background()
	// Get kubernetes client from viper

	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		return nil, err
	}

	metricsClient, err := kubernetes.NewMetricsClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes metrics client: %v", err)
		return nil, err
	}

	return &Analysis{
		Context:        ctx,
		Filters:        filters,
		Client:         client,
		MetricsClient:  metricsClient,
		AIClient:       aiClient,
		Namespace:      namespace,
		Cache:          cache.New(noCache),
		Explain:        explain,
		MaxConcurrency: maxConcurrency,
	}, nil
}

func (a *Analysis) RunAnalysis() {
	activeFilters := viper.GetStringSlice("active_filters")

	analyzerMap := analyzer.GetAnalyzerMap()

	analyzerConfig := common.Analyzer{
		Client:        a.Client,
		MetricsClient: a.MetricsClient,
		Context:       a.Context,
		Namespace:     a.Namespace,
		AIClient:      a.AIClient,
	}

	semaphore := make(chan struct{}, a.MaxConcurrency)
	// if there are no filters selected and no active_filters then run all of them
	if len(a.Filters) == 0 && len(activeFilters) == 0 {
		var wg sync.WaitGroup
		var mutex sync.Mutex
		for _, analyzer := range analyzerMap {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(analyzer common.IAnalyzer, wg *sync.WaitGroup, semaphore chan struct{}) {
				defer wg.Done()
				results, err := analyzer.Analyze(analyzerConfig)
				if err != nil {
					mutex.Lock()
					a.Errors = append(a.Errors, fmt.Sprintf(fmt.Sprintf("[%s] %s", reflect.TypeOf(analyzer).Name(), err)))
					mutex.Unlock()
				}
				mutex.Lock()
				a.Results = append(a.Results, results...)
				mutex.Unlock()
				<-semaphore
			}(analyzer, &wg, semaphore)

		}
		wg.Wait()
	}
	semaphore = make(chan struct{}, a.MaxConcurrency)
	// if the filters flag is specified
	if len(a.Filters) != 0 {
		fmt.Println("Running Analysis")
		var wg sync.WaitGroup
		var mutex sync.Mutex
		for _, filter := range a.Filters {
			if analyzer, ok := analyzerMap[filter]; ok {
				semaphore <- struct{}{}
				wg.Add(1)
				go func(analyzer common.IAnalyzer, filter string) {
					defer wg.Done()
					results, err := analyzer.Analyze(analyzerConfig)
					if err != nil {
						mutex.Lock()
						a.Errors = append(a.Errors, fmt.Sprintf(fmt.Sprintf("[%s] %s", filter, err)))
						mutex.Unlock()
					}
					mutex.Lock()
					a.Results = append(a.Results, results...)
					mutex.Unlock()
					<-semaphore
				}(analyzer, filter)
			} else {
				a.Errors = append(a.Errors, fmt.Sprintf(fmt.Sprintf("\"%s\" filter does not exist. Please run k8sgpt filters list.", filter)))
			}
		}
		wg.Wait()
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	semaphore = make(chan struct{}, a.MaxConcurrency)
	// use active_filters
	for _, filter := range activeFilters {
		if analyzer, ok := analyzerMap[filter]; ok {
			semaphore <- struct{}{}
			wg.Add(1)
			go func(analyzer common.IAnalyzer, filter string) {
				defer wg.Done()
				results, err := analyzer.Analyze(analyzerConfig)
				if err != nil {
					mutex.Lock()
					a.Errors = append(a.Errors, fmt.Sprintf("[%s] %s", filter, err))
					mutex.Unlock()
				}
				mutex.Lock()
				a.Results = append(a.Results, results...)
				mutex.Unlock()
				<-semaphore
			}(analyzer, filter)
		}
	}
	wg.Wait()
}

func (a *Analysis) GetAIResults(output string, anonymize bool) error {
	if len(a.Results) == 0 {
		return nil
	}
	var bar *progressbar.ProgressBar
	if output != "json" {
		bar = progressbar.Default(int64(len(a.Results)))
	}
    var texts []string
	var kind string
	var index int
	var analysis common.Result
	for index, analysis = range a.Results {
		
		//var prompt_data interface{}
		kind = analysis.Kind
		failure := analysis.NodeStatusResult
		status := fmt.Sprintf("%+v", failure)
		texts = append(texts, status)
	}
		parsedText, err := a.AIClient.Parse(a.Context, texts, a.Cache, kind)
		if err != nil {
			// FIXME: can we avoid checking if output is json multiple times?
			//   maybe implement the progress bar better?
			if output != "json" {
				bar.Exit()
			}

			// Check for exhaustion
			if strings.Contains(err.Error(), "status code: 429") {
				return fmt.Errorf("exhausted API quota for AI provider %s: %v", a.AIClient.GetName(), err)
			} else {
				return fmt.Errorf("failed while calling AI provider %s: %v", a.AIClient.GetName(), err)
			}
		}

		if anonymize {
			for _, failure := range analysis.Error {
				for _, s := range failure.Sensitive {
					parsedText = strings.ReplaceAll(parsedText, s.Masked, s.Unmasked)
				}
			}
		}

		analysis.Details = parsedText
		if output != "json" {
			bar.Add(1)
		}
		a.Results[index] = analysis
	
	return nil
}
