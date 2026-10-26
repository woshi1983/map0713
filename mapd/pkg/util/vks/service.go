package vks

import (
	"context"
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var releaseGVR = schema.GroupVersionResource{Group: "kubernetes.vmware.com", Version: "v1alpha1", Resource: "kubernetesreleases"}
var legacyReleaseGVR = schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha3", Resource: "tanzukubernetesreleases"}

type Client struct {
	dynamic     dynamic.Interface
	kube        kubernetes.Interface
	capiVersion string
}

func (c *Client) capiResource(resource string) schema.GroupVersionResource {
	version := c.capiVersion
	if version == "" {
		version = "v1beta1"
	}
	return schema.GroupVersionResource{Group: "cluster.x-k8s.io", Version: version, Resource: resource}
}

// NewClientFromProvider authenticates to the Supervisor using the existing
// MAP provider credential format, then uses that session for VKS resources.
func NewClientFromProvider(id uuid.UUID, spec *model.ClusterProviderVKSSpec) (*Client, *tanzuv2.VSphereWithTanzu, error) {
	block, err := cryptoutil.NewAESGCMCipher(id[:])
	if err != nil {
		return nil, nil, err
	}
	supervisor, err := tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify, CAData: spec.CACert,
	})
	if err != nil {
		return nil, nil, err
	}
	client, err := NewClient(supervisor.Config)
	return client, supervisor, err
}

func NewClient(config *rest.Config) (*Client, error) {
	clientConfig := rest.CopyConfig(config)
	if clientConfig.Timeout == 0 {
		clientConfig.Timeout = mapd.GlobalConfig.HTTPClientTimeout
	}
	d, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	k, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	client := &Client{dynamic: d, kube: k}
	for _, version := range []string{"v1beta1", "v1beta2"} {
		resources, err := k.Discovery().ServerResourcesForGroupVersion("cluster.x-k8s.io/" + version)
		if err != nil {
			continue
		}
		clusters, classes := false, false
		for _, resource := range resources.APIResources {
			if resource.Name == "clusters" {
				clusters = true
			}
			if resource.Name == "clusterclasses" {
				classes = true
			}
		}
		if clusters && classes {
			client.capiVersion = version
			break
		}
	}
	return client, nil
}

func (c *Client) GetCluster(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return c.dynamic.Resource(c.capiResource("clusters")).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ValidateCluster(ctx context.Context, id uuid.UUID, name, releaseName string, spec model.ClusterVKSSpec) error {
	obj, err := BuildCluster(id, name, releaseName, spec)
	if err != nil {
		return err
	}
	if err := configureClusterAPIVersion(obj, c.capiResource("clusters").Version); err != nil {
		return err
	}
	_, err = c.dynamic.Resource(c.capiResource("clusters")).Namespace(spec.Namespace).Create(ctx, obj, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})
	return err
}

// CreateCluster never updates an unowned or same-name remote object. A retry
// may continue only when the remote object already carries this MAP ID.
func (c *Client) CreateCluster(ctx context.Context, id uuid.UUID, name, releaseName string, spec model.ClusterVKSSpec) (*unstructured.Unstructured, error) {
	obj, err := BuildCluster(id, name, releaseName, spec)
	if err != nil {
		return nil, err
	}
	if err := configureClusterAPIVersion(obj, c.capiResource("clusters").Version); err != nil {
		return nil, err
	}
	existing, err := c.GetCluster(ctx, spec.Namespace, name)
	if err == nil {
		if existing.GetLabels()[OwnerLabel] != id.String() || (spec.RemoteUID != "" && string(existing.GetUID()) != spec.RemoteUID) {
			return nil, fmt.Errorf("VKS cluster %s/%s already exists and is not owned by MAP record %s", spec.Namespace, name, id)
		}
		return existing, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, err
	}
	return c.dynamic.Resource(c.capiResource("clusters")).Namespace(spec.Namespace).Create(ctx, obj, metav1.CreateOptions{})
}

func (c *Client) DeleteCluster(ctx context.Context, id uuid.UUID, name string, spec model.ClusterVKSSpec) error {
	obj, err := c.GetCluster(ctx, spec.Namespace, name)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if obj.GetLabels()[OwnerLabel] != id.String() || spec.RemoteUID == "" || string(obj.GetUID()) != spec.RemoteUID {
		return fmt.Errorf("VKS cluster %s/%s ownership or UID mismatch", spec.Namespace, name)
	}
	uid := obj.GetUID()
	return c.dynamic.Resource(c.capiResource("clusters")).Namespace(spec.Namespace).Delete(ctx, name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
}

func Available(obj *unstructured.Unstructured) bool {
	for _, path := range [][]string{{"status", "conditions"}, {"status", "v1beta2", "conditions"}} {
		conditions, _, _ := unstructured.NestedSlice(obj.Object, path...)
		for _, item := range conditions {
			condition, ok := item.(map[string]interface{})
			if ok && condition["type"] == "Available" && condition["status"] == "True" {
				return true
			}
		}
	}
	available, _, _ := unstructured.NestedBool(obj.Object, "status", "available")
	return available
}

func configureClusterAPIVersion(obj *unstructured.Unstructured, version string) error {
	obj.SetAPIVersion("cluster.x-k8s.io/" + version)
	if version != "v1beta2" {
		return nil
	}
	class, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "class")
	namespace, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "classNamespace")
	ref := map[string]interface{}{"name": class}
	if namespace != "" {
		ref["namespace"] = namespace
	}
	unstructured.RemoveNestedField(obj.Object, "spec", "topology", "class")
	unstructured.RemoveNestedField(obj.Object, "spec", "topology", "classNamespace")
	return unstructured.SetNestedMap(obj.Object, ref, "spec", "topology", "classRef")
}

