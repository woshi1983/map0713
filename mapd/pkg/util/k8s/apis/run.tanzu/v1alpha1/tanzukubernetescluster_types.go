// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Important: Run "make" to regenerate code after modifying this file

// TanzuKubernetesClusterAddonStatus is used to type the constants describing possible addon states.
type TanzuKubernetesClusterAddonStatus string

const (
	// TanzuKubernetesClusterAddonsStatusPending means addon not yet applied and
	// no error.
	TanzuKubernetesClusterAddonsStatusPending = TanzuKubernetesClusterAddonStatus("pending")

	// TanzuKubernetesClusterAddonsStatusError means an error happened.
	TanzuKubernetesClusterAddonsStatusError = TanzuKubernetesClusterAddonStatus("error")

	// TanzuKubernetesClusterAddonsStatusApplied means that the addon was
	// successfully applied.
	TanzuKubernetesClusterAddonsStatusApplied = TanzuKubernetesClusterAddonStatus("applied")

	// TanzuKubernetesClusterAddonsStatusUnmanaged means that the addon is not
	// being managed by the addons controller and the status is unknown. This
	// should not be considered as an error nor should it block rolling updates.
	TanzuKubernetesClusterAddonsStatusUnmanaged = TanzuKubernetesClusterAddonStatus("unmanaged")
)

// TanzuKubernetesClusterPhase is a type for the Tanzu Kubernetes cluster's
// phase constants.
type TanzuKubernetesClusterPhase string

const (
	// TanzuKubernetesClusterPhaseCreating means that the cluster is under creation.
	TanzuKubernetesClusterPhaseCreating = TanzuKubernetesClusterPhase("creating")

	// TanzuKubernetesClusterPhaseFailed means that cluster creation failed.
	// The system likely requires user intervention.
	TanzuKubernetesClusterPhaseFailed = TanzuKubernetesClusterPhase("failed")

	// TanzuKubernetesClusterPhaseRunning means that the cluster control plane,
	// add-ons and workers are ready.
	TanzuKubernetesClusterPhaseRunning = TanzuKubernetesClusterPhase("running")

	// TanzuKubernetesClusterPhaseUnhealthy means that the cluster was up and running,
	// but unhealthy now, the system likely requires user intervention.
	TanzuKubernetesClusterPhaseUnhealthy = TanzuKubernetesClusterPhase("unhealthy")

	// TanzuKubernetesClusterPhaseUpdating indicates that the cluster is in the
	// process of rolling update
	TanzuKubernetesClusterPhaseUpdating = TanzuKubernetesClusterPhase("updating")

	// TanzuKubernetesClusterPhaseUpdateFailed indicates that the cluster's
	// rolling update failed and likely requires user intervention.
	TanzuKubernetesClusterPhaseUpdateFailed = TanzuKubernetesClusterPhase("updateFailed")

	// TanzuKubernetesClusterPhaseDeleting means that the cluster is being
	// deleted.
	TanzuKubernetesClusterPhaseDeleting = TanzuKubernetesClusterPhase("deleting")

	// TanzuKubernetesClusterPhaseEmpty is useful for the initial reconcile,
	// before we even state the phase as creating.
	TanzuKubernetesClusterPhaseEmpty = TanzuKubernetesClusterPhase("")
)

// NodeStatus is used to type the constants describing possible node states.
type NodeStatus string

const (
	// NodeStatusPending means that the node has not yet joined.
	NodeStatusPending = NodeStatus("pending")

	// NodeStatusUnknown means that the node is joined, but returning Ready Condition Unknown.
	NodeStatusUnknown = NodeStatus("unknown")

	// NodeStatusReady means that the node is joined, but returning Ready Condition True.
	NodeStatusReady = NodeStatus("ready")

	// NodeStatusNotReady means that the node is joined, but returning Ready Condition False.
	NodeStatusNotReady = NodeStatus("notready")
)

// TanzuKubernetesClusterSpec defines the desired state of TanzuKubernetesCluster: its nodes, the software installed on those nodes and
// the way that software should be configured.
type TanzuKubernetesClusterSpec struct {
	// Topology specifies the topology for the Tanzu Kubernetes cluster: the number, purpose, and organization of the nodes which
	// form the cluster and the resources allocated for each.
	Topology Topology `json:"topology"`

	// Distribution specifies the distribution for the Tanzu Kubernetes cluster: the software installed on the control plane and
	// worker nodes, including Kubernetes itself.
	Distribution Distribution `json:"distribution"`

	// Settings specifies settings for the Tanzu Kubernetes cluster: the way an instance of a distribution is configured,
	// including information about pod networking and storage.
	// +optional
	Settings *Settings `json:"settings,omitempty"`
}

