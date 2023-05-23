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

package common

import (
	"context"

	trivy "github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
)

type IAnalyzer interface {
	Analyze(analysis Analyzer) ([]Result, error)
}

type Analyzer struct {
	Client          *kubernetes.Client
	MetricsClient   *kubernetes.MetricsClient
	Context         context.Context
	Namespace       string
	AIClient        ai.IAI
	PreAnalysis     map[string]PreAnalysis
	Results         []Result
}

type PreAnalysis struct {
	Pod                      v1.Pod
	FailureDetails           []Failure
	Deployment               appsv1.Deployment
	ReplicaSet               appsv1.ReplicaSet
	PersistentVolumeClaim    v1.PersistentVolumeClaim
	Endpoint                 v1.Endpoints
	Ingress                  networkv1.Ingress
	HorizontalPodAutoscalers autov1.HorizontalPodAutoscaler
	PodDisruptionBudget      policyv1.PodDisruptionBudget
	StatefulSet              appsv1.StatefulSet
	NetworkPolicy            networkv1.NetworkPolicy
	Node                     v1.Node
	// Integrations
	TrivyVulnerabilityReport trivy.VulnerabilityReport
	NodeStatusDetails        NodeStatus
}


type Result struct {
	Kind              string    `json:"kind"`
	Name              string    `json:"name"`
	Error             []Failure `json:"error"`
	Details           string    `json:"details"`
	ParentObject      string    `json:"parentObject"`
	NodeStatusResult  NodeStatus `json:"nodeStatus"`
}

type Failure struct {
	Text      string
	Sensitive []Sensitive
}


// type NodeStatus struct {	        
// 	//Conditions               []v1.NodeCondition       `json:"conditions,omitempty"`
// 	Allocatable              v1.ResourceList          `json:"allocatable,omitempty"`
// 	// Capacity                 v1.ResourceList          `json:"capacity,omitempty"`
// 	// Phase                    v1.NodePhase             `json:"phase,omitempty"`
// 	// HostIP                   string                   `json:"hostIP,omitempty"`
// 	// PodCIDR                  string                   `json:"podCIDR,omitempty"`
// 	// PodCIDRs                 []string                 `json:"podCIDRs,omitempty"`
// 	// PodCIDRSize              int                      `json:"podCIDRSize,omitempty"`
// 	// StartTime                *metav1.Time             `json:"startTime,omitempty"`
// 	// Unschedulable            bool                     `json:"unschedulable,omitempty"`
// 	// Addresses                []v1.NodeAddress         `json:"addresses,omitempty"`
// 	// DaemonEndpoints          v1.NodeDaemonEndpoints   `json:"daemonEndpoints,omitempty"`
// 	// NodeInfo                 v1.NodeSystemInfo        `json:"nodeInfo,omitempty"`
// 	// VolumesAttached          []v1.AttachedVolume      `json:"volumesAttached,omitempty"`
// 	// VolumesInUse             []v1.UniqueVolumeName    `json:"volumesInUse,omitempty"`
// 	// ConfigSource             v1.NodeConfigSource     `json:"configSource,omitempty"`
// 	// Sensitive                []Sensitive	
// }


type NodeStatus struct {	
	Name						     string        
	Role                             string
	CPUUsage					     string
	MemoryUsage                      string
	CPUCapacity                      string
	MemCapacity                      string
	CPUAllocatable                   string
	MemAllocatable                   string
	TotalPodCount						 int
	WorkloadPodCount				 int

}
type Sensitive struct {
	Unmasked string
	Masked   string
}
