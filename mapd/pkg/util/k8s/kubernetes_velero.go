// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	scheme "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned/scheme"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type VeleroKubernetes struct {
	*Kubernetes
}

func (w VeleroKubernetes) CreateBackupStorageLocation(obj *velerov1.BackupStorageLocation) (*velerov1.BackupStorageLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateBackupStorageLocationContext(ctx, obj)
}

func (w VeleroKubernetes) CreateBackupStorageLocationContext(ctx context.Context, obj *velerov1.BackupStorageLocation) (*velerov1.BackupStorageLocation, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.BackupStorageLocation{}
	err = clientset.VeleroV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("backupstoragelocations").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetBackupStorageLocation(namespace, name string) (*velerov1.BackupStorageLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBackupStorageLocationContext(ctx, namespace, name)
}

func (w VeleroKubernetes) GetBackupStorageLocationContext(ctx context.Context, namespace, name string) (*velerov1.BackupStorageLocation, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.BackupStorageLocation{}
	err = clientset.VeleroV1().RESTClient().Get().
		Namespace(namespace).
		Resource("backupstoragelocations").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetBackupStorageLocations(namespace string, selectors ...Selector) (*velerov1.BackupStorageLocationList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBackupStorageLocationsContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) GetBackupStorageLocationsContext(ctx context.Context, namespace string, selectors ...Selector) (*velerov1.BackupStorageLocationList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &velerov1.BackupStorageLocationList{}
	err = clientset.VeleroV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("backupstoragelocations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) DeleteBackupStorageLocation(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteBackupStorageLocationContext(ctx, namespace, name)
}

func (w VeleroKubernetes) DeleteBackupStorageLocationContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("backupstoragelocations").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) CreateBackup(obj *velerov1.Backup) (*velerov1.Backup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateBackupContext(ctx, obj)
}

func (w VeleroKubernetes) CreateBackupContext(ctx context.Context, obj *velerov1.Backup) (*velerov1.Backup, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Backup{}
	err = clientset.VeleroV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("backups").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetBackup(namespace, name string) (*velerov1.Backup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBackupContext(ctx, namespace, name)
}

func (w VeleroKubernetes) GetBackupContext(ctx context.Context, namespace, name string) (*velerov1.Backup, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Backup{}
	err = clientset.VeleroV1().RESTClient().Get().
		Namespace(namespace).
		Resource("backups").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetBackups(namespace string, selectors ...Selector) (*velerov1.BackupList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBackupsContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) GetBackupsContext(ctx context.Context, namespace string, selectors ...Selector) (*velerov1.BackupList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &velerov1.BackupList{}
	err = clientset.VeleroV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("backups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) DeleteBackup(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteBackupContext(ctx, namespace, name)
}

func (w VeleroKubernetes) DeleteBackupContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("backups").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) DeleteBackups(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteBackupsContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) DeleteBackupsContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("backups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) CreateSchedule(obj *velerov1.Schedule) (*velerov1.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateScheduleContext(ctx, obj)
}

func (w VeleroKubernetes) CreateScheduleContext(ctx context.Context, obj *velerov1.Schedule) (*velerov1.Schedule, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Schedule{}
	err = clientset.VeleroV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("schedules").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) UpdateSchedule(obj *velerov1.Schedule) (*velerov1.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateScheduleContext(ctx, obj)
}

func (w VeleroKubernetes) UpdateScheduleContext(ctx context.Context, obj *velerov1.Schedule) (*velerov1.Schedule, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Schedule{}
	err = clientset.VeleroV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("schedules").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetSchedule(namespace, name string) (*velerov1.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetScheduleContext(ctx, namespace, name)
}

func (w VeleroKubernetes) GetScheduleContext(ctx context.Context, namespace, name string) (*velerov1.Schedule, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Schedule{}
	err = clientset.VeleroV1().RESTClient().Get().
		Namespace(namespace).
		Resource("schedules").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetSchedules(namespace string, selectors ...Selector) (*velerov1.ScheduleList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetSchedulesContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) GetSchedulesContext(ctx context.Context, namespace string, selectors ...Selector) (*velerov1.ScheduleList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &velerov1.ScheduleList{}
	err = clientset.VeleroV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("schedules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) DeleteSchedule(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteScheduleContext(ctx, namespace, name)
}

func (w VeleroKubernetes) DeleteScheduleContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("schedules").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) DeleteSchedules(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteSchedulesContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) DeleteSchedulesContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("schedules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) CreateDeleteBackupRequest(obj *velerov1.DeleteBackupRequest) (*velerov1.DeleteBackupRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateDeleteBackupRequestContext(ctx, obj)
}

func (w VeleroKubernetes) CreateDeleteBackupRequestContext(ctx context.Context, obj *velerov1.DeleteBackupRequest) (*velerov1.DeleteBackupRequest, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.DeleteBackupRequest{}
	err = clientset.VeleroV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("deletebackuprequests").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetDeleteBackupRequest(namespace, name string) (*velerov1.DeleteBackupRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDeleteBackupRequestContext(ctx, namespace, name)
}

func (w VeleroKubernetes) GetDeleteBackupRequestContext(ctx context.Context, namespace, name string) (*velerov1.DeleteBackupRequest, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.DeleteBackupRequest{}
	err = clientset.VeleroV1().RESTClient().Get().
		Namespace(namespace).
		Resource("deletebackuprequests").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) CreateRestore(obj *velerov1.Restore) (*velerov1.Restore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateRestoreContext(ctx, obj)
}

func (w VeleroKubernetes) CreateRestoreContext(ctx context.Context, obj *velerov1.Restore) (*velerov1.Restore, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Restore{}
	err = clientset.VeleroV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("restores").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetRestore(namespace, name string) (*velerov1.Restore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetRestoreContext(ctx, namespace, name)
}

func (w VeleroKubernetes) GetRestoreContext(ctx context.Context, namespace, name string) (*velerov1.Restore, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &velerov1.Restore{}
	err = clientset.VeleroV1().RESTClient().Get().
		Namespace(namespace).
		Resource("restores").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) GetRestores(namespace string, selectors ...Selector) (*velerov1.RestoreList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetRestoresContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) GetRestoresContext(ctx context.Context, namespace string, selectors ...Selector) (*velerov1.RestoreList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &velerov1.RestoreList{}
	err = clientset.VeleroV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("restores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w VeleroKubernetes) DeleteRestore(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteRestoreContext(ctx, namespace, name)
}

func (w VeleroKubernetes) DeleteRestoreContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("restores").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w VeleroKubernetes) DeleteRestores(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteRestoresContext(ctx, namespace, selectors...)
}

func (w VeleroKubernetes) DeleteRestoresContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.VeleroV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("restores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func NewVelero(w *Kubernetes) *VeleroKubernetes {
	return &VeleroKubernetes{w}
}

func NewVeleroFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*VeleroKubernetes, error) {
	w, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewVelero(w), nil
}