// Volume defines a PVC attachment.
// These volumes are tied to the node lifecycle, created and deleted when the node is.
// The volumes are mounted in the node during the bootstrap process, prior to services being started (e.g. etcd, containerd).
type Volume struct {
	// Name is suffix used to name this PVC as: node.Name + "-" + Name
	Name string `json:"name"`
	// MountPath is the directory where the volume device is to be mounted
	MountPath string `json:"mountPath"`
	// Capacity is the PVC capacity
	Capacity corev1.ResourceList `json:"capacity"`
	// StorageClass is the storage class to be used for the disks.
	// Defaults to TopologySettings.StorageClass
	// +optional
	StorageClass string `json:"storageClass,omitempty"`
}

// Topology describes the number, purpose, and organization of nodes and the resources allocated for each. Nodes are
// grouped into pools based on their intended purpose. Each pool is homogeneous, having the same resource allocation and
// using the same storage.
type Topology struct {
	// ControlPlane specifies the topology of the cluster's control plane, including the number of control plane nodes
	// and resources allocated for each. The control plane must have an odd number of nodes.
	ControlPlane TopologySettings `json:"controlPlane"`

	// Workers specifies the topology of cluster's worker nodes, including the number of worker nodes and resources
	// allocated for each.
	Workers TopologySettings `json:"workers"`
}

