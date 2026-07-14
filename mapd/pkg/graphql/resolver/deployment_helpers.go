// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func getDeployments(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetDeploymentsContext(ctx, namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertDeployment(deployment *model.Deployment, obj *appsv1.Deployment) {
	deployment.Name = obj.ObjectMeta.Name
	deployment.Namespace = obj.ObjectMeta.Namespace
	deployment.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			id, err := uuid.Parse(app)
			if err == nil {
				deployment.Application = &model.Application{ID: id}
			}
		}
		com, ok := obj.ObjectMeta.Labels[defLabelComponent]
		if ok {
			id, err := uuid.Parse(com)
			if err == nil {
				deployment.Component = &model.Component{ID: id}
			}
		}
		ins, ok := obj.ObjectMeta.Labels[defLabelInstance]
		if ok {
			id, err := uuid.Parse(ins)
			if err == nil {
				deployment.Instance = &model.Instance{ID: id}
			}
		}
	}
	deployment.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		deployment.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		deployment.CreatedAt = time.Now()
	}

	deployment.Spec = &model.DeploymentSpec{}
	if obj.Spec.ProgressDeadlineSeconds != nil {
		deployment.Spec.ProgressDeadlineSeconds = util.Int32PtrToInt64Ptr(obj.Spec.ProgressDeadlineSeconds)
	}
	if obj.Spec.Replicas != nil {
		deployment.Spec.Replicas = util.Int32PtrToInt64Ptr(obj.Spec.Replicas)
	}
	if obj.Spec.RevisionHistoryLimit != nil {
		deployment.Spec.RevisionHistoryLimit = util.Int32PtrToInt64Ptr(obj.Spec.RevisionHistoryLimit)
	}
	deployment.Spec.Template = &model.PodTemplate{}
	convertPodTemplate(deployment.Spec.Template, obj.Spec.Template, deployment.Namespace, deployment.Cluster)
	deployment.Spec.Strategy = &model.DeploymentStrategy{}
	convertDeploymentStrategy(deployment.Spec.Strategy, obj.Spec.Strategy)

	deployment.Status = &model.DeploymentStatus{
		AvailableReplicas:   int64(obj.Status.AvailableReplicas),
		ReadyReplicas:       int64(obj.Status.ReadyReplicas),
		Replicas:            int64(obj.Status.Replicas),
		UnavailableReplicas: int64(obj.Status.UnavailableReplicas),
		UpdatedReplicas:     int64(obj.Status.UpdatedReplicas),
	}

	if deployment.CreatedAt.IsZero() {
		deployment.CreatedAt = time.Unix(0, 0)
	}
	if deployment.LastModified.IsZero() {
		deployment.LastModified = deployment.CreatedAt
	}
}

func convertDeploymentStrategy(stg *model.DeploymentStrategy, obj appsv1.DeploymentStrategy) {
	stg.Type = util.StringPtr(string(obj.Type))
	if obj.RollingUpdate != nil {
		stg.RollingUpdate = &model.RollingUpdateDeployment{
			MaxUnavailable: obj.RollingUpdate.MaxUnavailable,
			MaxSurge:       obj.RollingUpdate.MaxSurge,
		}
	}
}

