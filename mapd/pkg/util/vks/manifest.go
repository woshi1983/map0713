package vks

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const OwnerLabel = "map.huaruicloud.cn/cluster-id"

// BuildCluster produces one Supervisor CAPI Cluster. All topology inputs are
// validated before a remote write; unknown ClusterClass variables are not sent.
func BuildCluster(id uuid.UUID, name, releaseName string, spec model.ClusterVKSSpec) (*unstructured.Unstructured, error) {
	if id == uuid.Nil || name == "" || releaseName == "" || spec.Namespace == "" || spec.ClusterClass == "" {
		return nil, fmt.Errorf("VKS cluster ID, name, release, namespace and ClusterClass are required")
	}
	if spec.StorageClass == "" || spec.CtlPlaneVmClass == "" || spec.WorkerVmClass == "" || spec.OSName == "" {
		return nil, fmt.Errorf("VKS storage class, VM classes and OS name are required")
	}
	if spec.CtlPlaneReplicas < 1 || spec.WorkerReplicas < 1 {
		return nil, fmt.Errorf("VKS control plane and worker replicas must be positive")
	}
	if spec.CNI != "" && spec.CNI != "antrea" {
		return nil, fmt.Errorf("VKS does not support the selected CNI %q", spec.CNI)
	}
	if spec.ClusterCIDR == "" || spec.ServiceCIDR == "" || spec.ServiceDomain == "" {
		return nil, fmt.Errorf("VKS pod CIDR, service CIDR and service domain are required")
	}
	version := strings.ReplaceAll(releaseName, "---", "+")
	osSelector := "os-name=" + spec.OSName
	if spec.OSVersion != "" {
		osSelector += ",os-version=" + spec.OSVersion
	}
	workerVariables := []interface{}{map[string]interface{}{"name": "vmClass", "value": spec.WorkerVmClass}}
	if spec.WorkerContainerdVolumeSizeGi > 0 {
		workerVariables = append(workerVariables, map[string]interface{}{"name": "volumes", "value": []interface{}{map[string]interface{}{
			"name": "containerd-volume-pool-0", "mountPath": "/var/lib/containerd", "capacity": fmt.Sprintf("%dGi", spec.WorkerContainerdVolumeSizeGi), "storageClass": spec.StorageClass,
		}}})
	}
	controlPlane := map[string]interface{}{
		"replicas": spec.CtlPlaneReplicas,
		"metadata": map[string]interface{}{"annotations": map[string]interface{}{"run.tanzu.vmware.com/resolve-os-image": osSelector}},
	}
	if spec.CtlPlaneEtcdVolumeSizeGi > 0 {
		controlPlane["variables"] = map[string]interface{}{"overrides": []interface{}{map[string]interface{}{
			"name": "volumes", "value": []interface{}{map[string]interface{}{
				"name": "etcd-volume", "mountPath": "/var/lib/etcd", "capacity": fmt.Sprintf("%dGi", spec.CtlPlaneEtcdVolumeSizeGi), "storageClass": spec.StorageClass,
			}},
		}}}
	}
	topology := map[string]interface{}{
		"class":   spec.ClusterClass,
		"version": version,
		"variables": []interface{}{
			map[string]interface{}{"name": "vmClass", "value": spec.CtlPlaneVmClass},
			map[string]interface{}{"name": "storageClass", "value": spec.StorageClass},
		},
		"controlPlane": controlPlane,
		"workers": map[string]interface{}{"machineDeployments": []interface{}{map[string]interface{}{
			"class": "node-pool", "name": "pool-0", "replicas": spec.WorkerReplicas,
			"metadata":  map[string]interface{}{"annotations": map[string]interface{}{"run.tanzu.vmware.com/resolve-os-image": osSelector}},
			"variables": map[string]interface{}{"overrides": workerVariables},
		}}},
	}
	if len(spec.TrustedCAs) > 0 {
		certificates := make([]interface{}, 0, len(spec.TrustedCAs))
		for _, ca := range spec.TrustedCAs {
			if len(ca) == 0 {
				return nil, fmt.Errorf("VKS trusted CA certificate cannot be empty")
			}
			certificates = append(certificates, map[string]interface{}{"caCert": map[string]interface{}{"content": string(ca)}})
		}
		variables := topology["variables"].([]interface{})
		topology["variables"] = append(variables, map[string]interface{}{
			"name":  "osConfiguration",
			"value": map[string]interface{}{"trust": map[string]interface{}{"additionalTrustedCAs": certificates}},
		})
	}
	if spec.ClassNamespace != "" {
		topology["classNamespace"] = spec.ClassNamespace
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "cluster.x-k8s.io/v1beta1", "kind": "Cluster",
		"metadata": map[string]interface{}{"name": name, "namespace": spec.Namespace, "labels": map[string]interface{}{OwnerLabel: id.String(), "tkg-cluster-selector": name}},
		"spec": map[string]interface{}{
			"clusterNetwork": map[string]interface{}{
				"pods":          map[string]interface{}{"cidrBlocks": []interface{}{spec.ClusterCIDR}},
				"services":      map[string]interface{}{"cidrBlocks": []interface{}{spec.ServiceCIDR}},
				"serviceDomain": spec.ServiceDomain,
			},
			"topology": topology,
		},
	}}, nil
}
