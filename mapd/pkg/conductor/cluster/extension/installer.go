// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Installer handles lifecycle of ClusterExtension.
type Installer interface {
	GetEndpoints() (endpoints []string, err error)
	Install() error
	Uninstall() error
}

type GenericInstallerConfig struct {
	*k8s.Kubernetes
	Extension           *model.ClusterExtension
	ImageRegistry       *model.ImageRegistry
	LoadBalancerEnabled bool

	Cluster            string
	ClusterClass       model.ClusterClass
	JobNamespace       string
	Job                string
	JobDeadlineSeconds int64
	ServiceNamespace   string
	Services           []string
	Image              string
	CACerts            []byte
	UnmarshalEnvFunc   func(ext *model.ClusterExtension, env map[string]string) error
}

// GenericInstaller provides helpers for implementing a installer. It should be
// inherited by all Installer implementation.
type GenericInstaller struct {
	GenericInstallerConfig

	log *zap.SugaredLogger
}

func (k *GenericInstaller) GetEndpoints(defaultScheme string, mappingPortScheme map[int32]string) (endpoints []string, err error) {
	if len(k.Services) < 1 {
		return nil, nil
	}

	nodeAddressess := make(map[string]net.IP)

	namespace := k.ServiceNamespace
	if namespace == "" {
		namespace = k.JobNamespace
	}
	for _, service := range k.Services {
		svc, err := getService(k.Kubernetes, namespace, service)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(svc.Spec.Ports, func(i, j int) bool { return svc.Spec.Ports[i].Protocol < svc.Spec.Ports[j].Protocol })
		sort.SliceStable(svc.Spec.Ports, func(i, j int) bool { return svc.Spec.Ports[i].Port < svc.Spec.Ports[j].Port })

		switch svc.Spec.Type {
		case corev1.ServiceTypeLoadBalancer:
			if len(svc.Status.LoadBalancer.Ingress) < 1 {
				return nil, fmt.Errorf("service not ready")
			}

			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				ip := ingress.IP
				sort.SliceStable(ingress.Ports, func(i, j int) bool { return ingress.Ports[i].Protocol < ingress.Ports[j].Protocol })
				sort.SliceStable(ingress.Ports, func(i, j int) bool { return ingress.Ports[i].Port < ingress.Ports[j].Port })

				ports := []int32{}
				for _, each := range ingress.Ports {
					ports = append(ports, each.Port)
				}
				if len(ports) < 1 {
					for _, each := range svc.Spec.Ports {
						ports = append(ports, each.Port)
					}
				}

				for _, port := range ports {
					scheme := defaultScheme
					if mappingPortScheme != nil {
						v, ok := mappingPortScheme[port]
						if ok && v != "" {
							scheme = v
						}
					}
					u := &url.URL{Scheme: scheme}
					switch scheme {
					case "http":
						switch port {
						case 0, 80:
							u.Host = ip
						}
					case "https":
						switch port {
						case 0, 443:
							u.Host = ip
						}
					}
					if u.Host == "" {
						u.Host = fmt.Sprintf("%s:%d", ip, port)
					}
					endpoints = append(endpoints, u.String())
				}
			}

		case corev1.ServiceTypeNodePort:
			if len(nodeAddressess) < 1 {
				nodeList, err := k.Kubernetes.GetNodes(k8s.LabelNotExists("node-role.kubernetes.io/control-plane"), k8s.LabelNotExists("node-role.kubernetes.io/master"))
				if err != nil {
					return endpoints, err
				}
				for _, node := range nodeList.Items {
					if node.Status.Phase != corev1.NodePending && node.Status.Phase != corev1.NodeTerminated {
						for _, addr := range node.Status.Addresses {
							switch addr.Type {
							case corev1.NodeExternalIP, corev1.NodeInternalIP:
								ip := net.ParseIP(addr.Address)
								if ip != nil {
									nodeAddressess[ip.String()] = ip
								}
							}
						}
					}
				}
			}

			for _, each := range svc.Spec.Ports {
				port := each.NodePort

				scheme := defaultScheme
				if mappingPortScheme != nil {
					v, ok := mappingPortScheme[each.Port]
					if ok && v != "" {
						scheme = v
					}
				}

				for ipStr := range nodeAddressess {
					u := &url.URL{Scheme: scheme}
					switch scheme {
					case "http":
						switch port {
						case 0, 80:
							u.Host = ipStr
						}
					case "https":
						switch port {
						case 0, 443:
							u.Host = ipStr
						}
					}
					if u.Host == "" {
						u.Host = fmt.Sprintf("%s:%d", ipStr, port)
					}
					endpoints = append(endpoints, u.String())
				}
			}
		}
	}
	return endpoints, nil
}

