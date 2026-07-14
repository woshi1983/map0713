// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
)

func getPods(client *k8s.Kubernetes, namespace string, selectors ...k8s.Selector) ([]corev1.Pod, error) {
	list, err := client.GetPods(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertPod(pod *model.Pod, obj *corev1.Pod) {
	pod.Name = obj.ObjectMeta.Name
	pod.Namespace = obj.ObjectMeta.Namespace
	if obj.ObjectMeta.Labels != nil {
		app, ok := obj.ObjectMeta.Labels[defLabelApplication]
		if ok {
			pod.Application = &model.Application{Name: app}
		}
	}
	pod.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		pod.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		pod.CreatedAt = time.Now()
	}

	pod.Spec = &model.PodSpec{}
	for _, each := range obj.Spec.InitContainers {
		container := &model.PodSpecContainer{}
		convertPodContainer(container, each)
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)
	}
	for _, each := range obj.Spec.Containers {
		container := &model.PodSpecContainer{}
		convertPodContainer(container, each)
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}
	if obj.Spec.DNSPolicy != "" {
		pod.Spec.DNSPolicy = util.StringPtr(string(obj.Spec.DNSPolicy))
	}
	if obj.Spec.RestartPolicy != "" {
		pod.Spec.RestartPolicy = util.StringPtr(string(obj.Spec.RestartPolicy))
	}
	if obj.Spec.SchedulerName != "" {
		pod.Spec.SchedulerName = &obj.Spec.SchedulerName
	}
	if obj.Spec.ServiceAccountName != "" {
		pod.Spec.ServiceAccount = &obj.Spec.ServiceAccountName
	}
	pod.Spec.TerminationGracePeriodSeconds = obj.Spec.TerminationGracePeriodSeconds
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
				Namespace: pod.Namespace,
				Cluster:   pod.Cluster,
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
				Namespace: pod.Namespace,
				Cluster:   pod.Cluster,
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
				Namespace: pod.Namespace,
				Cluster:   pod.Cluster,
			}
			val.ReadOnly = &vol.PersistentVolumeClaim.ReadOnly
			v.Volume = val
		}
		pod.Spec.Volumes = append(pod.Spec.Volumes, v)
	}
	for _, secret := range obj.Spec.ImagePullSecrets {
		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, secret.Name)
	}

	pod.Status = &model.PodStatus{}
	if obj.Status.Phase != "" {
		pod.Status.Phase = util.StringPtr(string(obj.Status.Phase))
	}
	if obj.Status.Message != "" {
		pod.Status.Message = &obj.Status.Message
	}
	if obj.Status.Reason != "" {
		pod.Status.Reason = &obj.Status.Reason
	}
	if obj.Status.HostIP != "" {
		pod.Status.HostIP = &obj.Status.HostIP
	}
	for _, each := range obj.Status.PodIPs {
		pod.Status.PodIPs = append(pod.Status.PodIPs, each.IP)
	}
	if len(pod.Status.PodIPs) == 0 && obj.Status.PodIP != "" {
		pod.Status.PodIPs = append(pod.Status.PodIPs, obj.Status.PodIP)
	}
	for _, each := range obj.Status.InitContainerStatuses {
		status := &model.PodStatusContainerStatus{
			Ready:        &each.Ready,
			RestartCount: util.Int64Ptr(int64(each.RestartCount)),
			Started:      each.Started,
		}
		if each.ContainerID != "" {
			status.ContainerID = &each.ContainerID
		}
		if each.Image != "" {
			status.Image = &each.Image
		}
		if each.ImageID != "" {
			status.ImageID = &each.ImageID
		}
		if each.Name != "" {
			status.Name = &each.Name
		}
		pod.Status.InitContainerStatuses = append(pod.Status.InitContainerStatuses, status)
	}
	if obj.Status.QOSClass != "" {
		pod.Status.QosClass = util.StringPtr(string(obj.Status.QOSClass))
	}

	if pod.CreatedAt.IsZero() {
		pod.CreatedAt = time.Unix(0, 0)
	}
	if pod.LastModified.IsZero() {
		pod.LastModified = pod.CreatedAt
	}
}

