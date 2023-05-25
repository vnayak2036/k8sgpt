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

package analyzer

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type NodeStatusAnalyzer struct{}

func (NodeStatusAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "NodeStatus"

	list, err := a.Client.GetClient().CoreV1().Nodes().List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, node := range list.Items {
		var status common.NodeStatus

		nodeRole := getNodeRole(node)
		cpuUsage, memoryUsage := getNodeResourceUsage(a, node.Name)
		tolalPods, workloadPods := getPodDetails(a, node.Name)
		// // Get the number of non-system pods running on the node
		// nonSystemPodsCount := getNonSystemPodsCount(clientset, *nodeName)

		status = addNodeStatus(status, node, nodeRole, cpuUsage, memoryUsage, tolalPods, workloadPods)
		preAnalysis[node.Name] = common.PreAnalysis{
			Node:              node,
			NodeStatusDetails: status,
		}
	}
	//fmt.Println(len(preAnalysis))
	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:             kind,
			Name:             key,
			NodeStatusResult: value.NodeStatusDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Node.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}
	//fmt.Println(a.Results)
	return a.Results, err
}

// Get the role of the node
func getNodeRole(node v1.Node) string {
	// Check for specific labels that indicate the role of the node
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		return "master"
	}
	if _, ok := node.Labels["jumbo-worker"]; ok {
		return "jumbo"
	}
	if _, ok := node.Labels["node-role.kubernetes.io/storage"]; ok {
		return "storage"
	}
	return "worker"
}

func getNodeResourceUsage(a common.Analyzer, nodeName string) (cpuUsage, memoryUsage string) {
	nodeMetrics, _ := a.MetricsClient.GetMetricsClient().MetricsV1beta1().NodeMetricses().Get(a.Context, nodeName, metav1.GetOptions{})
	cpu := nodeMetrics.Usage[v1.ResourceCPU]
	mem := nodeMetrics.Usage[v1.ResourceMemory]
	return cpu.String(), mem.String()
}

func getPodDetails(a common.Analyzer, nodeName string) (totalPodCount, workLoadPodCount int) {
	totalPodList, _ := a.Client.GetClient().CoreV1().Pods("").List(a.Context, metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName)})
	workloadPodCount := 0
	for _, pod := range totalPodList.Items {
		if !strings.HasPrefix(pod.Namespace, "openshift") {
			workloadPodCount++
		}
	}
	return len(totalPodList.Items), workloadPodCount
}

func addNodeStatus(nodestatus common.NodeStatus, node v1.Node, nodeRole string, cpuUsage string, memoryUsage string, totalPoudCount int, workloadPodCount int) common.NodeStatus {
	cpuCapacity := node.Status.Capacity[v1.ResourceCPU]
	memCapacity := node.Status.Capacity[v1.ResourceMemory]
	cpuAllocatable := node.Status.Allocatable[v1.ResourceCPU]
	memAllocatable := node.Status.Allocatable[v1.ResourceMemory]
	nodestatus = common.NodeStatus{
		Name:             node.Name,
		Role:             nodeRole,
		CPUUsage:         cpuUsage,
		MemoryUsage:      memoryUsage,
		CPUCapacity:      cpuCapacity.String(),
		MemCapacity:      memCapacity.String(),
		CPUAllocatable:   cpuAllocatable.String(),
		MemAllocatable:   memAllocatable.String(),
		TotalPodCount:    totalPoudCount,
		WorkloadPodCount: workloadPodCount,
	}
	return nodestatus
}