// TopologySettings describes a homogeneous pool of nodes: the number of nodes in the pool and the properties of each of
// those nodes, including resource allocation and storage.
type TopologySettings struct {
	// Count is the number of nodes.
	Count int32 `json:"count"`

	// Class is the name of the VirtualMachineClass, which describes the virtual hardware settings, to be used each node
	// in the pool. This controls the hardware available to the node (CPU and memory) as well as the requests and limits
	// on those resources. Run `kubectl describe virtualmachineclasses` to see which VM classes are available to use.
	Class string `json:"class"`

	// StorageClass is the storage class to be used for storage of the disks which store the root filesystems of the
	// nodes. Run `kubectl describe ns` on your namespace to see which storage classes are available to use.
	StorageClass string `json:"storageClass"`

	// Volumes is the set of PVCs to be created and attached to each node.
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`
}

// Distribution specifies the version of software which should be installed on the control plane and worker nodes. This
// version information encompasses Kubernetes and its dependencies, the base OS of the node, and add-ons.
type Distribution struct {
	// Version specifies the fully-qualified desired Kubernetes distribution version of the Tanzu Kubernetes cluster. If the
	// cluster exists and is not of the specified version, it will be upgraded.
	//
	// Version is a semantic version string. The version may not be decreased. The major version may not be changed. If
	// the minor version is changed, it may only be incremented; skipping minor versions is not supported.
	//
	// The current observed version of the cluster is held by `status.version`.
	// +optional
	Version string `json:"fullVersion"`

	// VersionHint provides the version webhook with guidance about the desired Kubernetes distribution version of the
	// Tanzu Kubernetes cluster. If a hint is provided without a full version, the most recent distribution matching the hint
	// will be selected.
	//
	// The version selected based on the hint will be stored in the spec as the full version. This ensures that the same
	// version is used if the cluster is scaled out in the future.
	//
	// VersionHint is a semantic prefix of a full version number. (E.g., v1.15.1 matches any distribution of v1.15.1,
	// including v1.15.1+vmware.1-tkg.1 or v1.15.1+vmware.2-tkg.1, but not v1.15.10+vmware.1-tkg.1.)
	//
	// A hint that does not match the full version is invalid and will be rejected.
	//
	// To upgrade a cluster to the most recent version that still matches the hint, leave the hint alone and remove the
	// fullVersion from the spec. This will cause the hint to be re-resolved.
	// +optional
	VersionHint string `json:"version"`
}

// Settings specifies configuration information for a cluster.
type Settings struct {
	// Network specifies network-related settings for the cluster.
	// +optional
	Network *Network `json:"network,omitempty"`

	// Storage specifies storage-related settings for the cluster.
	//
	// The storage used for node's disks is controlled by TopologySettings.
	// +optional
	Storage *Storage `json:"storage,omitempty"`
}

// Network specifies network-related settings for a cluster.
type Network struct {
	// Services specify network settings for services.
	//
	// Defaults to 10.96.0.0/12.
	// +optional
	Services *NetworkRanges `json:"services,omitempty"`

	// Pods specify network settings for pods.
	//
	// Defaults to 192.168.0.0/16.
	// +optional
	Pods *NetworkRanges `json:"pods,omitempty"`

	// ServiceDomain specifies service domain for Tanzu Kubernetes cluster.
	//
	// Defaults to a cluster.local.
	// +optional
	ServiceDomain string `json:"serviceDomain,omitempty"`

	// CNI is the Container Networking Interface plugin for the Tanzu Kubernetes cluster.
	//
	// Defaults to Calico.
	// +optional
	CNI *CNIConfiguration `json:"cni,omitempty"`

	// Proxy specifies HTTP(s) proxy configuration for Tanzu Kubernetes cluster.
	//
	// If omitted, no proxy will be configured in the system.
	// +optional
	Proxy *ProxyConfiguration `json:"proxy,omitempty"`

	// Trust specifies certificate configuration for the Tanzu Kubernetes Cluster.
	//
	// If omitted, no certificate will be configured in the system.
	// +optional
	Trust *TrustConfiguration `json:"trust,omitempty"`
}

// NetworkRanges describes a collection of IP addresses as a list of ranges.
type NetworkRanges struct {
	// CIDRBlocks specifies one or more ranges of IP addresses.
	//
	// Note: supplying multiple ranges many not be supported by all CNI plugins.
	// +optional
	CIDRBlocks []string `json:"cidrBlocks,omitempty"`
}

// CNIConfiguration indicates which CNI should be used.
type CNIConfiguration struct {
	// Name is the name of the CNI plugin to use.
	//
	// Supported values: "calico", "antrea".
	Name string `json:"name"`
}

// ProxyConfiguration configures the HTTP(s) proxy to be used inside the Tanzu Kubernetes cluster.
type ProxyConfiguration struct {
	// HttpProxy specifies a proxy URL to use for creating HTTP connections outside the cluster.
	// Example: http://<user>:<pwd>@<ip>:<port>
	//
	// +optional
	HttpProxy *string `json:"httpProxy,omitempty"`

	// HttpsProxy specifies a proxy URL to use for creating HTTPS connections outside the cluster.
	// Example: http://<user>:<pwd>@<ip>:<port>
	//
	// +optional
	HttpsProxy *string `json:"httpsProxy,omitempty"`

	// NoProxy specifies a list of destination domain names, domains, IP addresses or other network CIDRs to exclude proxying.
	// Example: [localhost, 127.0.0.1, 10.10.10.0/24]
	//
	// +optional
	NoProxy []string `json:"noProxy,omitempty"`
}

// TrustConfiguration configures additional trust parameters to the cluster configuration
type TrustConfiguration struct {
	// AdditionalTrustedCAs specifies the additional trusted certificates (which
	// can be additional CAs or end certificates) to add to the cluster
	//
	// +optional
	AdditionalTrustedCAs []TLSCertificate `json:"additionalTrustedCAs,omitempty"`
}

// TLSCertificate specifies a single additional certificate name and contents
type TLSCertificate struct {
	// Name specifies the name of the additional certificate, used in the filename
	// Example: CompanyInternalCA
	Name string `json:"name"`

	// Data specifies the contents of the additional certificate, encoded as a
	// base64 string. Specifically, this is the PEM Public Certificate data as
	// a base64 string..
	// Example: LS0tLS1C...LS0tCg== (where "..." is the middle section of the long base64 string)
	Data string `json:"data"`
}

// Storage configures persistent storage for a cluster.
type Storage struct {
	// Classes is a list of storage classes from the supervisor namespace to expose within a cluster.
	//
	// If omitted, all storage classes from the supervisor namespace will be exposed within the cluster.
	// +optional
	Classes []string `json:"classes,omitempty"`
	// DefaultClass is the valid storage class name which is treated as the default storage class within a cluster.
	// If omitted, no default storage class is set
	// +optional
	DefaultClass string `json:"defaultClass,omitempty"`
}

// TanzuKubernetesClusterStatus defines the observed state of TanzuKubernetesCluster.
type TanzuKubernetesClusterStatus struct {
	// ClusterAPIStatus describes the abridged status of the underlying Cluster
	// API Cluster status.
	// +optional
	ClusterAPIStatus *ClusterAPIStatus `json:"clusterApiStatus,omitempty"`

	// VMStatus is the VM status from the vSphere Virtual Machine service machines
	VMStatus map[string]VirtualMachineState `json:"vmStatus,omitempty"`

	// NodeStatus is the NodeReadyCondition result from the K8S control plane perspective
	// +optional
	NodeStatus map[string]NodeStatus `json:"nodeStatus,omitempty"`

	// DEPRECATED: UpgradeStatus is not updated or honored anywhere.
	// +optional
	Upgrade *UpgradeStatus `json:"upgrade,omitempty"`

	// Version holds the observed version of the Tanzu Kubernetes cluster. While an upgrade is in progress this value will be the
	// version of the cluster when the upgrade began.
	// +optional
	Version string `json:"version,omitempty"`

	// Addons groups the statuses of a Tanzu Kubernetes cluster's add-ons.
	// +optional
	Addons *AddonsStatuses `json:"addons,omitempty"`

	// Phase of this TanzuKubernetesCluster.
	// +optional
	Phase TanzuKubernetesClusterPhase `json:"phase,omitempty"`

	// The number of nodes automatically remediated
	NodeRemediationCount int `json:"nodeRemediationCount,omitempty"`

	// Conditions defines current service state of the TanzuKubernetestCluster.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the list of conditions for a TanzuKubernetesCluster
// object.
func (tkc *TanzuKubernetesCluster) GetConditions() Conditions {
	return tkc.Status.Conditions
}

// SetConditions sets the status conditions for a TanzuKubernetesCluster object
func (tkc *TanzuKubernetesCluster) SetConditions(conditions Conditions) {
	tkc.Status.Conditions = conditions
}

// AddonsStatuses groups the statuses of a Tanzu Kubernetes cluster's add-ons.
// Currently we track application status.
type AddonsStatuses struct {
	// DNS holds the DNS creation status for the Tanzu Kubernetes cluster.
	// +optional
	DNS *AddonStatus `json:"dns,omitempty"`

	// Proxy holds the Proxy creation status for the Tanzu Kubernetes cluster.
	// +optional
	Proxy *AddonStatus `json:"proxy,omitempty"`

	// PSP holds the default Pod Security Policy creation status for the Tanzu Kubernetes cluster.
	// +optional
	PSP *AddonStatus `json:"psp,omitempty"`

	// CNI holds the Container Networking Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CNI *AddonStatus `json:"cni,omitempty"`

	// CSI holds the Container Storage Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CSI *AddonStatus `json:"csi,omitempty"`

	// Cloudprovider holds the Cloud Provider Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CloudProvider *AddonStatus `json:"cloudprovider,omitempty"`

	// AuthService holds the Auth service status for the Tanzu Kubernetes cluster.
	// +optional
	AuthService *AddonStatus `json:"authsvc,omitempty"`

	// MetricsServer holds the Metrics Server status for the Tanzu Kubernetes cluster.
	// +optional
	MetricsServer *AddonStatus `json:"metrics-server,omitempty"`
}

// AreAppliedOrUnmanaged iterates over all of the fields of the AddonsStatuses
// struct and ensures that they are not nil and are either applied or unmanaged
// NOTE: At the moment, AddonsStatuses must _only_ contain objects of the type
// AddonStatus, else this call will panic.
func (addonsStatuses *AddonsStatuses) AreAppliedOrUnmanaged() bool { //nolint:stylecheck
	v := reflect.ValueOf(*addonsStatuses)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		status, ok := f.Interface().(*AddonStatus)
		if !ok {
			panic(fmt.Sprintf("Unable to convert %q to AddonStatus", f.Type()))
		}
		if status == nil || !status.IsAppliedOrUnmanaged() {
			return false
		}
	}

	return true
}

// ClusterAPIStatus defines the observed state of the underlying CAPI cluster.
type ClusterAPIStatus struct {
	// APIEndpoints represents the endpoints to communicate with the control plane.
	// +optional
	APIEndpoints []APIEndpoint `json:"apiEndpoints,omitempty"`

	// ErrorReason indicates that there is a problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	ErrorReason *ClusterAPIStatusError `json:"errorReason,omitempty"`

	// ErrorMessage indicates that there is a problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`
}