func convertPodTemplate(template *model.PodTemplate, obj corev1.PodTemplateSpec, namespace string, cluster *model.Cluster) {
	template.Labels = obj.Labels
	template.Spec = &model.PodSpec{}
	for _, each := range obj.Spec.InitContainers {
		container := &model.PodSpecContainer{}
		convertPodContainer(container, each)
		template.Spec.InitContainers = append(template.Spec.InitContainers, container)
	}
	for _, each := range obj.Spec.Containers {
		container := &model.PodSpecContainer{}
		convertPodContainer(container, each)
		template.Spec.Containers = append(template.Spec.Containers, container)
	}
	if obj.Spec.DNSPolicy != "" {
		template.Spec.DNSPolicy = util.StringPtr(string(obj.Spec.DNSPolicy))
	}
	if obj.Spec.RestartPolicy != "" {
		template.Spec.RestartPolicy = util.StringPtr(string(obj.Spec.RestartPolicy))
	}
	if obj.Spec.SchedulerName != "" {
		template.Spec.SchedulerName = &obj.Spec.SchedulerName
	}
	if obj.Spec.ServiceAccountName != "" {
		template.Spec.ServiceAccount = &obj.Spec.ServiceAccountName
	}
	template.Spec.TerminationGracePeriodSeconds = obj.Spec.TerminationGracePeriodSeconds
	for _, vol := range obj.Spec.Volumes {
		v := &model.PodSpecVolume{
			Name: &vol.Name,
		}
		switch {
		case vol.HostPath != nil:
			val := &model.HostPath{Path: &vol.HostPath.Path}
			if vol.HostPath.Type != nil {
				val.Type = util.StringPtr(string(*vol.HostPath.Type))
			}
			v.Volume = val
		case vol.EmptyDir != nil:
			val := &model.EmptyDir{}
			if vol.EmptyDir.Medium != "" {
				val.Medium = util.StringPtr(string(vol.EmptyDir.Medium))
			}
			if vol.EmptyDir.SizeLimit != nil {
				val.Size = util.StringPtr(vol.EmptyDir.SizeLimit.String())
			}
			v.Volume = val
		case vol.ConfigMap != nil:
			val := &model.ConfigMapVolume{Name: &vol.ConfigMap.Name}
			val.Ref = &model.ConfigMap{
				Name:      vol.ConfigMap.Name,
				Namespace: namespace,
				Cluster:   cluster,
			}
			for _, each := range vol.ConfigMap.Items {
				p := &model.TemplateSpecVolumeKeyToPath{Key: &each.Key, Path: &each.Path}
				if each.Mode != nil {
					p.Mode = util.Int32PtrToInt64Ptr(each.Mode)
				}
				val.Items = append(val.Items, p)
			}
			if vol.ConfigMap.DefaultMode != nil {
				val.DefaultMode = util.Int32PtrToInt64Ptr(vol.ConfigMap.DefaultMode)
			}
			v.Volume = val
		case vol.Secret != nil:
			val := &model.SecretVolume{Name: &vol.Secret.SecretName}
			val.Ref = &model.Secret{
				Name:      vol.Secret.SecretName,
				Namespace: namespace,
				Cluster:   cluster,
			}
			for _, each := range vol.Secret.Items {
				p := &model.TemplateSpecVolumeKeyToPath{Key: &each.Key, Path: &each.Path}
				if each.Mode != nil {
					p.Mode = util.Int32PtrToInt64Ptr(each.Mode)
				}
				val.Items = append(val.Items, p)
			}
			if vol.Secret.DefaultMode != nil {
				val.DefaultMode = util.Int32PtrToInt64Ptr(vol.Secret.DefaultMode)
			}
			v.Volume = val
		case vol.PersistentVolumeClaim != nil:
			val := &model.PersistentVolumeClaimVolume{Name: &vol.PersistentVolumeClaim.ClaimName}
			val.Ref = &model.PersistentVolumeClaim{
				Name:      vol.PersistentVolumeClaim.ClaimName,
				Namespace: namespace,
				Cluster:   cluster,
			}
			val.ReadOnly = &vol.PersistentVolumeClaim.ReadOnly
			v.Volume = val
		}
		template.Spec.Volumes = append(template.Spec.Volumes, v)
	}
	for _, secret := range obj.Spec.ImagePullSecrets {
		template.Spec.ImagePullSecrets = append(template.Spec.ImagePullSecrets, secret.Name)
	}
}

