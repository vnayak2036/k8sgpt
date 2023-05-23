package ai

const (
	default_prompt = `Explain the following Kubernetes error message and provide the steps for a solution in %s: %s.  
	                  Also provide sample yaml code where possible`

	node_usage_prompt = `Given the below openshift nodes details in double quotes, do the below tasks. In the below data each node can be identified by its name which starts with "{Name:ip-...}". Start analysis after you have data available on all 45 nodes.
	                    1) List total number of nodes with Role:Master. 
						2) List total number of nodes with Role:Worker.
	                    3) Using CPU and Memory Usage, Capacity and Allocatble data list 3 least used nodes. Exclude nodes with Role:Master in the analysis. 
					    4) Identify the least used node with reasoning which includes current usage in GB and cores. Exclude nodes with Role:Master in the analysis.
						5) Can the number of workload pods in the least use node be move to other nodes ? Is there spare capacity in the cluster excluding nodes with Role:Master ? Explain
						6) Identify few nodes that can take on the workloadPods of the least used node.  Exclude nodes with Role:Master in the analysis. Roles should match.
						" Node Details: %s"`
)
