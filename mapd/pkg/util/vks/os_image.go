package vks

import (
	"context"
	"fmt"
	"sort"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type OSImageInfo struct {
	Name              string
	OSName            string
	OSVersion         string
	KubernetesVersion string
	Releases          []string
}

// ListOSImages returns OS images that users can select for a VKS cluster.
func (c *Client) ListOSImages(ctx context.Context, namespace string) ([]OSImageInfo, error) {
	images, available, err := c.listOSImages(ctx)
	if err != nil {
		return nil, err
	}
	if !available || len(images) == 0 {
		return c.listVMImages(ctx, namespace)
	}
	references, err := c.releaseImageReferences(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]OSImageInfo, 0, len(images))
	for _, image := range images {
		info, ok := selectableOSImage(&image)
		if ok {
			info.Releases = references[info.Name]
			result = append(result, info)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].KubernetesVersion != result[j].KubernetesVersion {
			return result[i].KubernetesVersion < result[j].KubernetesVersion
		}
		if result[i].OSName != result[j].OSName {
			return result[i].OSName < result[j].OSName
		}
		return result[i].OSVersion < result[j].OSVersion
	})
	return result, nil
}

func (c *Client) releaseImageReferences(ctx context.Context) (map[string][]string, error) {
	list, err := c.listReleases(ctx)
	if err != nil {
		return nil, err
	}
	references := make(map[string][]string)
	for _, release := range list.Items {
		if !releaseReady(&release) {
			continue
		}
		version, _, _ := unstructured.NestedString(release.Object, "spec", "version")
		if version == "" {
			version = strings.ReplaceAll(release.GetName(), "---", "+")
		}
		for _, name := range releaseOSImageNames(&release) {
			references[name] = append(references[name], version)
		}
	}
	return references, nil
}

func releaseOSImageNames(release *unstructured.Unstructured) []string {
	entries, _, _ := unstructured.NestedSlice(release.Object, "spec", "osImages")
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if value, ok := entry.(map[string]interface{}); ok {
			if name, ok := value["name"].(string); ok && name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

func (c *Client) listVMImages(ctx context.Context, namespace string) ([]OSImageInfo, error) {
	result := make([]OSImageInfo, 0)
	seen := make(map[string]bool)
	for _, resource := range []struct{ name, namespace string }{
		{name: "virtualmachineimages", namespace: namespace},
		{name: "clustervirtualmachineimages"},
	} {
		for _, version := range []string{"v1alpha2", "v1alpha1"} {
			gvr := schema.GroupVersionResource{Group: "vmoperator.vmware.com", Version: version, Resource: resource.name}
			var list *unstructured.UnstructuredList
			var err error
			if resource.namespace != "" {
				list, err = c.dynamic.Resource(gvr).Namespace(resource.namespace).List(ctx, metav1.ListOptions{})
			} else {
				list, err = c.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
			}
			if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			for _, item := range list.Items {
				info, ok := selectableVMImage(&item)
				if ok && !seen[info.Name] {
					seen[info.Name] = true
					result = append(result, info)
				}
			}
			break
		}
	}
	return result, nil
}

func selectableVMImage(image *unstructured.Unstructured) (OSImageInfo, bool) {
	if conditionFalse(image, "Ready") || conditionFalse(image, "Compatible") {
		return OSImageInfo{}, false
	}
	labels := image.GetLabels()
	info := OSImageInfo{Name: image.GetName()}
	info.KubernetesVersion = labels["vmi.vmoperator.vmware.com/kubernetes-version"]
	if info.KubernetesVersion == "" {
		info.KubernetesVersion = labels["image.capi.io/kubernetes-version"]
	}
	if info.KubernetesVersion == "" {
		info.KubernetesVersion = firstNestedString(image,
			[]string{"spec", "productInfo", "fullVersion"}, []string{"status", "productInfo", "fullVersion"},
			[]string{"status", "productInfo", "version"}, []string{"spec", "productInfo", "version"})
	}
	info.OSName = firstNestedString(image,
		[]string{"spec", "osInfo", "type"}, []string{"status", "osInfo", "type"},
		[]string{"spec", "osInfo", "id"}, []string{"status", "osInfo", "id"},
		[]string{"status", "image", "osInfo", "type"})
	if info.OSName == "" {
		for _, key := range []string{"vmi.vmoperator.vmware.com/os-type", "vmi.vmoperator.vmware.com/os-name", "image.capi.io/os-name"} {
			if info.OSName = labels[key]; info.OSName != "" {
				break
			}
		}
	}
	info.OSName = normalizeOSName(info.OSName)
	for _, key := range []string{"vmi.vmoperator.vmware.com/os-version", "image.capi.io/os-version"} {
		if info.OSVersion = labels[key]; info.OSVersion != "" {
			break
		}
	}
	if info.OSVersion == "" {
		info.OSVersion = firstNestedString(image,
			[]string{"spec", "osInfo", "version"}, []string{"status", "osInfo", "version"},
			[]string{"status", "image", "osInfo", "version"})
	}
	return info, info.Name != "" && info.KubernetesVersion != "" && info.OSName != "" && info.OSVersion != ""
}

func selectableOSImage(image *unstructured.Unstructured) (OSImageInfo, bool) {
	if conditionFalse(image, "Ready") || conditionFalse(image, "Compatible") {
		return OSImageInfo{}, false
	}
	info := OSImageInfo{Name: image.GetName()}
	info.KubernetesVersion = firstNestedString(image, []string{"spec", "kubernetesVersion"}, []string{"spec", "k8sVersion"}, []string{"status", "kubernetesVersion"})
	info.OSName = firstNestedString(image, []string{"spec", "os", "name"}, []string{"spec", "os", "type"}, []string{"spec", "osInfo", "name"}, []string{"spec", "osInfo", "type"})
	if info.OSName == "" {
		info.OSName = image.GetLabels()["os-name"]
	}
	info.OSName = normalizeOSName(info.OSName)
	info.OSVersion = firstNestedString(image, []string{"spec", "os", "version"}, []string{"spec", "osInfo", "version"})
	if info.OSVersion == "" {
		info.OSVersion = image.GetLabels()["os-version"]
	}
	return info, info.Name != "" && info.KubernetesVersion != "" && info.OSName != "" && info.OSVersion != ""
}

// ResolveOSVersion selects an OSImage matching the exact release selected for
// the cluster. Older Supervisors without the OSImage API rely on admission
// validation instead.
func (c *Client) ResolveOSVersion(ctx context.Context, releaseName, osName, requestedVersion string) (string, error) {
	images, available, err := c.listOSImages(ctx)
	if err != nil {
		return "", err
	}
	if !available {
		return requestedVersion, nil
	}
	releases, err := c.listReleases(ctx)
	if err == nil {
		for _, release := range releases.Items {
			if release.GetName() != releaseName {
				continue
			}
			if names := releaseOSImageNames(&release); len(names) > 0 {
				allowed := make(map[string]bool, len(names))
				for _, name := range names {
					allowed[name] = true
				}
				filtered := make([]unstructured.Unstructured, 0, len(images))
				for _, image := range images {
					if allowed[image.GetName()] {
						filtered = append(filtered, image)
					}
				}
				images = filtered
			}
			break
		}
	}
	return chooseOSVersion(images, releaseName, osName, requestedVersion)
}

func (c *Client) listOSImages(ctx context.Context) ([]unstructured.Unstructured, bool, error) {
	available := false
	for _, version := range []string{"v1alpha3", "v1alpha2", "v1alpha1"} {
		gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: version, Resource: "osimages"}
		list, err := c.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
		if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
			continue
		}
		if err != nil {
			return nil, false, err
		}
		available = true
		if len(list.Items) > 0 {
			return list.Items, true, nil
		}
	}
	return nil, available, nil
}