func convertPodContainer(container *model.PodSpecContainer, obj corev1.Container) {
	container.Name = obj.Name
	container.Image = obj.Image
	container.Command = obj.Command
	container.Args = obj.Args
	container.Env = make(map[string]string)
	if obj.WorkingDir != "" {
		container.WorkingDir = &obj.WorkingDir
	}
	for _, port := range obj.Ports {
		p := &model.PodSpecContainerPort{}
		if port.Name != "" {
			p.Name = &port.Name
		}
		if port.HostPort != 0 {
			p.HostPort = util.Int64Ptr(int64(port.HostPort))
		}
		if port.ContainerPort != 0 {
			p.ContainerPort = util.Int64Ptr(int64(port.ContainerPort))
		}
		if port.Protocol != "" {
			p.Protocol = util.StringPtr(string(port.Protocol))
		}
		if port.HostIP != "" {
			p.HostIP = &port.HostIP
		}
		container.Ports = append(container.Ports, p)
	}
	for _, env := range obj.Env {
		container.Env[env.Name] = env.Value
	}
	container.Resources = &model.PodSpecContainerResources{}
	if obj.Resources.Limits != nil {
		container.Resources.Limits = &model.PodSpecContainerResourcesLimits{}
		if v := obj.Resources.Limits.Cpu(); v != nil && !v.IsZero() {
			container.Resources.Limits.CPU = util.StringPtr(v.String())
		}
		if v := obj.Resources.Limits.Memory(); v != nil && !v.IsZero() {
			container.Resources.Limits.Memory = util.StringPtr(v.String())
		}
	}
	if obj.Resources.Requests != nil {
		container.Resources.Requests = &model.PodSpecContainerResourcesRequests{}
		if v := obj.Resources.Requests.Cpu(); v != nil && !v.IsZero() {
			container.Resources.Requests.CPU = util.StringPtr(v.String())
		}
		if v := obj.Resources.Requests.Memory(); v != nil && !v.IsZero() {
			container.Resources.Requests.Memory = util.StringPtr(v.String())
		}
	}
	for _, vol := range obj.VolumeMounts {
		v := &model.PodSpecContainerVolumeMount{
			Name:      &vol.Name,
			ReadOnly:  &vol.ReadOnly,
			MountPath: &vol.MountPath,
		}
		if vol.SubPathExpr != "" {
			v.SubPathExpr = &vol.SubPathExpr
		} else {
			v.SubPath = &vol.SubPath
		}
		if vol.MountPropagation != nil {
			v.MountPropagation = util.StringPtr(string(*vol.MountPropagation))
		}
		container.VolumeMounts = append(container.VolumeMounts, v)
	}
	if probe := obj.LivenessProbe; probe != nil {
		container.LivenessProbe = convertPodSpecContainerProbe(probe)
	}
	if probe := obj.ReadinessProbe; probe != nil {
		container.ReadinessProbe = convertPodSpecContainerProbe(probe)
	}
	if probe := obj.StartupProbe; probe != nil {
		container.StartupProbe = convertPodSpecContainerProbe(probe)
	}
	if obj.Lifecycle != nil {
		container.Lifecycle = &model.PodSpecContainerLifecycle{}
		if obj.Lifecycle.PostStart != nil {
			container.Lifecycle.PostStart = convertPodSpecContainerHandler(obj.Lifecycle.PostStart)
		}
		if obj.Lifecycle.PreStop != nil {
			container.Lifecycle.PreStop = convertPodSpecContainerHandler(obj.Lifecycle.PreStop)
		}
	}
	if obj.ImagePullPolicy != "" {
		container.ImagePullPolicy = util.StringPtr(string(obj.ImagePullPolicy))
	}
}

func convertPodSpecContainerProbe(probe *corev1.Probe) *model.PodSpecContainerProbe {
	obj := &model.PodSpecContainerProbe{}
	if probe.Exec != nil {
		obj.Exec = &model.PodSpecContainerExecAction{Command: probe.Exec.Command}
	}
	if probe.HTTPGet != nil {
		obj.HTTPGet = &model.PodSpecContainerHTTPGetAction{}
		if probe.HTTPGet.Path != "" {
			obj.HTTPGet.Path = &probe.HTTPGet.Path
		}
		if v := probe.HTTPGet.Port.String(); v != "" {
			obj.HTTPGet.Port = &probe.HTTPGet.Port
		}
		if probe.HTTPGet.Host != "" {
			obj.HTTPGet.Host = &probe.HTTPGet.Host
		}
		if probe.HTTPGet.Scheme != "" {
			obj.HTTPGet.Path = util.StringPtr(string(probe.HTTPGet.Scheme))
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
		obj.TCPSocket = &model.PodSpecContainerTCPSocketAction{}
		if v := probe.TCPSocket.Port.String(); v != "" {
			obj.TCPSocket.Port = &probe.TCPSocket.Port
		}
		if probe.TCPSocket.Host != "" {
			obj.TCPSocket.Host = &probe.TCPSocket.Host
		}
	}
	if probe.InitialDelaySeconds != 0 {
		obj.InitialDelaySeconds = util.Int64Ptr(int64(probe.InitialDelaySeconds))
	}
	if probe.TimeoutSeconds != 0 {
		obj.TimeoutSeconds = util.Int64Ptr(int64(probe.TimeoutSeconds))
	}
	if probe.PeriodSeconds != 0 {
		obj.PeriodSeconds = util.Int64Ptr(int64(probe.PeriodSeconds))
	}
	if probe.SuccessThreshold != 0 {
		obj.SuccessThreshold = util.Int64Ptr(int64(probe.SuccessThreshold))
	}
	if probe.FailureThreshold != 0 {
		obj.FailureThreshold = util.Int64Ptr(int64(probe.FailureThreshold))
	}
	obj.TerminationGracePeriodSeconds = probe.TerminationGracePeriodSeconds
	return obj
}

func convertPodSpecContainerHandler(handler *corev1.LifecycleHandler) *model.PodSpecContainerHandler {
	obj := &model.PodSpecContainerHandler{}
	if handler.Exec != nil {
		obj.Exec = &model.PodSpecContainerExecAction{Command: handler.Exec.Command}
	}
	if handler.HTTPGet != nil {
		obj.HTTPGet = &model.PodSpecContainerHTTPGetAction{}
		if handler.HTTPGet.Path != "" {
			obj.HTTPGet.Path = &handler.HTTPGet.Path
		}
		if v := handler.HTTPGet.Port.String(); v != "" {
			obj.HTTPGet.Port = &handler.HTTPGet.Port
		}
		if handler.HTTPGet.Host != "" {
			obj.HTTPGet.Host = &handler.HTTPGet.Host
		}
		if handler.HTTPGet.Scheme != "" {
			obj.HTTPGet.Scheme = util.StringPtr(string(handler.HTTPGet.Scheme))
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
		obj.TCPSocket = &model.PodSpecContainerTCPSocketAction{}
		if v := handler.TCPSocket.Port.String(); v != "" {
			obj.TCPSocket.Port = &handler.TCPSocket.Port
		}
		if handler.TCPSocket.Host != "" {
			obj.TCPSocket.Host = &handler.TCPSocket.Host
		}
	}
	return obj
}