type ClusterAPIStatusError string

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int `json:"port"`
}

// AddonStatus represents the status of an addon. Used for PSP, CNI, CSI, CloudProvider, DNS and Proxy. See AddonsStatuses.
type AddonStatus struct {
	// Name of the add-on used.
	Name string `json:"name"`

	// LastErrorMessage contains any error that may have happened before and up to the current status. If it is set, it
	// means that status equaled error at some point in the past.
	// +optional
	LastErrorMessage string `json:"lastErrorMessage,omitempty"`

	// Status is the current state: pending, applied or error.
	Status TanzuKubernetesClusterAddonStatus `json:"status"`

	// Version of the distribution applied
	Version string `json:"version,omitempty"`

	// Conditions defines the current conditions of the add-on.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the list of conditions for an add-on
func (as *AddonStatus) GetConditions() Conditions {
	return as.Conditions
}

// SetConditions sets the conditions for an add-on
func (as *AddonStatus) SetConditions(conditions Conditions) {
	as.Conditions = conditions
}

// SetUnmanaged assigns the status to unmanaged
func (as *AddonStatus) SetUnmanaged() {
	as.Status = TanzuKubernetesClusterAddonsStatusUnmanaged
}

// SetError assigns both the Status to error and the LastErrorMessage according to err.
func (as *AddonStatus) SetError(err error) {
	if err == nil {
		return
	}

	as.Status = TanzuKubernetesClusterAddonsStatusError
	as.LastErrorMessage = err.Error()
}

// IsAppliedOrUnmanaaged returns true if the status is applied or unmanaged
func (as *AddonStatus) IsAppliedOrUnmanaged() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusApplied || as.Status == TanzuKubernetesClusterAddonsStatusUnmanaged
}

