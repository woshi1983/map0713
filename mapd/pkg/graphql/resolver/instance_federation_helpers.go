// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"
)

func convertInstanceFederationFromSQLModel(federation *model1.InstanceFederation) *model.InstanceFederation {
	obj := &model.InstanceFederation{
		ID:           federation.ID,
		Name:         federation.Name,
		Description:  federation.Description,
		Status:       model.InstanceFederationStatus(federation.Status.String()),
		CreatedAt:    federation.CreatedAt,
		LastModified: federation.UpdatedAt,
	}
	if federation.ComponentID.Valid {
		obj.Component = &model.Component{ID: federation.ComponentID.UUID}
	}
	if federation.ApplicationID.Valid {
		obj.Application = &model.Application{ID: federation.ApplicationID.UUID}
	}
	if federation.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: federation.OrganizationID.UUID}
	}
	return obj
}

func finalizeInstanceFederationQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`name` LIKE ?")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "component":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`component_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "application":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`application_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "organization":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`organization_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	return db
}

func postRenderManifest(k *k8s.Kubernetes, manifest []byte, namespace string, labels map[string]string) ([]byte, []*model1.InstanceStatusTask, error) {
	c, err := discovery.NewDiscoveryClientForConfig(k.Config)
	if err != nil {
		return nil, nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c))
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	buf := bytes.NewBuffer(nil)
	tasks := []*model1.InstanceStatusTask{}
	if err := k8s.WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, gvk, err := ser.Decode(p, nil, obj)
		if err != nil {
			if runtime.IsMissingKind(err) || runtime.IsMissingVersion(err) {
				return nil
			}
			return err
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		if mapping.Scope.Name() == meta.RESTScopeNameNamespace && obj.GetNamespace() == "" {
			// namespaced resources should specify the namespace
			obj.SetNamespace(manifestTemplateNSPlaceholder)
		}

		finalLabels := make(map[string]string)
		for k, v := range labels {
			finalLabels[k] = v
		}
		oldLabels := obj.GetLabels()
		for k, v := range oldLabels {
			if _, ok := finalLabels[k]; ok {
				continue
			}
			finalLabels[k] = v
		}
		obj.SetLabels(finalLabels)

		switch obj.GetKind() {
		case "DaemonSet":
			for k, v := range labels {
				unstructured.SetNestedField(obj.Object, v, "spec", "template", "metadata", "labels", k)
			}
		case "Deployment", "Job", "StatefulSet":
			for k, v := range labels {
				unstructured.SetNestedField(obj.Object, v, "spec", "template", "metadata", "labels", k)
			}
			unstructured.SetNestedField(obj.Object, "true", "spec", "template", "metadata", "labels", defLabelRestart)
		case "Namespace":
			obj.SetName(manifestTemplateNSPlaceholder)
		}

		if err := ser.Encode(obj, buf); err != nil {
			return err
		}

		tasks = append(tasks, &model1.InstanceStatusTask{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Kind:      obj.GetKind(),
		})
		return nil
	}); err != nil {
		return nil, nil, err
	}

	t, err := template.New("main").Parse(buf.String())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load template")
	}

	buf.Reset()
	if err := t.Execute(buf, &manifestTemplateRenderValue{Namespace: namespace}); err != nil {
		return nil, nil, errors.Wrap(err, "failed to render manifest")
	}
	return buf.Bytes(), tasks, nil
}

func compareManifestForBackendRef(p1, p2 []byte) (s1, s2 []*unstructured.Unstructured, err error) {
	m1 := make(map[string]*unstructured.Unstructured)
	i1 := make(map[string]int)
	m2 := make(map[string]*unstructured.Unstructured)
	i2 := make(map[string]int)

	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	if err = k8s.WalkManifest(p1, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}
		switch obj.GetKind() {
		case "Service", "Ingress":
		default:
			return nil
		}
		index := fmt.Sprintf("%s-%s", obj.GetKind(), obj.GetName())
		m1[index] = obj
		i1[index] = len(i1) + 1
		return nil
	}); err != nil {
		return
	}
	if err = k8s.WalkManifest(p2, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}
		switch obj.GetKind() {
		case "Service", "Ingress":
		default:
			return nil
		}
		index := fmt.Sprintf("%s-%s", obj.GetKind(), obj.GetName())
		m2[index] = obj
		i2[index] = len(i2) + 1
		return nil
	}); err != nil {
		return
	}

	for k, o1 := range m1 {
		if _, ok := m2[k]; !ok {
			s1 = append(s1, o1)
		}
	}
	for k, o2 := range m2 {
		o1, ok := m1[k]
		if !ok {
			s2 = append(s2, o2)
		} else {
			v1, v2 := "", ""
			f1, f2 := o1.DeepCopy(), o2.DeepCopy()
			removeUnstructuredField(f1, "metadata", "namespace")
			removeUnstructuredField(f2, "metadata", "namespace")
			removeUnstructuredField(f1, "metadata", "labels")
			removeUnstructuredField(f2, "metadata", "labels")
			removeUnstructuredField(f1, "metadata", "annotations")
			removeUnstructuredField(f2, "metadata", "annotations")
			removeUnstructuredField(f1, "status")
			removeUnstructuredField(f2, "status")
			v1, err = stringifyUnstructured(ser, f1)
			if err != nil {
				return
			}
			v2, err = stringifyUnstructured(ser, f2)
			if err != nil {
				return
			}
			if v1 != v2 {
				s1 = append(s1, o1)
				s2 = append(s2, o2)
			}
		}
	}
	sort.SliceStable(s1, func(i, j int) bool {
		o1, o2 := s1[i], s1[j]
		return i1[fmt.Sprintf("%s-%s", o1.GetKind(), o1.GetName())] < i1[fmt.Sprintf("%s-%s", o2.GetKind(), o2.GetName())]
	})
	sort.SliceStable(s2, func(i, j int) bool {
		o1, o2 := s2[i], s2[j]
		return i2[fmt.Sprintf("%s-%s", o1.GetKind(), o1.GetName())] < i2[fmt.Sprintf("%s-%s", o2.GetKind(), o2.GetName())]
	})
	return
}

func stringifyUnstructured(ser runtime.Serializer, o *unstructured.Unstructured) (string, error) {
	var buf bytes.Buffer
	if err := ser.Encode(o, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func removeUnstructuredField(obj *unstructured.Unstructured, fields ...string) {
	unstructured.RemoveNestedField(obj.Object, fields...)
}
