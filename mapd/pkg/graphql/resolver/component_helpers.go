// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	yamlv2 "gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func convertComponentFromSQLModel(com *model1.Component) *model.Component {
	obj := &model.Component{
		ID:           com.ID,
		Name:         com.Name,
		Description:  com.Description,
		Type:         model.ComponentType(com.Type.String()),
		CreatedAt:    com.CreatedAt,
		LastModified: com.UpdatedAt,
	}
	if com.Avatar != "" {
		obj.Avatar = abstract.MustParseURL(com.Avatar)
	}
	switch com.Type {
	case model1.ComponentTypeBasic:
		template := &model.CreateBasicComponentTemplate{}
		if err := json.Unmarshal(com.Template, template); err != nil {
			panic(err)
		}
		t := &model.BasicComponentTemplate{}
		convertBasicComponentTemplate(t, template)
		obj.Template = t
	case model1.ComponentTypeHelm:
		v := &model1.ComponentHelmTemplate{}
		if err := json.Unmarshal(com.Template, v); err != nil {
			panic(err)
		}
		t := &model.HelmComponentTemplate{
			URL:       abstract.MustParseURL(v.URL),
			Values:    v.Values,
			ExtraInfo: v.ExtraInfo,
		}
		obj.Template = t
	case model1.ComponentTypeRaw:
		t := &model.RAWComponentTemplate{Raw: string(com.Template)}
		obj.Template = t
	}
	if com.ApplicationID.Valid {
		obj.Application = &model.Application{ID: com.ApplicationID.UUID}
	}
	if com.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: com.OrganizationID.UUID}
	}
	return obj
}

func convertChartDetailsFromQualifiedChart(chrt *chart.Chart) *model.ChartDetails {
	obj := &model.ChartDetails{
		Name:        chrt.Metadata.Name,
		Description: chrt.Metadata.Description,
		Version:     chrt.Metadata.Version,
		Keywords:    chrt.Metadata.Keywords,
		Files:       make(map[string]string),
	}
	if chrt.Metadata.Icon != "" {
		v, err := url.Parse(chrt.Metadata.Icon)
		if err == nil {
			obj.Avatar = v
		}
	}
	for _, each := range chrt.Metadata.Sources {
		v, err := url.Parse(each)
		if err == nil {
			obj.Sources = append(obj.Sources, v)
		}
	}
	if chrt.Metadata.Home != "" {
		v, err := url.Parse(chrt.Metadata.Home)
		if err == nil {
			obj.Homepage = v
		}
	}
	for _, each := range chrt.Metadata.Dependencies {
		dep := &model.ChartDetailsDependency{
			Name: each.Name,
		}
		if each.Repository != "" {
			v, err := url.Parse(each.Repository)
			if err == nil {
				dep.Repository = v
			}
		}
		if each.Version != "" {
			dep.Version = &each.Version
		}
		obj.Dependencies = append(obj.Dependencies, dep)
	}
	for _, each := range chrt.Files {
		obj.Files[each.Name] = string(each.Data)
	}
	b, err := json.Marshal(chrt.Values)
	if err != nil {
		panic(err)
	}
	obj.Values = b
	if _, ok := obj.Files["values.yaml"]; !ok {
		b, err := yamlv2.Marshal(chrt.Values)
		if err != nil {
			panic(err)
		}
		obj.Files["values.yaml"] = string(b)
	}
	return obj
}

func finalizeComponentQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "clusterBackupSchedule":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` IN (SELECT `component_id` FROM `cluster_backup_schedule_component` WHERE `schedule_id` = ?)")
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

func validateManifestAndConvertToTemplate(b []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := yamlv2.NewEncoder(buf)
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err := k8s.WalkManifest(b, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}

		if obj.GetKind() == "Namespace" {
			obj.SetName(manifestTemplateNSPlaceholder)
		}
		if obj.GetNamespace() != "" {
			obj.SetNamespace(manifestTemplateNSPlaceholder)
		}
		return enc.Encode(obj.Object)
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderChart(releaseName string, fpath string, values map[string]interface{}) (crd []byte, manifest []byte, err error) {
	// workaround to prevent panic when converting YAML to JSON due to "{{" is
	// a special character in YAML.
	//
	// this is a very special placeholder to reduce the oppotunity to match other contents.
	placeholder := "??NAMESPACE_PLACEHOLDER-" + randutil.StringN(8) + "{{.Namespace}}??"
	crd, manifest, err = k8s.RenderChart(placeholder, releaseName, fpath, values)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(nil)
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err = k8s.WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			if runtime.IsMissingKind(err) || runtime.IsMissingVersion(err) {
				return nil
			}
			return err
		}
		return ser.Encode(obj, buf)
	}); err != nil {
		return
	}

	manifest = bytes.ReplaceAll(buf.Bytes(), []byte(placeholder), []byte(manifestTemplateNSPlaceholder))
	buf.Reset()
	enc := yamlv2.NewEncoder(buf)

	if err = k8s.WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}
		return enc.Encode(obj.Object)
	}); err != nil {
		return
	}

	return crd, buf.Bytes(), nil
}

func convertBasicComponentTemplate(template *model.BasicComponentTemplate, obj *model.CreateBasicComponentTemplate) {
	template.Deployments = template.Deployments[:0]
	for _, u := range obj.Deployments {
		deployment := &model.BasicComponentTemplateDeployment{
			Name:   u.Name,
			Labels: u.Labels,
		}
		if u.Spec != nil {
			deployment.Spec = convertDeploymentSpecFromInput(u.Spec)
		}
		if u.Hpa != nil {
			deployment.Hpa = convertHorizontalPodAutoscalerSpecFromInput(u.Hpa)
		}
		template.Deployments = append(template.Deployments, deployment)
	}
	template.Services = template.Services[:0]
	for _, u := range obj.Services {
		svc := &model.BasicComponentTemplateService{
			Name:   u.Name,
			Labels: u.Labels,
		}
		if u.Spec != nil {
			svc.Spec = convertServiceSpecFromInput(u.Spec)
		}
		template.Services = append(template.Services, svc)
	}
	template.Secrets = template.Secrets[:0]
LOOP:
	for _, u := range obj.Secrets {
		var secret model.BasicComponentTemplateSecret
		switch {
		case u.Generic != nil:
			v := &model.BasicComponentTemplateGenericSecret{
				Name:   u.Name,
				Labels: u.Labels,
			}
			convertGenericSecretFromInput(v, u.Generic)
			secret = v
		case u.DockerConfigJSON != nil:
			v := &model.BasicComponentTemplateDockerConfigJSONSecret{
				Name:   u.Name,
				Labels: u.Labels,
			}
			convertDockerConfigJSONSecretFromInput(v, u.DockerConfigJSON)
			secret = v
		case u.BasicAuth != nil:
			v := &model.BasicComponentTemplateBasicAuthSecret{
				Name:   u.Name,
				Labels: u.Labels,
			}
			convertBasicAuthSecretFromInput(v, u.BasicAuth)
			secret = v
		case u.SSHAuth != nil:
			v := &model.BasicComponentTemplateSSHAuthSecret{
				Name:   u.Name,
				Labels: u.Labels,
			}
			convertSSHAuthSecretFromInput(v, u.SSHAuth)
			secret = v
		case u.TLS != nil:
			v := &model.BasicComponentTemplateTLSSecret{
				Name:   u.Name,
				Labels: u.Labels,
			}
			convertTLSSecretFromInput(v, u.TLS)
			secret = v
		default:
			continue LOOP
		}
		template.Secrets = append(template.Secrets, secret)
	}
	template.ConfigMaps = template.ConfigMaps[:0]
	for _, u := range obj.ConfigMaps {
		cm := &model.BasicComponentTemplateConfigMap{
			Name:   u.Name,
			Labels: u.Labels,
			Data:   u.Data,
		}
		template.ConfigMaps = append(template.ConfigMaps, cm)
	}
	template.PersistentVolumeClaims = template.PersistentVolumeClaims[:0]
	for _, u := range obj.PersistentVolumeClaims {
		pvc := &model.BasicComponentTemplatePersistentVolumeClaim{
			Name:   u.Name,
			Labels: u.Labels,
			Spec:   convertPersistentVolumeClaimSpecFromInput(u.Spec),
		}
		template.PersistentVolumeClaims = append(template.PersistentVolumeClaims, pvc)
	}
}

func convertGenericSecretFromInput(obj *model.BasicComponentTemplateGenericSecret, input *model.AddBasicComponentTemplateGenericSecret) {
	obj.Type = &input.Type
	obj.Data = input.Data
}

func convertDockerConfigJSONSecretFromInput(obj *model.BasicComponentTemplateDockerConfigJSONSecret, input *model.AddBasicComponentTemplateDockerConfigJSONSecret) {
	obj.Type = util.StringPtr(string(corev1.SecretTypeDockerConfigJson))
	obj.Endpoint = input.Endpoint
	obj.AuthUser = &input.AuthUser
	obj.Email = input.Email
}

func convertBasicAuthSecretFromInput(obj *model.BasicComponentTemplateBasicAuthSecret, input *model.AddBasicComponentTemplateBasicAuthSecret) {
	obj.Type = util.StringPtr(string(corev1.SecretTypeBasicAuth))
	obj.AuthUser = &input.AuthUser
}

func convertSSHAuthSecretFromInput(obj *model.BasicComponentTemplateSSHAuthSecret, input *model.AddBasicComponentTemplateSSHAuthSecret) {
	obj.Type = util.StringPtr(string(corev1.SecretTypeSSHAuth))
}

func convertTLSSecretFromInput(obj *model.BasicComponentTemplateTLSSecret, input *model.AddBasicComponentTemplateTLSSecret) {
	obj.Type = util.StringPtr(string(corev1.SecretTypeTLS))
	obj.Cert = &input.Cert
}

func convertPersistentVolumeClaimSpecFromInput(input *model.CreatePersistentVolumeClaimSpec) *model.PersistentVolumeClaimSpec {
	spec := &model.PersistentVolumeClaimSpec{
		AccessModes:  input.AccessModes,
		StorageClass: input.StorageClass,
		VolumeMode:   input.VolumeMode,
	}
	if input.Resources != nil {
		spec.Resources = &model.PersistentVolumeClaimSpecResources{}
		if input.Resources.Limits != nil {
			spec.Resources.Limits = &model.PersistentVolumeClaimSpecResourcesLimits{
				Storage: input.Resources.Limits.Storage,
			}
		}
		if input.Resources.Requests != nil {
			spec.Resources.Requests = &model.PersistentVolumeClaimSpecResourcesRequests{
				Storage: &input.Resources.Requests.Storage,
			}
		}
	}
	return spec
}

func renderBasicComponentManifest(template *model.CreateBasicComponentTemplate) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for _, each := range template.Secrets {
		if err := serializeSecret(buf, ser, each); err != nil {
			return nil, err
		}
	}
	for _, each := range template.ConfigMaps {
		if err := serializeConfigMap(buf, ser, each); err != nil {
			return nil, err
		}
	}
	for _, each := range template.PersistentVolumeClaims {
		if err := serializePersistentVolumeClaim(buf, ser, each); err != nil {
			return nil, err
		}
	}
	for _, each := range template.Services {
		if err := serializeService(buf, ser, each.Name, each.Labels, each.Spec); err != nil {
			return nil, err
		}
	}
	for _, each := range template.Deployments {
		if err := serializeDeployment(buf, ser, each.Name, each.Labels, each.Spec); err != nil {
			return nil, err
		}
		if each.Hpa != nil {
			if err := serializeHorizontalPodAutoscaler(buf, ser, each.Name, each.Labels, each.Hpa); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func serializeSecret(buf io.Writer, ser runtime.Serializer, params *model.AddBasicComponentTemplateSecret) error {
	switch {
	case params.Generic != nil:
		return serializeGenericSecret(buf, ser, params.Name, params.Labels, params.Generic)
	case params.DockerConfigJSON != nil:
		return serializeDockerConfigJSONSecret(buf, ser, params.Name, params.Labels, params.DockerConfigJSON)
	case params.BasicAuth != nil:
		return serializeBasicAuthSecret(buf, ser, params.Name, params.Labels, params.BasicAuth)
	case params.SSHAuth != nil:
		return serializeSSHAuthSecret(buf, ser, params.Name, params.Labels, params.SSHAuth)
	case params.TLS != nil:
		return serializeTLSSecret(buf, ser, params.Name, params.Labels, params.TLS)
	}
	return nil
}

func serializeGenericSecret(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.AddBasicComponentTemplateGenericSecret) error {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    labels,
		},
		Data: make(map[string][]byte),
		Type: corev1.SecretType(spec.Type),
	}
	for k, v := range spec.Data {
		secret.Data[k] = []byte(v)
	}
	return ser.Encode(secret, buf)
}

func serializeDockerConfigJSONSecret(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.AddBasicComponentTemplateDockerConfigJSONSecret) error {
	cred := &k8s.DockerConfigJSON{
		Auths: map[string]*k8s.DockerConfigJSONAuth{
			spec.Endpoint.String(): {
				Username: spec.AuthUser,
				Password: string(spec.AuthPassword),
				Auth:     []byte(strings.Join([]string{spec.AuthUser, string(spec.AuthPassword)}, ":")),
			},
		},
	}
	if spec.Email != nil {
		cred.Auths[spec.Endpoint.String()].Email = *spec.Email
	}
	b, err := json.Marshal(cred)
	if err != nil {
		panic(err)
	}

	obj := &model.AddBasicComponentTemplateGenericSecret{
		Data: map[string]string{".dockerconfigjson": string(b)},
		Type: "kubernetes.io/dockerconfigjson",
	}
	return serializeGenericSecret(buf, ser, name, labels, obj)
}

func serializeBasicAuthSecret(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.AddBasicComponentTemplateBasicAuthSecret) error {
	obj := &model.AddBasicComponentTemplateGenericSecret{
		Data: map[string]string{
			"username": spec.AuthUser,
			"password": spec.AuthPassword,
		},
		Type: "kubernetes.io/basic-auth",
	}
	return serializeGenericSecret(buf, ser, name, labels, obj)
}

func serializeSSHAuthSecret(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.AddBasicComponentTemplateSSHAuthSecret) error {
	obj := &model.AddBasicComponentTemplateGenericSecret{
		Data: map[string]string{
			"ssh-privatekey": spec.PrivateKey,
		},
		Type: "kubernetes.io/ssh-auth",
	}
	return serializeGenericSecret(buf, ser, name, labels, obj)
}

func serializeTLSSecret(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.AddBasicComponentTemplateTLSSecret) error {
	obj := &model.AddBasicComponentTemplateGenericSecret{
		Data: map[string]string{
			"tls.key": spec.Key,
			"tls.crt": spec.Cert,
		},
		Type: "kubernetes.io/tls",
	}
	return serializeGenericSecret(buf, ser, name, labels, obj)
}

func serializeConfigMap(buf io.Writer, ser runtime.Serializer, params *model.AddBasicComponentTemplateConfigMap) error {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    params.Labels,
		},
		Data: params.Data,
	}
	return ser.Encode(cm, buf)
}

func serializePersistentVolumeClaim(buf io.Writer, ser runtime.Serializer, params *model.AddBasicComponentTemplatePersistentVolumeClaim) error {
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    params.Labels,
		},
	}
	convertK8sPersistentVolumeClaimSpec(&pvc.Spec, params.Spec)
	return ser.Encode(pvc, buf)
}

func convertK8sPersistentVolumeClaimSpec(spec *corev1.PersistentVolumeClaimSpec, obj *model.CreatePersistentVolumeClaimSpec) {
	for _, each := range obj.AccessModes {
		switch each {
		case "ReadOnlyMany":
			spec.AccessModes = append(spec.AccessModes, corev1.ReadOnlyMany)
		case "ReadWriteMany":
			spec.AccessModes = append(spec.AccessModes, corev1.ReadWriteMany)
		default:
			spec.AccessModes = append(spec.AccessModes, corev1.ReadWriteOnce)
		}
	}
	if obj.Resources != nil {
		if obj.Resources.Limits != nil {
			spec.Resources.Limits = make(corev1.ResourceList)
			if obj.Resources.Limits.Storage != nil {
				v, err := resource.ParseQuantity(*obj.Resources.Limits.Storage)
				if err == nil {
					spec.Resources.Limits[corev1.ResourceStorage] = v
				}
			}
		}
		if obj.Resources.Requests != nil {
			spec.Resources.Requests = make(corev1.ResourceList)
			v, err := resource.ParseQuantity(obj.Resources.Requests.Storage)
			if err == nil {
				spec.Resources.Requests[corev1.ResourceStorage] = v
			}
		}
	}
	if obj.StorageClass != nil {
		spec.StorageClassName = obj.StorageClass
	}
	if obj.VolumeMode != nil {
		v := corev1.PersistentVolumeMode(*obj.VolumeMode)
		spec.VolumeMode = &v
	}
}

func serializeHorizontalPodAutoscaler(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.CreateHorizontalPodAutoscalerSpec) error {
	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    labels,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			// refer to deployment of the same name
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       name,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		},
	}
	if spec.MinReplicas != nil {
		hpa.Spec.MinReplicas = util.Int64PtrToInt32Ptr(spec.MinReplicas)
	} else {
		hpa.Spec.MinReplicas = util.Int32Ptr(1)
	}
	if spec.MaxReplicas != nil && *spec.MaxReplicas > int64(*hpa.Spec.MinReplicas) {
		hpa.Spec.MaxReplicas = int32(*spec.MaxReplicas)
	} else {
		hpa.Spec.MaxReplicas = *hpa.Spec.MinReplicas
	}
	if spec.TargetCPUUtilizationPercentage != nil {
		hpa.Spec.Metrics = append(hpa.Spec.Metrics, autoscalingv2beta2.MetricSpec{
			Type: autoscalingv2beta2.ResourceMetricSourceType,
			Resource: &autoscalingv2beta2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.UtilizationMetricType,
					AverageUtilization: util.Int64PtrToInt32Ptr(spec.TargetCPUUtilizationPercentage),
				},
			},
		})
	}
	if spec.TargetMemoryUtilizationPercentage != nil {
		hpa.Spec.Metrics = append(hpa.Spec.Metrics, autoscalingv2beta2.MetricSpec{
			Type: autoscalingv2beta2.ResourceMetricSourceType,
			Resource: &autoscalingv2beta2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2beta2.MetricTarget{
					Type:               autoscalingv2beta2.UtilizationMetricType,
					AverageUtilization: util.Int64PtrToInt32Ptr(spec.TargetMemoryUtilizationPercentage),
				},
			},
		})
	}
	return ser.Encode(hpa, buf)
}

func serializeDeployment(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.CreateDeploymentSpec) error {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}
	if spec.Strategy != nil {
		if spec.Strategy.Type != nil && *spec.Strategy.Type != "" {
			deployment.Spec.Strategy.Type = appsv1.DeploymentStrategyType(*spec.Strategy.Type)
		}
		if spec.Strategy.RollingUpdate != nil {
			deployment.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
				MaxUnavailable: spec.Strategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       spec.Strategy.RollingUpdate.MaxSurge,
			}
		}
	}
	if spec.Replicas != nil {
		deployment.Spec.Replicas = util.Int32Ptr(int32(*spec.Replicas))
	}
	if spec.ProgressDeadlineSeconds != nil {
		deployment.Spec.ProgressDeadlineSeconds = util.Int32Ptr(int32(*spec.ProgressDeadlineSeconds))
	}
	if spec.RevisionHistoryLimit != nil {
		deployment.Spec.RevisionHistoryLimit = util.Int32Ptr(int32(*spec.RevisionHistoryLimit))
	}
	convertK8sPodSpec(&deployment.Spec.Template.Spec, spec.Template.Spec)
	return ser.Encode(deployment, buf)
}

func convertK8sPodSpec(spec *corev1.PodSpec, obj *model.CreatePodSpec) {
	for _, each := range obj.InitContainers {
		container := corev1.Container{}
		convertK8sContainer(&container, each)
		spec.InitContainers = append(spec.InitContainers, container)
	}
	for _, each := range obj.Containers {
		container := corev1.Container{}
		convertK8sContainer(&container, each)
		spec.Containers = append(spec.Containers, container)
	}
	if obj.DNSPolicy != nil {
		spec.DNSPolicy = corev1.DNSPolicy(*obj.DNSPolicy)
	}
	if obj.RestartPolicy != nil {
		spec.RestartPolicy = corev1.RestartPolicy(*obj.RestartPolicy)
	}
	if obj.SchedulerName != nil {
		spec.SchedulerName = *obj.SchedulerName
	}
	if obj.ServiceAccount != nil {
		spec.ServiceAccountName = *obj.ServiceAccount
	}
	spec.TerminationGracePeriodSeconds = obj.TerminationGracePeriodSeconds
	for _, each := range obj.Volumes {
		vol := corev1.Volume{Name: each.Name}
		switch {
		case each.HostPath != nil:
			vol.HostPath = &corev1.HostPathVolumeSource{Path: each.HostPath.Path}
			if each.HostPath.Type != nil {
				v := corev1.HostPathType(*each.HostPath.Type)
				vol.HostPath.Type = &v
			}
		case each.EmptyDir != nil:
			vol.EmptyDir = &corev1.EmptyDirVolumeSource{}
			if each.EmptyDir.Medium != nil {
				vol.EmptyDir.Medium = corev1.StorageMedium(*each.EmptyDir.Medium)
			}
			if each.EmptyDir.Size != nil {
				v, err := resource.ParseQuantity(*each.EmptyDir.Size)
				if err == nil {
					vol.EmptyDir.SizeLimit = &v
				}
			}
		case each.ConfigMap != nil:
			vol.ConfigMap = &corev1.ConfigMapVolumeSource{}
			vol.ConfigMap.Name = each.ConfigMap.Name
			for _, item := range each.ConfigMap.Items {
				p := corev1.KeyToPath{Key: item.Key, Path: item.Path}
				if item.Mode != nil {
					p.Mode = util.Int32Ptr(int32(*item.Mode))
				}
				vol.ConfigMap.Items = append(vol.ConfigMap.Items, p)
			}
			if each.ConfigMap.DefaultMode != nil {
				vol.ConfigMap.DefaultMode = util.Int32Ptr(int32(*each.ConfigMap.DefaultMode))
			}
		case each.Secret != nil:
			vol.Secret = &corev1.SecretVolumeSource{SecretName: each.Secret.Name}
			for _, item := range each.Secret.Items {
				p := corev1.KeyToPath{Key: item.Key, Path: item.Path}
				if item.Mode != nil {
					p.Mode = util.Int32Ptr(int32(*item.Mode))
				}
				vol.Secret.Items = append(vol.Secret.Items, p)
			}
			if each.Secret.DefaultMode != nil {
				vol.Secret.DefaultMode = util.Int32Ptr(int32(*each.Secret.DefaultMode))
			}
		case each.PersistentVolumeClaim != nil:
			vol.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{ClaimName: each.PersistentVolumeClaim.Name}
			if each.PersistentVolumeClaim.ReadOnly != nil {
				vol.PersistentVolumeClaim.ReadOnly = *each.PersistentVolumeClaim.ReadOnly
			}
		}
		spec.Volumes = append(spec.Volumes, vol)
	}
	for _, each := range obj.ImagePullSecrets {
		spec.ImagePullSecrets = append(spec.ImagePullSecrets, corev1.LocalObjectReference{Name: each})
	}
}

func convertK8sContainer(container *corev1.Container, obj *model.AddPodSpecContainer) {
	container.Name = obj.Name
	container.Image = obj.Image
	container.Command = obj.Command
	container.Args = obj.Args
	if obj.WorkingDir != nil {
		container.WorkingDir = *obj.WorkingDir
	}
	for _, each := range obj.Ports {
		port := corev1.ContainerPort{
			ContainerPort: int32(each.ContainerPort),
		}
		if each.Protocol == nil || *each.Protocol == "" {
			each.Protocol = util.StringPtr("TCP")
		}
		switch strings.ToUpper(*each.Protocol) {
		case "UDP":
			port.Protocol = corev1.ProtocolUDP
		case "SCTP":
			port.Protocol = corev1.ProtocolSCTP
		default:
			port.Protocol = corev1.ProtocolTCP
		}
		if each.Name == nil || *each.Name == "" {
			each.Name = util.StringPtr(fmt.Sprintf("%s-%d", strings.ToLower(string(port.Protocol)), port.ContainerPort))
		}
		port.Name = *each.Name
		container.Ports = append(container.Ports, port)
	}
	for k, v := range obj.Env {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	if obj.Resources != nil {
		if obj.Resources.Requests != nil {
			container.Resources.Requests = make(corev1.ResourceList)
			if obj.Resources.Requests.CPU != nil {
				v, err := resource.ParseQuantity(*obj.Resources.Requests.CPU)
				if err == nil {
					container.Resources.Requests[corev1.ResourceCPU] = v
				}
			}
			if obj.Resources.Requests.Memory != nil {
				v, err := resource.ParseQuantity(*obj.Resources.Requests.Memory)
				if err == nil {
					container.Resources.Requests[corev1.ResourceMemory] = v
				}
			}
		}
		if obj.Resources.Limits != nil {
			container.Resources.Limits = make(corev1.ResourceList)
			if obj.Resources.Limits.CPU != nil {
				v, err := resource.ParseQuantity(*obj.Resources.Limits.CPU)
				if err == nil {
					container.Resources.Limits[corev1.ResourceCPU] = v
				}
			}
			if obj.Resources.Limits.Memory != nil {
				v, err := resource.ParseQuantity(*obj.Resources.Limits.Memory)
				if err == nil {
					container.Resources.Limits[corev1.ResourceMemory] = v
				}
			}
		}
	}
	for _, mount := range obj.VolumeMounts {
		val := corev1.VolumeMount{Name: mount.Name, MountPath: mount.MountPath}
		if mount.ReadOnly != nil {
			val.ReadOnly = *mount.ReadOnly
		}
		if mount.SubPath != nil {
			val.SubPath = *mount.SubPath
		} else if mount.SubPathExpr != nil {
			val.SubPathExpr = *mount.SubPathExpr
		}
		if mount.MountPropagation != nil {
			v := corev1.MountPropagationMode(*mount.MountPropagation)
			val.MountPropagation = &v
		}
		container.VolumeMounts = append(container.VolumeMounts, val)
	}
	if obj.LivenessProbe != nil {
		container.LivenessProbe = &corev1.Probe{}
		convertK8sContainerProbe(container.LivenessProbe, obj.LivenessProbe)
	}
	if obj.ReadinessProbe != nil {
		container.ReadinessProbe = &corev1.Probe{}
		convertK8sContainerProbe(container.ReadinessProbe, obj.ReadinessProbe)
	}
	if obj.StartupProbe != nil {
		container.StartupProbe = &corev1.Probe{}
		convertK8sContainerProbe(container.StartupProbe, obj.StartupProbe)
	}
	if obj.Lifecycle != nil {
		container.Lifecycle = &corev1.Lifecycle{}
		if obj.Lifecycle.PostStart != nil {
			container.Lifecycle.PostStart = &corev1.LifecycleHandler{}
			convertK8sContainerHandler(container.Lifecycle.PostStart, obj.Lifecycle.PostStart)
		}
		if obj.Lifecycle.PreStop != nil {
			container.Lifecycle.PreStop = &corev1.LifecycleHandler{}
			convertK8sContainerHandler(container.Lifecycle.PreStop, obj.Lifecycle.PreStop)
		}
	}
	if obj.ImagePullPolicy != nil {
		container.ImagePullPolicy = corev1.PullPolicy(*obj.ImagePullPolicy)
	}
}

func convertK8sContainerProbe(probe *corev1.Probe, obj *model.InputPodSpecContainerProbe) {
	if obj.InitialDelaySeconds != nil {
		probe.InitialDelaySeconds = int32(*obj.InitialDelaySeconds)
	}
	if obj.TimeoutSeconds != nil {
		probe.TimeoutSeconds = int32(*obj.TimeoutSeconds)
	}
	if obj.PeriodSeconds != nil {
		probe.PeriodSeconds = int32(*obj.PeriodSeconds)
	}
	if obj.SuccessThreshold != nil {
		probe.SuccessThreshold = int32(*obj.SuccessThreshold)
	}
	if obj.FailureThreshold != nil {
		probe.FailureThreshold = int32(*obj.FailureThreshold)
	}
	probe.TerminationGracePeriodSeconds = obj.TerminationGracePeriodSeconds
	switch {
	case obj.Exec != nil:
		probe.Exec = &corev1.ExecAction{Command: obj.Exec.Command}
	case obj.HTTPGet != nil:
		probe.HTTPGet = &corev1.HTTPGetAction{}
		if obj.HTTPGet.Path != nil {
			probe.HTTPGet.Path = *obj.HTTPGet.Path
		}
		if obj.HTTPGet.Port != nil {
			probe.HTTPGet.Port = *obj.HTTPGet.Port
		}
		if obj.HTTPGet.Host != nil {
			probe.HTTPGet.Host = *obj.HTTPGet.Host
		}
		if obj.HTTPGet.Scheme != nil {
			probe.HTTPGet.Scheme = corev1.URIScheme(*obj.HTTPGet.Scheme)
		}
		for _, header := range obj.HTTPGet.HTTPHeaders {
			probe.HTTPGet.HTTPHeaders = append(probe.HTTPGet.HTTPHeaders, corev1.HTTPHeader{Name: header.Name, Value: header.Value})
		}
	case obj.TCPSocket != nil:
		probe.TCPSocket = &corev1.TCPSocketAction{}
		if obj.TCPSocket.Port != nil {
			probe.TCPSocket.Port = *obj.TCPSocket.Port
		}
		if obj.TCPSocket.Host != nil {
			probe.TCPSocket.Host = *obj.TCPSocket.Host
		}
	}
}

func convertK8sContainerHandler(handler *corev1.LifecycleHandler, obj *model.InputPodSpecContainerHandler) {
	switch {
	case obj.Exec != nil:
		handler.Exec = &corev1.ExecAction{Command: obj.Exec.Command}
	case obj.HTTPGet != nil:
		handler.HTTPGet = &corev1.HTTPGetAction{}
		if obj.HTTPGet.Path != nil {
			handler.HTTPGet.Path = *obj.HTTPGet.Path
		}
		if obj.HTTPGet.Port != nil {
			handler.HTTPGet.Port = *obj.HTTPGet.Port
		}
		if obj.HTTPGet.Host != nil {
			handler.HTTPGet.Host = *obj.HTTPGet.Host
		}
		if obj.HTTPGet.Scheme != nil {
			handler.HTTPGet.Scheme = corev1.URIScheme(*obj.HTTPGet.Scheme)
		}
		for _, header := range obj.HTTPGet.HTTPHeaders {
			handler.HTTPGet.HTTPHeaders = append(handler.HTTPGet.HTTPHeaders, corev1.HTTPHeader{Name: header.Name, Value: header.Value})
		}
	case obj.TCPSocket != nil:
		handler.TCPSocket = &corev1.TCPSocketAction{}
		if obj.TCPSocket.Port != nil {
			handler.TCPSocket.Port = *obj.TCPSocket.Port
		}
		if obj.TCPSocket.Host != nil {
			handler.TCPSocket.Host = *obj.TCPSocket.Host
		}
	}
}

func serializeService(buf io.Writer, ser runtime.Serializer, name string, labels map[string]string, spec *model.CreateServiceSpec) error {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: manifestTemplateNSPlaceholder,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
		},
	}
	for _, each := range spec.Ports {
		port := corev1.ServicePort{
			Port:       int32(each.Port),
			TargetPort: each.TargetPort,
		}
		if each.Protocol == nil || *each.Protocol == "" {
			each.Protocol = util.StringPtr("TCP")
		}
		switch strings.ToUpper(*each.Protocol) {
		case "UDP":
			port.Protocol = corev1.ProtocolUDP
		case "SCTP":
			port.Protocol = corev1.ProtocolSCTP
		default:
			port.Protocol = corev1.ProtocolTCP
		}
		if each.Name == nil || *each.Name == "" {
			each.Name = util.StringPtr(fmt.Sprintf("%s-%d", strings.ToLower(string(port.Protocol)), port.Port))
		}
		port.Name = *each.Name
		service.Spec.Ports = append(service.Spec.Ports, port)
	}
	if spec.Type == nil || *spec.Type == "" {
		spec.Type = util.StringPtr("ClusterIP")
	}
	switch *spec.Type {
	case "NodePort":
		service.Spec.Type = corev1.ServiceTypeNodePort
	case "LoadBalancer":
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
	case "ExternalName":
		service.Spec.Type = corev1.ServiceTypeExternalName
	default:
		service.Spec.Type = corev1.ServiceTypeClusterIP
	}
	return ser.Encode(service, buf)
}

func getCACertificateFromMatchedImageRegistries(tx *gorm.DB, url string) ([]byte, error) {
	var (
		buf        bytes.Buffer
		registries []model1.ImageRegistry
	)
	if err := tx.Where("`endpoint` LIKE ?", url+"%").Find(&registries).Error; err != nil {
		return nil, err
	}
	for _, registry := range registries {
		buf.Write(bytes.TrimSpace(registry.CACert))
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}