// IsUnmanaged returns true if the status is unmanaged
func (as *AddonStatus) IsUnmanaged() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusUnmanaged
}

// IsApplied returns true if the Status is applied
func (as *AddonStatus) IsApplied() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusApplied
}

// SetApplied sets the status to applied, records the addon name and version,
// and clears the last error message
func (as *AddonStatus) SetApplied(addonName string, version string) {
	as.Name = addonName
	as.Status = TanzuKubernetesClusterAddonsStatusApplied
	as.LastErrorMessage = ""
	as.Version = version
}

// NewAddonStatus returns an AddonStatus object set to pending.
func NewAddonStatus() *AddonStatus {
	return &AddonStatus{Status: TanzuKubernetesClusterAddonsStatusPending}
}

// UpgradeStatus represents the status of a cluster upgrade.
type UpgradeStatus struct {
	// ToVersion is the target version of the upgrade.
	ToVersion string `json:"toVersion"`

	// Phase is initially empty. Then it has the value of the latest/current phase of the upgrade.
	// In case of successful upgrade, the value is "Success".
	Phase Phase `json:"phase"`

	// UpgradeID is an opaque identifier unique to the latest/current upgrade phase.
	// +optional
	UpgradeID string `json:"upgradeId,omitempty"`

	// Finished is false initially and until the upgrade has finished, successfully or not.
	Finished bool `json:"done"`

	// Errors is the list of errors happened during an upgrade. If the upgrade has finished and was not successful, this
	// list should not be empty.
	// +optional
	Errors []Error `json:"errors,omitempty"`
}

// Error holds information about an error condition.
type Error struct {
	// Message is the textual description of the error.
	Message string `json:"message"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TanzuKubernetesCluster is the schema for the Tanzu Kubernetes Grid service for vSphere API.
type TanzuKubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TanzuKubernetesClusterSpec   `json:"spec,omitempty"`
	Status TanzuKubernetesClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TanzuKubernetesClusterList contains a list of TanzuKubernetesCluster
type TanzuKubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesCluster `json:"items"`
}

// VirtualMachineState describes the state of a VM.
type VirtualMachineState string

const (
	// VirtualMachineStateNotFound is the string representing a VM that cannot be located.
	VirtualMachineStateNotFound = VirtualMachineState("notfound")

	// VirtualMachineStatePending is the string representing a VM with an in-flight task.
	VirtualMachineStatePending = VirtualMachineState("pending")

	// VirtualMachineStateCreated is the string representing a VM that's been created
	VirtualMachineStateCreated = VirtualMachineState("created")

	// VirtualMachineStatePoweredOn is the string representing a VM that has successfully powered on
	VirtualMachineStatePoweredOn = VirtualMachineState("poweredon")

	// VirtualMachineStateReady is the string representing a powered-on VM with reported IP addresses.
	VirtualMachineStateReady = VirtualMachineState("ready")

	// VirtualMachineStateDeleting is the string representing a machine that still exists, but has a deleteTimestamp
	// Note that once a VirtualMachine is finally deleted, its state will be VirtualMachineStateNotFound
	VirtualMachineStateDeleting = VirtualMachineState("deleting")

	// VirtualMachineStateError is reported if an error occurs determining the status
	VirtualMachineStateError = VirtualMachineState("error")
)