func convertDeploymentSpecFromInput(input *model.CreateDeploymentSpec) *model.DeploymentSpec {
	spec := &model.DeploymentSpec{
		ProgressDeadlineSeconds: input.ProgressDeadlineSeconds,
		Replicas:                input.Replicas,
		RevisionHistoryLimit:    input.RevisionHistoryLimit,
		Template: &model.PodTemplate{
			Spec: &model.PodSpec{
				DNSPolicy:                     input.Template.Spec.DNSPolicy,
				RestartPolicy:                 input.Template.Spec.RestartPolicy,
				SchedulerName:                 input.Template.Spec.SchedulerName,
				ServiceAccount:                input.Template.Spec.ServiceAccount,
				TerminationGracePeriodSeconds: input.Template.Spec.TerminationGracePeriodSeconds,
				ImagePullSecrets:              input.Template.Spec.ImagePullSecrets,
			},
		},
	}
	for _, each := range input.Template.Spec.InitContainers {
		container := &model.PodSpecContainer{}
		convertPodSpecContainerFromInput(container, each)
		spec.Template.Spec.InitContainers = append(spec.Template.Spec.InitContainers, container)
	}
	for _, each := range input.Template.Spec.Containers {
		container := &model.PodSpecContainer{}
		convertPodSpecContainerFromInput(container, each)
		spec.Template.Spec.Containers = append(spec.Template.Spec.Containers, container)
	}
	for _, vol := range input.Template.Spec.Volumes {
		v := &model.PodSpecVolume{
			Name: &vol.Name,
		}
		switch {
		case vol.HostPath != nil:
			val := &model.HostPath{Path: &vol.HostPath.Path}
			if vol.HostPath.Type != nil {
				val.Type = util.StringPtr(string(*vol.HostPath.Type))
			}
			v.Volume = val
		case vol.EmptyDir != nil:
			val := &model.EmptyDir{
				Medium: vol.EmptyDir.Medium,
				Size:   vol.EmptyDir.Size,
			}
			v.Volume = val
		case vol.ConfigMap != nil:
			val := &model.ConfigMapVolume{
				Name:        &vol.ConfigMap.Name,
				DefaultMode: vol.ConfigMap.DefaultMode,
			}
			for _, each := range vol.ConfigMap.Items {
				p := &model.TemplateSpecVolumeKeyToPath{
					Key:  &each.Key,
					Path: &each.Path,
					Mode: each.Mode,
				}
				val.Items = append(val.Items, p)
			}
			v.Volume = val
		case vol.Secret != nil:
			val := &model.SecretVolume{
				Name:        &vol.Secret.Name,
				DefaultMode: vol.Secret.DefaultMode,
			}
			for _, each := range vol.Secret.Items {
				p := &model.TemplateSpecVolumeKeyToPath{
					Key:  &each.Key,
					Path: &each.Path,
					Mode: each.Mode,
				}
				val.Items = append(val.Items, p)
			}
			v.Volume = val
		case vol.PersistentVolumeClaim != nil:
			val := &model.PersistentVolumeClaimVolume{
				Name:     &vol.PersistentVolumeClaim.Name,
				ReadOnly: vol.PersistentVolumeClaim.ReadOnly,
			}
			v.Volume = val
		}
		spec.Template.Spec.Volumes = append(spec.Template.Spec.Volumes, v)
	}
	if input.Strategy != nil {
		spec.Strategy = &model.DeploymentStrategy{
			Type: input.Strategy.Type,
		}
		if input.Strategy.RollingUpdate != nil {
			spec.Strategy.RollingUpdate = &model.RollingUpdateDeployment{
				MaxUnavailable: input.Strategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       input.Strategy.RollingUpdate.MaxSurge,
			}
		}
	}
	return spec
}

func convertHorizontalPodAutoscalerSpecFromInput(input *model.CreateHorizontalPodAutoscalerSpec) *model.HorizontalPodAutoscalerSpec {
	return &model.HorizontalPodAutoscalerSpec{
		MaxReplicas:                       input.MaxReplicas,
		MinReplicas:                       input.MinReplicas,
		TargetCPUUtilizationPercentage:    input.TargetCPUUtilizationPercentage,
		TargetMemoryUtilizationPercentage: input.TargetMemoryUtilizationPercentage,
	}
}

