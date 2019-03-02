package constant

import (
	k8sCore "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	// LabelArhatNodeType node type label
	// the label name is `arhat.node.kubernetes.io/type`
	LabelArhatNodeType = "arhat." + k8sCore.LabelNamespaceSuffixNode + "/type"
)

type NodeType string

const (
	// NodeTypeServer the node is a server
	NodeTypeServer = NodeType("server")
	// NodeTypeDevice the node is a IoT device
	NodeTypeDevice = NodeType("device")
)
