// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"context"
	"fmt"
	"sync"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	SharedEnvVarEnableExtension    = "MAPD_ENABLE_EXTENSION"
	SharedEnvVarClusterName        = "MAPD_CLUSTER_NAME"
	SharedEnvVarClusterClass       = "MAPD_CLUSTER_CLASS"
	SharedEnvVarImageRegistry      = "MAPD_IMAGE_REGISTRY"
	SharedEnvVarRootCACertsBase64  = "MAPD_ROOTCA_CERTS_BASE64"
	SharedEnvVarUseNodePortService = "MAPD_USE_NODEPORT_SERVICE"
	SharedEnvVarDockerConfig       = "DOCKER_CONFIG"

	defRegistrySecret   = "vmware-system-map-registry"
	defKubeConfigSecret = "vmware-system-map-kubeconfig"
	defServiceAccount   = "vmware-system-map"

	defLabelManaged  = "map.tanzu.vmware.com/managed"
	defLabelInternal = "map.tanzu.vmware.com/internal"
	defLabelIssuer   = "map.tanzu.vmware.com/issuer"

	defJobDeadlineSecs   = 900
	graceJobDeadlineSecs = 1200
	defDockerCfgDir      = "/data/.docker"
)

type ErrCriticalFailure struct {
	Err error
}

func (e *ErrCriticalFailure) Error() string {
	return e.Err.Error()
}

func IsCriticalError(err error) bool {
	_, ok := err.(*ErrCriticalFailure)
	return ok
}

func postJob(client *k8s.Kubernetes, job *batchv1.Job) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.CreateJobContext(ctx, job)
}

func getJob(client *k8s.Kubernetes, namespace, name string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.GetJobContext(ctx, namespace, name)
}

func deleteJob(client *k8s.Kubernetes, namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	err := client.DeleteJobContext(ctx, namespace, name)
	if err != nil && k8s.IsErrObjectNotExists(err) {
		return nil
	}
	return err
}

func getService(client *k8s.Kubernetes, namespace, name string) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return client.GetServiceContext(ctx, namespace, name)
}

type InstallerOptions struct {
	ClusterName         string
	ClusterClass        model.ClusterClass
	Kubernetes          *k8s.Kubernetes
	ImageRegistry       *model.ImageRegistry
	LoadBalancerEnabled bool
	CACerts             []byte
	Namespace           string
	Extension           *model.ClusterExtension
	Logger              *zap.SugaredLogger
}

type NewFunc func(*InstallerOptions) Installer

type registry struct {
	mux sync.RWMutex
	fn  map[model.ClusterExtensionName]NewFunc
}

func (r *registry) Register(name model.ClusterExtensionName, newFn NewFunc) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.fn[name] = newFn
}

func (r *registry) NewInstaller(opts *InstallerOptions) (Installer, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	newFn, ok := r.fn[opts.Extension.Name]
	if !ok {
		return nil, fmt.Errorf("unknown extension name: %s", opts.Extension.Name)
	}
	return newFn(opts), nil
}

func newRegistry() *registry {
	return &registry{fn: make(map[model.ClusterExtensionName]NewFunc)}
}

var Registry = newRegistry()