func convertPodSpecContainerFromInput(container *model.PodSpecContainer, obj *model.AddPodSpecContainer) {
	container.Name = obj.Name
	container.Image = obj.Image
	container.Command = obj.Command
	container.Args = obj.Args
	container.Env = obj.Env
	container.WorkingDir = obj.WorkingDir
	for _, port := range obj.Ports {
		p := &model.PodSpecContainerPort{
			Name:          port.Name,
			ContainerPort: util.Int64Ptr(port.ContainerPort),
			Protocol:      port.Protocol,
		}
		container.Ports = append(container.Ports, p)
	}
	if obj.Resources != nil {
		container.Resources = &model.PodSpecContainerResources{}
		if obj.Resources.Limits != nil {
			container.Resources.Limits = &model.PodSpecContainerResourcesLimits{
				CPU:    obj.Resources.Limits.CPU,
				Memory: obj.Resources.Limits.Memory,
			}
		}
		if obj.Resources.Requests != nil {
			container.Resources.Requests = &model.PodSpecContainerResourcesRequests{
				CPU:    obj.Resources.Requests.CPU,
				Memory: obj.Resources.Requests.Memory,
			}
		}
	}
	for _, vol := range obj.VolumeMounts {
		v := &model.PodSpecContainerVolumeMount{
			Name:             &vol.Name,
			ReadOnly:         vol.ReadOnly,
			MountPath:        &vol.MountPath,
			SubPath:          vol.SubPath,
			SubPathExpr:      vol.SubPathExpr,
			MountPropagation: vol.MountPropagation,
		}
		container.VolumeMounts = append(container.VolumeMounts, v)
	}
	if probe := obj.LivenessProbe; probe != nil {
		container.LivenessProbe = convertPodSpecContainerProbeFromInput(probe)
	}
	if probe := obj.ReadinessProbe; probe != nil {
		container.ReadinessProbe = convertPodSpecContainerProbeFromInput(probe)
	}
	if probe := obj.StartupProbe; probe != nil {
		container.StartupProbe = convertPodSpecContainerProbeFromInput(probe)
	}
	if obj.Lifecycle != nil {
		container.Lifecycle = &model.PodSpecContainerLifecycle{}
		if obj.Lifecycle.PostStart != nil {
			container.Lifecycle.PostStart = convertPodSpecContainerHandlerFromInput(obj.Lifecycle.PostStart)
		}
		if obj.Lifecycle.PreStop != nil {
			container.Lifecycle.PreStop = convertPodSpecContainerHandlerFromInput(obj.Lifecycle.PreStop)
		}
	}
	container.ImagePullPolicy = obj.ImagePullPolicy
}

func convertPodSpecContainerProbeFromInput(probe *model.InputPodSpecContainerProbe) *model.PodSpecContainerProbe {
	obj := &model.PodSpecContainerProbe{}
	if probe.Exec != nil {
		obj.Exec = &model.PodSpecContainerExecAction{Command: probe.Exec.Command}
	}
	if probe.HTTPGet != nil {
		obj.HTTPGet = &model.PodSpecContainerHTTPGetAction{
			Path:   probe.HTTPGet.Path,
			Port:   probe.HTTPGet.Port,
			Host:   probe.HTTPGet.Host,
			Scheme: probe.HTTPGet.Scheme,
		}
		for _, header := range probe.HTTPGet.HTTPHeaders {
			h := &model.PodSpecContainerHandlerHTTPGetActionHTTPHeader{
				Name:  &header.Name,
				Value: &header.Value,
			}
			obj.HTTPGet.HTTPHeaders = append(obj.HTTPGet.HTTPHeaders, h)
		}
	}
	if probe.TCPSocket != nil {
		obj.TCPSocket = &model.PodSpecContainerTCPSocketAction{
			Port: probe.TCPSocket.Port,
			Host: probe.TCPSocket.Host,
		}
	}
	obj.InitialDelaySeconds = probe.InitialDelaySeconds
	obj.TimeoutSeconds = probe.TimeoutSeconds
	obj.PeriodSeconds = probe.PeriodSeconds
	obj.SuccessThreshold = probe.SuccessThreshold
	obj.FailureThreshold = probe.FailureThreshold
	obj.TerminationGracePeriodSeconds = probe.TerminationGracePeriodSeconds
	return obj
}

func convertPodSpecContainerHandlerFromInput(handler *model.InputPodSpecContainerHandler) *model.PodSpecContainerHandler {
	obj := &model.PodSpecContainerHandler{}
	if handler.Exec != nil {
		obj.Exec = &model.PodSpecContainerExecAction{Command: handler.Exec.Command}
	}
	if handler.HTTPGet != nil {
		obj.HTTPGet = &model.PodSpecContainerHTTPGetAction{
			Path:   handler.HTTPGet.Path,
			Port:   handler.HTTPGet.Port,
			Host:   handler.HTTPGet.Host,
			Scheme: handler.HTTPGet.Scheme,
		}
		for _, header := range handler.HTTPGet.HTTPHeaders {
			h := &model.PodSpecContainerHandlerHTTPGetActionHTTPHeader{
				Name:  &header.Name,
				Value: &header.Value,
			}
			obj.HTTPGet.HTTPHeaders = append(obj.HTTPGet.HTTPHeaders, h)
		}
	}
	if handler.TCPSocket != nil {
		obj.TCPSocket = &model.PodSpecContainerTCPSocketAction{
			Port: handler.TCPSocket.Port,
			Host: handler.TCPSocket.Host,
		}
	}
	return obj
}