func Endpoint(obj *unstructured.Unstructured) string {
	host, _, _ := unstructured.NestedString(obj.Object, "spec", "controlPlaneEndpoint", "host")
	port, _, _ := unstructured.NestedInt64(obj.Object, "spec", "controlPlaneEndpoint", "port")
	if host == "" || port == 0 {
		return ""
	}
	return fmt.Sprintf("https://%s:%d", host, port)
}

func (c *Client) GetKubeconfig(ctx context.Context, namespace, name string) ([]byte, error) {
	secret, err := c.kube.CoreV1().Secrets(namespace).Get(ctx, name+"-kubeconfig", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	value := secret.Data["value"]
	if len(value) == 0 {
		return nil, fmt.Errorf("VKS kubeconfig secret %s/%s-kubeconfig has no value", namespace, name)
	}
	return value, nil
}

func (c *Client) GetCA(ctx context.Context, namespace, name string) ([]byte, error) {
	secret, err := c.kube.CoreV1().Secrets(namespace).Get(ctx, name+"-ca", metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if err == nil && len(secret.Data["tls.crt"]) > 0 {
		return secret.Data["tls.crt"], nil
	}
	config, err := c.GetKubeconfig(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	parsed, err := clientcmd.Load(config)
	if err != nil {
		return nil, err
	}
	current := parsed.Contexts[parsed.CurrentContext]
	if current == nil || parsed.Clusters[current.Cluster] == nil {
		return nil, fmt.Errorf("VKS kubeconfig has no current cluster CA")
	}
	ca := parsed.Clusters[current.Cluster].CertificateAuthorityData
	if len(ca) == 0 {
		return nil, fmt.Errorf("VKS kubeconfig has no embedded CA")
	}
	return ca, nil
}

func (c *Client) ListClusterClasses(ctx context.Context, namespace string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string
	for _, ns := range []string{namespace, "vmware-system-vks-public"} {
		if seen[ns] {
			continue
		}
		seen[ns] = true
		list, err := c.dynamic.Resource(c.capiResource("clusterclasses")).Namespace(ns).List(ctx, metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, item := range list.Items {
			result = append(result, ns+"/"+item.GetName())
		}
	}
	return result, nil
}

// ResolveRelease maps the semver shown by MAP to the exact Supervisor resource
// name used by topology.version. It refuses ambiguous and non-ready releases.
func (c *Client) ResolveRelease(ctx context.Context, requested string) (string, error) {
	list, err := c.listReleases(ctx)
	if err != nil {
		return "", err
	}
	want, err := semver.NewVersion(requested)
	if err != nil {
		return "", err
	}
	var exact, matches []string
	for _, item := range list.Items {
		if !releaseReady(&item) {
			continue
		}
		version, _, _ := unstructured.NestedString(item.Object, "spec", "version")
		if version == "" {
			version = strings.ReplaceAll(item.GetName(), "---", "+")
		}
		parsed, err := semver.NewVersion(version)
		if err == nil {
			if parsed.String() == want.String() {
				exact = append(exact, item.GetName())
			}
			if parsed.Equal(want) {
				matches = append(matches, item.GetName())
			}
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return "", fmt.Errorf("VKS release %s has %d exact matches", requested, len(exact))
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("VKS release %s has %d compatible matches", requested, len(matches))
	}
	return matches[0], nil
}

func (c *Client) listReleases(ctx context.Context) (*unstructured.UnstructuredList, error) {
	list, err := c.dynamic.Resource(releaseGVR).List(ctx, metav1.ListOptions{})
	if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
		return c.dynamic.Resource(legacyReleaseGVR).List(ctx, metav1.ListOptions{})
	}
	return list, err
}

func (c *Client) ListReadyVersions(ctx context.Context) ([]*semver.Version, error) {
	list, err := c.listReleases(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var versions []*semver.Version
	for _, item := range list.Items {
		if !releaseReady(&item) {
			continue
		}
		version, _, _ := unstructured.NestedString(item.Object, "spec", "version")
		if version == "" {
			version = strings.ReplaceAll(item.GetName(), "---", "+")
		}
		parsed, err := semver.NewVersion(version)
		if err != nil {
			continue
		}
		if !seen[parsed.String()] {
			seen[parsed.String()] = true
			versions = append(versions, parsed)
		}
	}
	return versions, nil
}

func releaseReady(item *unstructured.Unstructured) bool {
	conditions, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions")
	ready, compatible := false, false
	for _, raw := range conditions {
		c, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if c["type"] == "Ready" && c["status"] == "True" {
			ready = true
		}
		if c["type"] == "Compatible" && c["status"] == "True" {
			compatible = true
		}
	}
	return ready && compatible
}
