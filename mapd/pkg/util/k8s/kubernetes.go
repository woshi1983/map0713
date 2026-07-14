// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Kubernetes struct {
	Config *rest.Config
}

func (w *Kubernetes) Version() (*version.Info, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.VersionContext(ctx)
}

func (w *Kubernetes) VersionContext(ctx context.Context) (*version.Info, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	body, err := clientset.Discovery().RESTClient().Get().
		AbsPath("/version").
		Do(ctx).
		Raw()
	if err != nil {
		return nil, err
	}
	var info version.Info
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the server version: %v", err)
	}
	return &info, nil
}

func (w *Kubernetes) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PingContext(ctx)
}

func (w *Kubernetes) PingContext(ctx context.Context) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	return clientset.Discovery().RESTClient().Get().
		AbsPath("/healthz").
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateNamespace(obj *corev1.Namespace) (*corev1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateNamespaceContext(ctx, obj)
}

func (w *Kubernetes) CreateNamespaceContext(ctx context.Context, obj *corev1.Namespace) (*corev1.Namespace, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Namespace{}
	err = clientset.CoreV1().RESTClient().Post().
		Resource("namespaces").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateNamespace(obj *corev1.Namespace) (*corev1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateNamespaceContext(ctx, obj)
}

func (w *Kubernetes) UpdateNamespaceContext(ctx context.Context, obj *corev1.Namespace) (*corev1.Namespace, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Namespace{}
	err = clientset.CoreV1().RESTClient().Put().
		Resource("namespaces").
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

func (w *Kubernetes) GetNamespace(name string) (*corev1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNamespaceContext(ctx, name)
}

func (w *Kubernetes) GetNamespaceContext(ctx context.Context, name string) (*corev1.Namespace, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Namespace{}
	err = clientset.CoreV1().RESTClient().Get().
		Resource("namespaces").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetNamespaces(selectors ...Selector) (*corev1.NamespaceList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNamespacesContext(ctx, selectors...)
}

func (w *Kubernetes) GetNamespacesContext(ctx context.Context, selectors ...Selector) (*corev1.NamespaceList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.NamespaceList{}
	err = clientset.CoreV1().RESTClient().Get().
		Resource("namespaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteNamespace(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteNamespaceContext(ctx, name)
}

func (w *Kubernetes) DeleteNamespaceContext(ctx context.Context, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Resource("namespaces").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateHorizontalPodAutoscaler(obj *autoscalingv2beta2.HorizontalPodAutoscaler) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateHorizontalPodAutoscalerContext(ctx, obj)
}

func (w *Kubernetes) CreateHorizontalPodAutoscalerContext(ctx context.Context, obj *autoscalingv2beta2.HorizontalPodAutoscaler) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err = clientset.AutoscalingV2beta2().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("horizontalpodautoscalers").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateHorizontalPodAutoscaler(obj *autoscalingv2beta2.HorizontalPodAutoscaler) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateHorizontalPodAutoscalerContext(ctx, obj)
}

func (w *Kubernetes) UpdateHorizontalPodAutoscalerContext(ctx context.Context, obj *autoscalingv2beta2.HorizontalPodAutoscaler) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err = clientset.AutoscalingV2beta2().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("horizontalpodautoscalers").
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

func (w *Kubernetes) GetHorizontalPodAutoscaler(namespace, name string) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetHorizontalPodAutoscalerContext(ctx, namespace, name)
}

func (w *Kubernetes) GetHorizontalPodAutoscalerContext(ctx context.Context, namespace, name string) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	err = clientset.AutoscalingV2beta2().RESTClient().Get().
		Namespace(namespace).
		Resource("horizontalpodautoscalers").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetHorizontalPodAutoscalers(namespace string, selectors ...Selector) (*autoscalingv2beta2.HorizontalPodAutoscalerList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetHorizontalPodAutoscalersContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetHorizontalPodAutoscalersContext(ctx context.Context, namespace string, selectors ...Selector) (*autoscalingv2beta2.HorizontalPodAutoscalerList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &autoscalingv2beta2.HorizontalPodAutoscalerList{}
	err = clientset.AutoscalingV2beta2().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("secrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteHorizontalPodAutoscaler(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteHorizontalPodAutoscalerContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteHorizontalPodAutoscalerContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.AutoscalingV2beta2().RESTClient().Delete().
		Namespace(namespace).
		Resource("horizontalpodautoscalers").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) GetResourceQuotas(namespace string, selectors ...Selector) (*corev1.ResourceQuotaList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetResourceQuotasContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetResourceQuotasContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.ResourceQuotaList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.ResourceQuotaList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("resourcequotas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) CreateSecret(obj *corev1.Secret) (*corev1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateSecretContext(ctx, obj)
}

func (w *Kubernetes) CreateSecretContext(ctx context.Context, obj *corev1.Secret) (*corev1.Secret, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Secret{}
	err = clientset.CoreV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("secrets").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateSecret(obj *corev1.Secret) (*corev1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateSecretContext(ctx, obj)
}

func (w *Kubernetes) UpdateSecretContext(ctx context.Context, obj *corev1.Secret) (*corev1.Secret, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Secret{}
	err = clientset.CoreV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("secrets").
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

func (w *Kubernetes) GetSecret(namespace, name string) (*corev1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetSecretContext(ctx, namespace, name)
}

func (w *Kubernetes) GetSecretContext(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Secret{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("secrets").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetSecrets(namespace string, selectors ...Selector) (*corev1.SecretList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetSecretsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetSecretsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.SecretList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.SecretList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("secrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteSecret(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteSecretContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteSecretContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("secrets").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateConfigMap(obj *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateConfigMapContext(ctx, obj)
}

func (w *Kubernetes) CreateConfigMapContext(ctx context.Context, obj *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ConfigMap{}
	err = clientset.CoreV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("configmaps").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateConfigMap(obj *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateConfigMapContext(ctx, obj)
}

func (w *Kubernetes) UpdateConfigMapContext(ctx context.Context, obj *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ConfigMap{}
	err = clientset.CoreV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("configmaps").
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

func (w *Kubernetes) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetConfigMapContext(ctx, namespace, name)
}

func (w *Kubernetes) GetConfigMapContext(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ConfigMap{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("configmaps").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetConfigMaps(namespace string, selectors ...Selector) (*corev1.ConfigMapList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetConfigMapsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetConfigMapsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.ConfigMapList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.ConfigMapList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("configmaps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteConfigMap(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteConfigMapContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteConfigMapContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("configmaps").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreatePersistentVolumeClaim(obj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreatePersistentVolumeClaimContext(ctx, obj)
}

func (w *Kubernetes) CreatePersistentVolumeClaimContext(ctx context.Context, obj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.PersistentVolumeClaim{}
	err = clientset.CoreV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("persistentvolumeclaims").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdatePersistentVolumeClaim(obj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdatePersistentVolumeClaimContext(ctx, obj)
}

func (w *Kubernetes) UpdatePersistentVolumeClaimContext(ctx context.Context, obj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.PersistentVolumeClaim{}
	err = clientset.CoreV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("persistentvolumeclaims").
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

func (w *Kubernetes) GetPersistentVolumeClaim(namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPersistentVolumeClaimContext(ctx, namespace, name)
}

func (w *Kubernetes) GetPersistentVolumeClaimContext(ctx context.Context, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.PersistentVolumeClaim{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("persistentvolumeclaims").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetPersistentVolumeClaims(namespace string, selectors ...Selector) (*corev1.PersistentVolumeClaimList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPersistentVolumeClaimsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetPersistentVolumeClaimsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.PersistentVolumeClaimList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.PersistentVolumeClaimList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("persistentvolumeclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeletePersistentVolumeClaim(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeletePersistentVolumeClaimContext(ctx, namespace, name)
}

func (w *Kubernetes) DeletePersistentVolumeClaimContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("persistentvolumeclaims").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateStorageClass(obj *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateStorageClassContext(ctx, obj)
}

func (w *Kubernetes) CreateStorageClassContext(ctx context.Context, obj *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &storagev1.StorageClass{}
	err = clientset.StorageV1().RESTClient().Post().
		Resource("storageclasses").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateStorageClass(obj *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateStorageClassContext(ctx, obj)
}

func (w *Kubernetes) UpdateStorageClassContext(ctx context.Context, obj *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &storagev1.StorageClass{}
	err = clientset.StorageV1().RESTClient().Put().
		Resource("storageclasses").
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

func (w *Kubernetes) GetStorageClass(name string) (*storagev1.StorageClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetStorageClassContext(ctx, name)
}

func (w *Kubernetes) GetStorageClassContext(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &storagev1.StorageClass{}
	err = clientset.StorageV1().RESTClient().Get().
		Resource("storageclasses").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetStorageClasses(selectors ...Selector) (*storagev1.StorageClassList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetStorageClassesContext(ctx, selectors...)
}

func (w *Kubernetes) GetStorageClassesContext(ctx context.Context, selectors ...Selector) (*storagev1.StorageClassList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &storagev1.StorageClassList{}
	err = clientset.StorageV1().RESTClient().Get().
		Resource("storageclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteStorageClass(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteStorageClassContext(ctx, name)
}

func (w *Kubernetes) DeleteStorageClassContext(ctx context.Context, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.StorageV1().RESTClient().Delete().
		Resource("storageclasses").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CordonNode(name string) (*corev1.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CordonNodeContext(ctx, name)
}

func (w *Kubernetes) CordonNodeContext(ctx context.Context, name string) (*corev1.Node, error) {
	return w.PatchNodeContext(ctx, name, types.StrategicMergePatchType, []byte(`{"spec":{"unschedulable":true}}`))
}

func (w *Kubernetes) UncordonNode(name string) (*corev1.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UncordonNodeContext(ctx, name)
}

func (w *Kubernetes) UncordonNodeContext(ctx context.Context, name string) (*corev1.Node, error) {
	return w.PatchNodeContext(ctx, name, types.StrategicMergePatchType, []byte(`{"spec":{"unschedulable":null}}`))
}

func (w *Kubernetes) PatchNode(name string, pt types.PatchType, patch []byte) (*corev1.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PatchNodeContext(ctx, name, pt, patch)
}

func (w *Kubernetes) PatchNodeContext(ctx context.Context, name string, pt types.PatchType, patch []byte) (*corev1.Node, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Node{}
	err = clientset.CoreV1().RESTClient().Patch(pt).
		Resource("nodes").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetNode(name string) (*corev1.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNodeContext(ctx, name)
}

func (w *Kubernetes) GetNodeContext(ctx context.Context, name string) (*corev1.Node, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Node{}
	err = clientset.CoreV1().RESTClient().Get().
		Resource("nodes").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetNodes(selectors ...Selector) (*corev1.NodeList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNodesContext(ctx, selectors...)
}

func (w *Kubernetes) GetNodesContext(ctx context.Context, selectors ...Selector) (*corev1.NodeList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.NodeList{}
	err = clientset.CoreV1().RESTClient().Get().
		Resource("nodes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetPod(namespace, name string) (*corev1.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPodContext(ctx, namespace, name)
}

func (w *Kubernetes) GetPodContext(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Pod{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetPodLog(namespace, name, container string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPodLogContext(ctx, namespace, name, container)
}

func (w *Kubernetes) GetPodLogContext(ctx context.Context, namespace, name, container string) ([]byte, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		SubResource("log").
		VersionedParams(&corev1.PodLogOptions{Container: container, TailLines: util.Int64Ptr(1000)}, scheme.ParameterCodec).
		Do(ctx).
		Raw()
}

func (w *Kubernetes) WatchPodLog(ctx context.Context, namespace, name, container string) (io.ReadCloser, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		SubResource("log").
		VersionedParams(&corev1.PodLogOptions{Container: container, Follow: true, TailLines: util.Int64Ptr(1000)}, scheme.ParameterCodec).
		Stream(ctx)
}

func (w *Kubernetes) WatchPodExec(ctx context.Context, namespace, name, container string, cmd ...string) (stdin io.Writer, out io.Reader, done <-chan struct{}, err error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, nil, nil, err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{Container: container, Stdin: true, Stdout: true, TTY: true, Command: cmd}, scheme.ParameterCodec)

	// use combined output, as the stderr will be sent to stdout if tty is enabled.
	stdinr, stdin := io.Pipe()
	out, outw := io.Pipe()
	exec, err := NewSPDYExecutor(w.Config, "POST", req.URL())
	if err != nil {
		return nil, nil, nil, err
	}
	done, err = exec.Stream(ctx, remotecommand.StreamOptions{
		Stdin:  stdinr,
		Stdout: outw,
		Stderr: outw,
		Tty:    true,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return stdin, out, done, nil
}

func (w *Kubernetes) GetPods(namespace string, selectors ...Selector) (*corev1.PodList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPodsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetPodsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.PodList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.PodList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeletePod(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeletePodContext(ctx, namespace, name)
}

func (w *Kubernetes) DeletePodContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) DeletePods(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeletePodsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) DeletePodsContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateService(obj *corev1.Service) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateServiceContext(ctx, obj)
}

func (w *Kubernetes) CreateServiceContext(ctx context.Context, obj *corev1.Service) (*corev1.Service, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Service{}
	err = clientset.CoreV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("services").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateService(obj *corev1.Service) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateServiceContext(ctx, obj)
}

func (w *Kubernetes) UpdateServiceContext(ctx context.Context, obj *corev1.Service) (*corev1.Service, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Service{}
	err = clientset.CoreV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("services").
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

func (w *Kubernetes) GetService(namespace, name string) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetServiceContext(ctx, namespace, name)
}

func (w *Kubernetes) GetServiceContext(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.Service{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("services").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetServices(namespace string, selectors ...Selector) (*corev1.ServiceList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetServicesContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetServicesContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.ServiceList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.ServiceList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteService(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteServiceContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteServiceContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("services").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) DeleteServices(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteServicesContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) DeleteServicesContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateIngress(obj *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateIngressContext(ctx, obj)
}

func (w *Kubernetes) CreateIngressContext(ctx context.Context, obj *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &networkingv1.Ingress{}
	err = clientset.NetworkingV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("ingresses").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateIngress(obj *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateIngressContext(ctx, obj)
}

func (w *Kubernetes) UpdateIngressContext(ctx context.Context, obj *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &networkingv1.Ingress{}
	err = clientset.NetworkingV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("ingresses").
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

func (w *Kubernetes) GetIngress(namespace, name string) (*networkingv1.Ingress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetIngressContext(ctx, namespace, name)
}

func (w *Kubernetes) GetIngressContext(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &networkingv1.Ingress{}
	err = clientset.NetworkingV1().RESTClient().Get().
		Namespace(namespace).
		Resource("ingresses").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetIngresses(namespace string, selectors ...Selector) (*networkingv1.IngressList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetIngressesContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetIngressesContext(ctx context.Context, namespace string, selectors ...Selector) (*networkingv1.IngressList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &networkingv1.IngressList{}
	err = clientset.NetworkingV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("ingresses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteIngress(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteIngressContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteIngressContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.NetworkingV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("ingresses").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) GetIngressClass(name string) (*networkingv1.IngressClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetIngressClassContext(ctx, name)
}

func (w *Kubernetes) GetIngressClassContext(ctx context.Context, name string) (*networkingv1.IngressClass, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &networkingv1.IngressClass{}
	err = clientset.NetworkingV1().RESTClient().Get().
		Resource("ingressclasses").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetIngressClasses(selectors ...Selector) (*networkingv1.IngressClassList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetIngressClassesContext(ctx, selectors...)
}

func (w *Kubernetes) GetIngressClassesContext(ctx context.Context, selectors ...Selector) (*networkingv1.IngressClassList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &networkingv1.IngressClassList{}
	err = clientset.NetworkingV1().RESTClient().Get().
		Resource("ingressclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) CreateServiceAccount(obj *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateServiceAccountContext(ctx, obj)
}

func (w *Kubernetes) CreateServiceAccountContext(ctx context.Context, obj *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ServiceAccount{}
	err = clientset.CoreV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("serviceaccounts").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateServiceAccount(obj *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateServiceAccountContext(ctx, obj)
}

func (w *Kubernetes) UpdateServiceAccountContext(ctx context.Context, obj *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ServiceAccount{}
	err = clientset.CoreV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("serviceaccounts").
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

func (w *Kubernetes) GetServiceAccount(namespace, name string) (*corev1.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetServiceAccountContext(ctx, namespace, name)
}

func (w *Kubernetes) GetServiceAccountContext(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &corev1.ServiceAccount{}
	err = clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("serviceaccounts").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetServiceAccounts(namespace string, selectors ...Selector) (*corev1.ServiceAccountList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetServiceAccountsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetServiceAccountsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.ServiceAccountList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &corev1.ServiceAccountList{}
	err = clientset.CoreV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("serviceaccounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteServiceAccount(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteServiceAccountContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteServiceAccountContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.CoreV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("serviceaccounts").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateClusterRoleBinding(obj *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateClusterRoleBindingContext(ctx, obj)
}

func (w *Kubernetes) CreateClusterRoleBindingContext(ctx context.Context, obj *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &rbacv1.ClusterRoleBinding{}
	err = clientset.RbacV1().RESTClient().Post().
		Resource("clusterrolebindings").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateClusterRoleBinding(obj *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateClusterRoleBindingContext(ctx, obj)
}

func (w *Kubernetes) UpdateClusterRoleBindingContext(ctx context.Context, obj *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &rbacv1.ClusterRoleBinding{}
	err = clientset.RbacV1().RESTClient().Put().
		Resource("clusterrolebindings").
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

func (w *Kubernetes) GetClusterRoleBinding(name string) (*rbacv1.ClusterRoleBinding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetClusterRoleBindingContext(ctx, name)
}

func (w *Kubernetes) GetClusterRoleBindingContext(ctx context.Context, name string) (*rbacv1.ClusterRoleBinding, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &rbacv1.ClusterRoleBinding{}
	err = clientset.RbacV1().RESTClient().Get().
		Resource("clusterrolebindings").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteClusterRoleBinding(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteClusterRoleBindingContext(ctx, name)
}

func (w *Kubernetes) DeleteClusterRoleBindingContext(ctx context.Context, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.RbacV1().RESTClient().Delete().
		Resource("clusterrolebindings").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateDaemonSet(obj *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateDaemonSetContext(ctx, obj)
}

func (w *Kubernetes) CreateDaemonSetContext(ctx context.Context, obj *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.DaemonSet{}
	err = clientset.AppsV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("daemonsets").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateDaemonSet(obj *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateDaemonSetContext(ctx, obj)
}

func (w *Kubernetes) UpdateDaemonSetContext(ctx context.Context, obj *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.DaemonSet{}
	err = clientset.AppsV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("daemonsets").
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

func (w *Kubernetes) GetDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDaemonSetContext(ctx, namespace, name)
}

func (w *Kubernetes) GetDaemonSetContext(ctx context.Context, namespace, name string) (*appsv1.DaemonSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.DaemonSet{}
	err = clientset.AppsV1().RESTClient().Get().
		Namespace(namespace).
		Resource("daemonsets").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetDaemonSets(namespace string, selectors ...Selector) (*appsv1.DaemonSetList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDaemonSetsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetDaemonSetsContext(ctx context.Context, namespace string, selectors ...Selector) (*appsv1.DaemonSetList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &appsv1.DaemonSetList{}
	err = clientset.AppsV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("daemonsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteDaemonSet(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteDaemonSetContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteDaemonSetContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("daemonsets").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateStatefulSet(obj *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateStatefulSetContext(ctx, obj)
}

func (w *Kubernetes) CreateStatefulSetContext(ctx context.Context, obj *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.StatefulSet{}
	err = clientset.AppsV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("statefulsets").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateStatefulSet(obj *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateStatefulSetContext(ctx, obj)
}

func (w *Kubernetes) UpdateStatefulSetContext(ctx context.Context, obj *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.StatefulSet{}
	err = clientset.AppsV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("statefulsets").
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

func (w *Kubernetes) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetStatefulSetContext(ctx, namespace, name)
}

func (w *Kubernetes) GetStatefulSetContext(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.StatefulSet{}
	err = clientset.AppsV1().RESTClient().Get().
		Namespace(namespace).
		Resource("statefulsets").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetStatefulSets(namespace string, selectors ...Selector) (*appsv1.StatefulSetList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetStatefulSetsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetStatefulSetsContext(ctx context.Context, namespace string, selectors ...Selector) (*appsv1.StatefulSetList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &appsv1.StatefulSetList{}
	err = clientset.AppsV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("statefulsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteStatefulSet(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteStatefulSetContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteStatefulSetContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("statefulsets").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateReplicaSet(obj *appsv1.ReplicaSet) (*appsv1.ReplicaSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateReplicaSetContext(ctx, obj)
}

func (w *Kubernetes) CreateReplicaSetContext(ctx context.Context, obj *appsv1.ReplicaSet) (*appsv1.ReplicaSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.ReplicaSet{}
	err = clientset.AppsV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("replicasets").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateReplicaSet(obj *appsv1.ReplicaSet) (*appsv1.ReplicaSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateReplicaSetContext(ctx, obj)
}

func (w *Kubernetes) UpdateReplicaSetContext(ctx context.Context, obj *appsv1.ReplicaSet) (*appsv1.ReplicaSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.ReplicaSet{}
	err = clientset.AppsV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("replicasets").
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

func (w *Kubernetes) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetReplicaSetContext(ctx, namespace, name)
}

func (w *Kubernetes) GetReplicaSetContext(ctx context.Context, namespace, name string) (*appsv1.ReplicaSet, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.ReplicaSet{}
	err = clientset.AppsV1().RESTClient().Get().
		Namespace(namespace).
		Resource("replicasets").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetReplicaSets(namespace string, selectors ...Selector) (*appsv1.ReplicaSetList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetReplicaSetsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetReplicaSetsContext(ctx context.Context, namespace string, selectors ...Selector) (*appsv1.ReplicaSetList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &appsv1.ReplicaSetList{}
	err = clientset.AppsV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("replicasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteReplicaSet(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteReplicaSetContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteReplicaSetContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("replicasets").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) DeleteReplicaSets(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteReplicaSetsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) DeleteReplicaSetsContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("replicasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateDeployment(obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateDeploymentContext(ctx, obj)
}

func (w *Kubernetes) CreateDeploymentContext(ctx context.Context, obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.Deployment{}
	err = clientset.AppsV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("deployments").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) UpdateDeployment(obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateDeploymentContext(ctx, obj)
}

func (w *Kubernetes) UpdateDeploymentContext(ctx context.Context, obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.Deployment{}
	err = clientset.AppsV1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("deployments").
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

func (w *Kubernetes) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDeploymentContext(ctx, namespace, name)
}

func (w *Kubernetes) GetDeploymentContext(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &appsv1.Deployment{}
	err = clientset.AppsV1().RESTClient().Get().
		Namespace(namespace).
		Resource("deployments").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetDeployments(namespace string, selectors ...Selector) (*appsv1.DeploymentList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDeploymentsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetDeploymentsContext(ctx context.Context, namespace string, selectors ...Selector) (*appsv1.DeploymentList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &appsv1.DeploymentList{}
	err = clientset.AppsV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteDeployment(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteDeploymentContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteDeploymentContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("deployments").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) DeleteDeployments(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteDeploymentsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) DeleteDeploymentsContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.AppsV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) CreateJob(obj *batchv1.Job) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateJobContext(ctx, obj)
}

func (w *Kubernetes) CreateJobContext(ctx context.Context, obj *batchv1.Job) (*batchv1.Job, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &batchv1.Job{}
	err = clientset.BatchV1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("jobs").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetJob(namespace, name string) (*batchv1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetJobContext(ctx, namespace, name)
}

func (w *Kubernetes) GetJobContext(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &batchv1.Job{}
	err = clientset.BatchV1().RESTClient().Get().
		Namespace(namespace).
		Resource("jobs").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) GetJobs(namespace string, selectors ...Selector) (*batchv1.JobList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetJobsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) GetJobsContext(ctx context.Context, namespace string, selectors ...Selector) (*batchv1.JobList, error) {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &batchv1.JobList{}
	err = clientset.BatchV1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("jobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Kubernetes) DeleteJob(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteJobContext(ctx, namespace, name)
}

func (w *Kubernetes) DeleteJobContext(ctx context.Context, namespace, name string) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.BatchV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("jobs").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) DeleteJobs(namespace string, selectors ...Selector) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteJobsContext(ctx, namespace, selectors...)
}

func (w *Kubernetes) DeleteJobsContext(ctx context.Context, namespace string, selectors ...Selector) error {
	clientset, err := kubernetes.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	bg := metav1.DeletePropagationBackground

	return clientset.BatchV1().RESTClient().Delete().
		Namespace(namespace).
		Resource("jobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w *Kubernetes) Apply(manifest []byte, timeout time.Duration, options ...BatchOption) error {
	cli, err := dynamic.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	c, err := discovery.NewDiscoveryClientForConfig(w.Config)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c))
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	return WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, gvk, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dri = cli.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dri = cli.Resource(mapping.Resource)
		}

		var postBatchOptions []*PostBatchOption
		for i := range options {
			if opt, ok := options[i].(*PreBatchOption); ok {
				if err := opt.Do(obj, mapping); err != nil {
					if err == ErrManifestInterceptorObjectAbort {
						return nil
					}
					return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
				}
			} else if opt, ok := options[i].(*PostBatchOption); ok {
				postBatchOptions = append(postBatchOptions, opt)
			}
		}

		oldObj, err := get(dri, timeout, obj.GetName(), metav1.GetOptions{})
		if err == nil {
			// NOTE: Patch object with MergePatchType to prevent
			// conflict error caused by `managedFields`` like:
			//
			// ```
			// error patching Secret test-wordpress321/mysql-pass:
			// Apply failed with 4 conflicts: conflicts with "velero-server" using v1:
			// - .metadata.labels.map.tanzu.vmware.com/component
			// - .metadata.labels.map.tanzu.vmware.com/federation
			// - .metadata.labels.map.tanzu.vmware.com/instance
			// - .metadata.labels.map.tanzu.vmware.com/revision
			// ```
			//
			// As using MergePatchType may cause some unpredictable behavior,
			// we only use it for patching the `managedFields`. We will then patch
			// the object with ApplyPatchType (equivalent to the `--server-side`
			// flag in `kubectl`) again to make sure the object's value is exactly
			// what we want.
			//
			// Reference:
			// - https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#notes-on-the-strategic-merge-patch
			// - https://kubernetes.io/docs/reference/using-api/server-side-apply/#clearing-managedfields
			if len(oldObj.GetManagedFields()) > 0 {
				var foundManagedField *metav1.ManagedFieldsEntry
				managedFields := oldObj.GetManagedFields()
				for i := range managedFields {
					each := &managedFields[i]
					if each.Manager == FieldManager && each.Operation == metav1.ManagedFieldsOperationApply {
						foundManagedField = each
						break
					}
				}

				if foundManagedField == nil {
					obj.SetManagedFields(append(managedFields, metav1.ManagedFieldsEntry{
						Manager:    FieldManager,
						Operation:  metav1.ManagedFieldsOperationApply,
						APIVersion: oldObj.GetAPIVersion(),
						FieldsType: "FieldsV1",
					}))

					var buf bytes.Buffer
					if err := ser.Encode(obj, &buf); err != nil {
						return &manifestAPIError{Phase: "encode", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
					}
					if _, err := patch(dri, timeout, obj.GetName(), types.MergePatchType, buf.Bytes(), metav1.PatchOptions{FieldManager: FieldManager}); err != nil {
						if len(postBatchOptions) > 0 {
							for i := range postBatchOptions {
								if e := postBatchOptions[i].Do(obj, mapping, err); e != nil {
									return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
								}
							}
							return nil
						}
						return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
					}
				}
			}
		} else if !IsErrObjectNotExists(err) {
			return &manifestAPIError{Phase: "get", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
		}

		// equivalent to: --server-side --force-conflicts
		if _, err := patch(dri, timeout, obj.GetName(), types.ApplyPatchType, p, metav1.PatchOptions{FieldManager: FieldManager, Force: util.BoolPtr(true)}); err != nil {
			if len(postBatchOptions) > 0 {
				for i := range postBatchOptions {
					if e := postBatchOptions[i].Do(obj, mapping, err); e != nil {
						return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
					}
				}
				return nil
			}
			return &manifestAPIError{Phase: "patch", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
		}
		return nil
	})
}

func (w *Kubernetes) Delete(manifest []byte, timeout time.Duration, options ...BatchOption) error {
	cli, err := dynamic.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	c, err := discovery.NewDiscoveryClientForConfig(w.Config)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c))
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	prop := metav1.DeletePropagationForeground

	manifest, err = ReverseManifest(manifest)
	if err != nil {
		return err
	}

	return WalkManifest(manifest, func(p []byte) error {
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
			if meta.IsNoMatchError(err) {
				return nil
			}
			return &manifestAPIError{Phase: "delete", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dri = cli.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dri = cli.Resource(mapping.Resource)
		}

		var postBatchOptions []*PostBatchOption
		for i := range options {
			if opt, ok := options[i].(*PreBatchOption); ok {
				if err := opt.Do(obj, mapping); err != nil {
					if err == ErrManifestInterceptorObjectAbort {
						return nil
					}
					return &manifestAPIError{Phase: "delete", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
				}
			} else if opt, ok := options[i].(*PostBatchOption); ok {
				postBatchOptions = append(postBatchOptions, opt)
			}
		}

		ctx := context.Background()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &prop}
		if err := dri.Delete(ctx, obj.GetName(), deleteOptions); err != nil && !IsErrObjectNotExists(err) {
			if len(postBatchOptions) > 0 {
				for i := range postBatchOptions {
					if e := postBatchOptions[i].Do(obj, mapping, err); e != nil {
						return &manifestAPIError{Phase: "delete", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
					}
				}
				return nil
			}
			return &manifestAPIError{Phase: "delete", Kind: obj.GetKind(), Namespace: obj.GetNamespace(), Name: obj.GetName(), Err: err}
		}
		return nil
	})
}

func NewKubernetesFromKubeConfig(p []byte) (*Kubernetes, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(p)
	if err != nil {
		return nil, err
	}
	return &Kubernetes{Config: cfg}, nil
}

func NewKubernetesFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*Kubernetes, error) {
	return &Kubernetes{Config: &rest.Config{
		Host:            endpoint,
		BearerToken:     string(token),
		TLSClientConfig: clientCfg,
	}}, nil
}

func NewKubernetes(cfg *rest.Config) (*Kubernetes, error) {
	return &Kubernetes{Config: cfg}, nil
}

type manifestAPIError struct {
	Kind      string
	Namespace string
	Name      string
	Phase     string
	Err       error
}

func (e *manifestAPIError) Error() string {
	name := e.Name
	if ns := e.Namespace; ns != "" {
		name = ns + "/" + name
	}
	return fmt.Sprintf("failed to %s %s %s: %v", e.Phase, e.Kind, name, e.Err)
}

func GetClusterNameFromKubeConfig(p []byte) (string, error) {
	cfg, err := clientcmd.Load(p)
	if err != nil {
		return "", err
	}
	if cfg.Contexts == nil {
		return "", errors.New("no context found")
	}
	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok || ctx == nil {
		return "", errors.New("no context found")
	}
	return ctx.Cluster, nil
}

func qualifySelectors(selectors ...Selector) (fieldSelector, labelSelector string) {
	var fields, labels []string
	for _, selector := range selectors {
		switch selector.(type) {
		case FieldSelector:
			fields = append(fields, selector.String())
		case LabelSelector:
			labels = append(labels, selector.String())
		}
	}
	return strings.Join(fields, ","), strings.Join(labels, ",")
}

func ReverseManifest(manifest []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	var objs []runtime.Object

	if err := WalkManifest(manifest, func(p []byte) error {
		obj := &unstructured.Unstructured{}
		_, _, err := ser.Decode(p, nil, obj)
		if err != nil {
			return err
		}
		objs = append(objs, obj)
		return nil
	}); err != nil {
		return nil, err
	}

	for i, j := 0, len(objs)-1; i < j; i, j = i+1, j-1 {
		objs[i], objs[j] = objs[j], objs[i]
	}
	for _, each := range objs {
		if err := ser.Encode(each, buf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func WalkManifest(manifest []byte, do func(p []byte) error) error {
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 100)

	for {
		raw := &runtime.RawExtension{}
		err := decoder.Decode(raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if do != nil {
			if err := do(raw.Raw); err != nil {
				return err
			}
		}
	}
	return nil
}

func get(dri dynamic.ResourceInterface, timeout time.Duration, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return dri.Get(ctx, name, options, subresources...)
}

func patch(dri dynamic.ResourceInterface, timeout time.Duration, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return dri.Patch(ctx, name, pt, data, options, subresources...)
}

type BatchOption interface {
	isBatchOption()
}

type PreBatchOption struct {
	do func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error
}

func (o *PreBatchOption) isBatchOption() {}

func (o *PreBatchOption) Do(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error {
	return o.do(obj, mapping)
}

func NewPreBatchOption(fn PreBatchOptionFunc) *PreBatchOption {
	return &PreBatchOption{do: fn}
}

func isPreBatchOption(opt BatchOption) bool {
	_, ok := opt.(*PreBatchOption)
	return ok
}

type PostBatchOption struct {
	do func(obj *unstructured.Unstructured, mapping *meta.RESTMapping, err error) error
}

func (o *PostBatchOption) isBatchOption() {}

func (o *PostBatchOption) Do(obj *unstructured.Unstructured, mapping *meta.RESTMapping, err error) error {
	return o.do(obj, mapping, err)
}

func NewPostBatchOption(fn PostBatchOptionFunc) *PostBatchOption {
	return &PostBatchOption{do: fn}
}

func isPostBatchOption(opt BatchOption) bool {
	_, ok := opt.(*PostBatchOption)
	return ok
}

type PreBatchOptionFunc func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error
type PostBatchOptionFunc func(obj *unstructured.Unstructured, mapping *meta.RESTMapping, err error) error

var ErrManifestInterceptorObjectAbort = fmt.Errorf("abort object")
