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
		if conditionFalse(&image, "Ready") || conditionFalse(&image, "Compatible") {
			continue
		}
		kubernetesVersion := firstNestedString(&image, []string{"spec", "kubernetesVersion"}, []string{"spec", "k8sVersion"}, []string{"status", "kubernetesVersion"})
		if normalizeReleaseName(kubernetesVersion) != wantedRelease {
			continue
		}
		name := firstNestedString(&image, []string{"spec", "os", "name"}, []string{"spec", "os", "type"}, []string{"spec", "osInfo", "name"}, []string{"spec", "osInfo", "type"})
		if name == "" {
			name = image.GetLabels()["os-name"]
		}
		if normalizeOSName(name) != wantedOS {
			continue
		}
		version := firstNestedString(&image, []string{"spec", "os", "version"}, []string{"spec", "osInfo", "version"})
		if version == "" {
			version = image.GetLabels()["os-version"]
		}
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
		if ok && condition["type"] == name && condition["status"] == "False" {
			return true
		}
	}
	return false
}