func (k *GenericInstaller) provision(ext *model.ClusterExtension, enable bool) error {
	defer k.log.Sync()
	jobDeadlineSeconds := int64(defJobDeadlineSecs)
	graceDeadlineSeconds := int64(graceJobDeadlineSecs)
	if k.JobDeadlineSeconds > 0 {
		jobDeadlineSeconds = k.JobDeadlineSeconds
		graceDeadlineSeconds = k.JobDeadlineSeconds + graceJobDeadlineSecs - defJobDeadlineSecs
	}

	if k.UnmarshalEnvFunc == nil {
		panic(fmt.Errorf("UnmarshalEnvFunc must not be nil"))
	}

	env := make(map[string]string)
	if err := k.UnmarshalEnvFunc(k.Extension, env); err != nil {
		k.log.Errorf("unexpected error: %v.", err)
		return err
	}
	env[SharedEnvVarEnableExtension] = strconv.FormatBool(enable)
	env[SharedEnvVarClusterName] = k.Cluster
	env[SharedEnvVarClusterClass] = k.ClusterClass.String()
	env[SharedEnvVarImageRegistry] = k.ImageRegistry.Endpoint
	env[SharedEnvVarDockerConfig] = defDockerCfgDir
	env[SharedEnvVarUseNodePortService] = strconv.FormatBool(!k.LoadBalancerEnabled)
	if k.CACerts != nil {
		env[SharedEnvVarRootCACertsBase64] = base64.StdEncoding.EncodeToString(k.CACerts)
	}

	job, err := getJob(k.Kubernetes, k.JobNamespace, k.Job)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}

		u, err := url.Parse(k.ImageRegistry.Endpoint)
		if err != nil {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}

		container := corev1.Container{
			Name:            "main",
			Image:           filepath.Join(u.Host, mapd.GlobalConfig.ClusterImageProject, k.Image),
			ImagePullPolicy: "Always",
			VolumeMounts: []corev1.VolumeMount{{
				MountPath: defDockerCfgDir,
				Name:      "docker-registry-secret",
				ReadOnly:  true,
			}},
		}
		for k, v := range env {
			container.Env = append(container.Env, corev1.EnvVar{Name: k, Value: v})
		}
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      k.Job,
				Namespace: k.JobNamespace,
				Labels: map[string]string{
					defLabelInternal: "true",
					defLabelManaged:  "true",
					defLabelIssuer:   "cluster",
				},
			},
			Spec: batchv1.JobSpec{
				BackoffLimit:            util.Int32Ptr(2),
				Completions:             util.Int32Ptr(1),
				Parallelism:             util.Int32Ptr(1),
				ActiveDeadlineSeconds:   util.Int64Ptr(jobDeadlineSeconds),
				TTLSecondsAfterFinished: util.Int32Ptr(3600),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							defLabelInternal: "true",
							defLabelManaged:  "true",
							defLabelIssuer:   "cluster",
						},
					},
					Spec: corev1.PodSpec{
						Containers:         []corev1.Container{container},
						ImagePullSecrets:   []corev1.LocalObjectReference{{Name: defRegistrySecret}},
						RestartPolicy:      "Never",
						ServiceAccountName: defServiceAccount,
						Volumes: []corev1.Volume{{
							Name: "docker-registry-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									DefaultMode: util.Int32Ptr(420),
									Items: []corev1.KeyToPath{{
										Key:  ".dockerconfigjson",
										Path: "config.json",
									}},
									SecretName: defRegistrySecret,
								},
							},
						}},
					},
				},
			},
		}
		k.log.Infof("job %s/%s does not exist, now launch on cluster %s...", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		if _, err := postJob(k.Kubernetes, job); err != nil {
			k.log.Errorf("failed to provision job %s/%s to cluster %s due to: %v.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster), err)
			return err
		}
		k.log.Infof("job %s/%s is provisioning on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		return fmt.Errorf("job is still running")
	}

	oldEnv := make(map[string]string)
	recreate := job.Spec.ActiveDeadlineSeconds == nil || *job.Spec.ActiveDeadlineSeconds != jobDeadlineSeconds
LOOP:
	for _, c := range job.Spec.Template.Spec.Containers {
		for _, e := range c.Env {
			v, ok := env[e.Name]
			if !ok || v != e.Value {
				recreate = true
				break LOOP
			}
			oldEnv[e.Name] = e.Value
		}
	}
	if !recreate {
		for k := range env {
			if _, ok := oldEnv[k]; !ok {
				recreate = true
				break
			}
		}
	}

	for _, each := range job.Status.Conditions {
		if each.Type == "Failed" {
			if each.Reason == "DeadlineExceeded" {
				k.log.Errorf("job %s/%s timed out for some reason. For details, please check events on cluster %s manually.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
				return &ErrCriticalFailure{fmt.Errorf("job is failed")}
			}
			break
		}
	}
	if job.Status.Active > 0 {
		if time.Since(job.Status.StartTime.Time) > time.Duration(graceDeadlineSeconds)*time.Second {
			k.log.Errorf("job %s/%s timed out for some reason. For details, please check pod logs on cluster %s manually.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
			return &ErrCriticalFailure{fmt.Errorf("job is failed")}
		}
		if recreate {
			k.log.Warnf("job %s/%s requires recreation but it is still running. Pending update on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		}
		k.log.Infof("job %s/%s is still running on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		return fmt.Errorf("job is still running")
	}
	if job.Status.Failed > 0 && job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit {
		if recreate {
			goto RECREATE
		}

		k.log.Errorf("job %s/%s failed for some reason. For details, please check pod logs on cluster %s manually.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		return &ErrCriticalFailure{fmt.Errorf("job is failed")}
	}
	if job.Status.Succeeded < 1 {
		// we wait for another 10 seconds to make sure the job is failed
		if time.Since(job.CreationTimestamp.Time) < time.Second*time.Duration(graceDeadlineSeconds) {
			if recreate {
				k.log.Warnf("job %s/%s requires recreation but it is still running. Pending update on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
			}
			k.log.Infof("job %s/%s is still running on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
			return fmt.Errorf("job is still running")
		}

		if recreate {
			goto RECREATE
		}

		k.log.Errorf("job %s/%s is neither succeeded nor running, and we assume this job has failed. For details, please check pod logs on cluster %s manually.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
		return &ErrCriticalFailure{fmt.Errorf("job is failed")}
	}
	if recreate {
		goto RECREATE
	}
	return nil

RECREATE:
	k.log.Infof("job %s/%s requires recreation on cluster %s.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster))
	if err := deleteJob(k.Kubernetes, k.JobNamespace, k.Job); err != nil {
		k.log.Errorf("error removing previous job %s/%s on cluster %s: %v.", k.JobNamespace, k.Job, strconv.Quote(k.Cluster), err)
		return fmt.Errorf("job cannot be removed")
	}
	return fmt.Errorf("job is pending creation")
}

func (k *GenericInstaller) Install() error {
	return k.provision(k.Extension, true)
}

func (k *GenericInstaller) Uninstall() error {
	return k.provision(k.Extension, false)
}

// NewGenericInstaller creates GenericInstaller from given cfg and log.
func NewGenericInstaller(cfg GenericInstallerConfig, log *zap.SugaredLogger) *GenericInstaller {
	return &GenericInstaller{
		GenericInstallerConfig: cfg,
		log:                    log,
	}
}

func unmarshalEnvFromExtension(ext *model.ClusterExtension, env map[string]string) error {
	spec := ext.Spec
	if len(spec) == 0 {
		spec = []byte("{}")
	}
	if err := json.Unmarshal(spec, &env); err != nil {
		return err
	}
	return nil
}
