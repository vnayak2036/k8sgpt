package ai

const (
	default_prompt = `Explain the following Kubernetes error message and provide the steps for a solution in %s: %s.  
	                  Also provide sample yaml code where possible`

	node_usage_prompt = `Given the below openshift nodes details in double quotes, do the below tasks. 
    In the below data each node can be identified by its name which starts with "{Name:ip-...}". Every node has a Role associated with it, 
    like Role:master or Role:worker or Role:storage or Role:jumbo. Every node only has one role. 
    Start analysis after you have data available on all %d nodes. 
    1) List number of nodes with Role:master.  
    2) List number of nodes with Role:storage.
    3) Using CPU usage, Memory usage, Capacity and Allocatable data, list 3 least used nodes. 
       Only exclude nodes with Role:master and nodes with Role:storage in the analysis. List the role of each node.
    4) Identify the least used node with reasoning that includes current usage of CPU in cores and memory in GB. 
       Also take workloadPod count into account. Ideally the least used node should also have a low workload pod count.
       Only exclude nodes with Role:master and nodes with Role:storage in the analysis. List the node role.
    5) Can the number of workload pods in the least used 3 nodes be moved to other nodes ? 
       Is there spare capacity in the cluster after excluding nodes with Role:master and  nodes with Role:storage ? Explain
    6) Identify few nodes that can take on the workloadPods of the 3 least used nodes. 
       Exclude nodes with Role:master and nodes with Role:storage in the analysis.
       Roles of the suggested nodes should match the roles of the least used nodes. 
       List the role of each suggested node.

    "Node Details: %s"`
)
