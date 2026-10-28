package vks

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetDefaultStorageClassPreservesOtherAvailableClasses(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{"topology": map[string]interface{}{
			"variables": []interface{}{
				map[string]interface{}{"name": "vmClass", "value": "small"},
				map[string]interface{}{"name": "vsphereOptions", "value": map[string]interface{}{
					"persistentVolumes": map[string]interface{}{
						"availableStorageClasses": []interface{}{"slow", "archive"},
						"defaultStorageClass":     "slow",
					},
				}},
			},
		}},
	}}
	updated, changed, err := setDefaultStorageClass(obj, "fast")
	if err != nil || !changed {
		t.Fatalf("setDefaultStorageClass() changed=%v err=%v", changed, err)
	}
	got, _, _ := unstructured.NestedSlice(updated.Object, "spec", "topology", "variables")
	options := got[1].(map[string]interface{})["value"].(map[string]interface{})
	persistent := options["persistentVolumes"].(map[string]interface{})
	if persistent["defaultStorageClass"] != "fast" || !reflect.DeepEqual(persistent["availableStorageClasses"], []interface{}{"slow", "archive", "fast"}) {
		t.Fatalf("unexpected persistentVolumes: %v", persistent)
	}
	original, _, _ := unstructured.NestedSlice(obj.Object, "spec", "topology", "variables")
	if original[1].(map[string]interface{})["value"].(map[string]interface{})["persistentVolumes"].(map[string]interface{})["defaultStorageClass"] != "slow" {
		t.Fatal("input Cluster was mutated")
	}
	_, changed, err = setDefaultStorageClass(updated, "fast")
	if err != nil || changed {
		t.Fatalf("idempotent update changed=%v err=%v", changed, err)
	}
}

func TestSetDefaultStorageClassAddsOptionsWithoutRestrictingClasses(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{"topology": map[string]interface{}{
			"variables": []interface{}{map[string]interface{}{"name": "storageClass", "value": "fast"}},
		}},
	}}
	updated, changed, err := setDefaultStorageClass(obj, "fast")
	if err != nil || !changed {
		t.Fatalf("setDefaultStorageClass() changed=%v err=%v", changed, err)
	}
	variables, _, _ := unstructured.NestedSlice(updated.Object, "spec", "topology", "variables")
	options := variables[1].(map[string]interface{})["value"].(map[string]interface{})["persistentVolumes"].(map[string]interface{})
	if options["defaultStorageClass"] != "fast" {
		t.Fatalf("default class = %v", options["defaultStorageClass"])
	}
	if _, restricted := options["availableStorageClasses"]; restricted {
		t.Fatalf("available classes were restricted: %v", options)
	}
}