func chooseOSVersion(images []unstructured.Unstructured, releaseName, osName, requestedVersion string) (string, error) {
	wantedRelease := normalizeReleaseName(releaseName)
	wantedOS := normalizeOSName(osName)
	seen := make(map[string]bool)
	var versions []string
	for _, image := range images {
		info, ok := selectableOSImage(&image)
		if !ok {
			continue
		}
		if normalizeReleaseName(info.KubernetesVersion) != wantedRelease {
			continue
		}
		if info.OSName != wantedOS {
			continue
		}
		version := info.OSVersion
		if version != "" && !seen[version] {
			seen[version] = true
			versions = append(versions, version)
		}
	}
	if requestedVersion != "" {
		if seen[requestedVersion] {
			return requestedVersion, nil
		}
		return "", fmt.Errorf("VKS OSImage %s %s is unavailable for release %s", osName, requestedVersion, releaseName)
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no ready VKS OSImage for %s and release %s", osName, releaseName)
	}
	sort.Strings(versions)
	return versions[len(versions)-1], nil
}

func normalizeReleaseName(value string) string {
	value = strings.ReplaceAll(value, "---", "+")
	if index := strings.Index(value, "-vkr."); index >= 0 {
		value = value[:index]
	}
	return value
}

func normalizeOSName(value string) string {
	value = strings.ToLower(value)
	switch {
	case strings.HasPrefix(value, "vmwarephoton"), strings.HasPrefix(value, "photon"):
		return "photon"
	case strings.HasPrefix(value, "ubuntu"):
		return "ubuntu"
	case strings.HasPrefix(value, "windows"):
		return "windows"
	default:
		return value
	}
}

func firstNestedString(obj *unstructured.Unstructured, paths ...[]string) string {
	for _, path := range paths {
		if value, _, _ := unstructured.NestedString(obj.Object, path...); value != "" {
			return value
		}
	}
	return ""
}

func conditionFalse(obj *unstructured.Unstructured, name string) bool {
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for _, item := range conditions {
		condition, ok := item.(map[string]interface{})
		if ok && condition["type"] == name && condition["status"] != "True" {
			return true
		}
	}
	return false
}
