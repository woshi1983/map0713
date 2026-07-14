// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	runv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha1"
	vmoperatorv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/vmoperator/v1alpha1"
	runversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned"
	runscheme "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned/scheme"
	vmoperatorversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/vmoperator/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// TanzuKubernetesGridService is a supervisor cluster.
type TanzuKubernetesGridService struct {
	*Kubernetes

	LoginUser     string
	LoginPassword []byte
}

func (w *TanzuKubernetesGridService) getToken(ctx context.Context) (t string, err error) {
	token := w.Config.BearerToken
	off, err := validateUnverifiedJWTToken(token)
	if err == nil && off > 0 {
		return token, nil
	}

	for r := 0; r < 3; r++ {
		ctxL, cancel := context.WithTimeout(ctx, time.Second*5)
		token, err = w.LoginContext(ctxL, "", "")
		cancel()
		if err != nil {
			continue
		}
		w.Config.BearerToken = token
		break
	}
	if err != nil {
		return w.Config.BearerToken, errors.Wrap(err, "failed to login")
	}
	return w.Config.BearerToken, err
}

func (w *TanzuKubernetesGridService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PingContext(ctx)
}

// PingContext checks supervisor cluster health.
func (w *TanzuKubernetesGridService) PingContext(ctx context.Context) error {
	_, err := w.getToken(ctx)
	if err != nil {
		return err
	}

	return w.Kubernetes.PingContext(ctx)
}

func (w *TanzuKubernetesGridService) GetNamespace(name string) (*corev1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNamespaceContext(ctx, name)
}

func (w *TanzuKubernetesGridService) GetNamespaceContext(ctx context.Context, name string) (*corev1.Namespace, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	return w.Kubernetes.GetNamespaceContext(ctx, name)
}

func (w *TanzuKubernetesGridService) GetNamespaces(selectors ...Selector) (*corev1.NamespaceList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNamespacesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGridService) GetNamespacesContext(ctx context.Context, selectors ...Selector) (*corev1.NamespaceList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	return w.Kubernetes.GetNamespacesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGridService) GetSecret(namespace, name string) (*corev1.Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetSecretContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGridService) GetSecretContext(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	return w.Kubernetes.GetSecretContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGridService) GetSecrets(namespace string, selectors ...Selector) (*corev1.SecretList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetSecretsContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGridService) GetSecretsContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.SecretList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	return w.Kubernetes.GetSecretsContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGridService) GetResourceQuotas(namespace string, selectors ...Selector) (*corev1.ResourceQuotaList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetResourceQuotasContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGridService) GetResourceQuotasContext(ctx context.Context, namespace string, selectors ...Selector) (*corev1.ResourceQuotaList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	return w.Kubernetes.GetResourceQuotasContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGridService) PatchTkgServiceConfiguration(name string, pt types.PatchType, patch []byte) (*runv1alpha1.TkgServiceConfiguration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PatchTkgServiceConfigurationContext(ctx, name, pt, patch)
}

func (w *TanzuKubernetesGridService) PatchTkgServiceConfigurationContext(ctx context.Context, name string, pt types.PatchType, patch []byte) (*runv1alpha1.TkgServiceConfiguration, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &runv1alpha1.TkgServiceConfiguration{}
	err = clientset.RunV1alpha1().RESTClient().Patch(pt).
		Resource("tkgserviceconfigurations").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, runscheme.ParameterCodec).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *TanzuKubernetesGridService) GetTkgServiceConfiguration(name string) (*runv1alpha1.TkgServiceConfiguration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTkgServiceConfigurationContext(ctx, name)
}

func (w *TanzuKubernetesGridService) GetTkgServiceConfigurationContext(ctx context.Context, name string) (*runv1alpha1.TkgServiceConfiguration, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha1().TkgServiceConfigurations().Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesRelease(name string) (*runv1alpha1.TanzuKubernetesRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleaseContext(ctx, name)
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesReleaseContext(ctx context.Context, name string) (*runv1alpha1.TanzuKubernetesRelease, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha1().TanzuKubernetesReleases().Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesReleases(selectors ...Selector) (*runv1alpha1.TanzuKubernetesReleaseList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleasesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesReleasesContext(ctx context.Context, selectors ...Selector) (*runv1alpha1.TanzuKubernetesReleaseList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.RunV1alpha1().TanzuKubernetesReleases().List(ctx, opts)
}

func (w *TanzuKubernetesGridService) CreateTanzuKubernetesCluster(obj *runv1alpha1.TanzuKubernetesCluster) (*runv1alpha1.TanzuKubernetesCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateTanzuKubernetesClusterContext(ctx, obj)
}

func (w *TanzuKubernetesGridService) CreateTanzuKubernetesClusterContext(ctx context.Context, obj *runv1alpha1.TanzuKubernetesCluster) (*runv1alpha1.TanzuKubernetesCluster, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha1().TanzuKubernetesClusters(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
}

func (w *TanzuKubernetesGridService) UpdateTanzuKubernetesCluster(obj *runv1alpha1.TanzuKubernetesCluster) (*runv1alpha1.TanzuKubernetesCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateTanzuKubernetesClusterContext(ctx, obj)
}

func (w *TanzuKubernetesGridService) UpdateTanzuKubernetesClusterContext(ctx context.Context, obj *runv1alpha1.TanzuKubernetesCluster) (*runv1alpha1.TanzuKubernetesCluster, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha1().TanzuKubernetesClusters(obj.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesCluster(namespace, name string) (*runv1alpha1.TanzuKubernetesCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesClusterContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesClusterContext(ctx context.Context, namespace, name string) (*runv1alpha1.TanzuKubernetesCluster, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha1().TanzuKubernetesClusters(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesClusters(namespace string, selectors ...Selector) (*runv1alpha1.TanzuKubernetesClusterList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesClustersContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGridService) GetTanzuKubernetesClustersContext(ctx context.Context, namespace string, selectors ...Selector) (*runv1alpha1.TanzuKubernetesClusterList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &runv1alpha1.TanzuKubernetesClusterList{}
	err = clientset.RunV1alpha1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("tanzukubernetesclusters").
		VersionedParams(&opts, runscheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *TanzuKubernetesGridService) DeleteTanzuKubernetesCluster(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteTanzuKubernetesClusterContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGridService) DeleteTanzuKubernetesClusterContext(ctx context.Context, namespace, name string) error {
	_, err := w.getToken(ctx)
	if err != nil {
		return err
	}

	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground
	return clientset.RunV1alpha1().TanzuKubernetesClusters(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &bg})
}

func (w *TanzuKubernetesGridService) GetVirtualMachineClass(name string) (*vmoperatorv1alpha1.VirtualMachineClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVirtualMachineClassContext(ctx, name)
}

func (w *TanzuKubernetesGridService) GetVirtualMachineClassContext(ctx context.Context, name string) (*vmoperatorv1alpha1.VirtualMachineClass, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := vmoperatorversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.VmoperatorV1alpha1().VirtualMachineClasses().Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGridService) GetVirtualMachineClasses(selectors ...Selector) (*vmoperatorv1alpha1.VirtualMachineClassList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVirtualMachineClassesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGridService) GetVirtualMachineClassesContext(ctx context.Context, selectors ...Selector) (*vmoperatorv1alpha1.VirtualMachineClassList, error) {
	_, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := vmoperatorversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.VmoperatorV1alpha1().VirtualMachineClasses().List(ctx, opts)
}

func (w *TanzuKubernetesGridService) Login(namespace, name string) (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.LoginContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGridService) LoginContext(ctx context.Context, namespace, name string) (token string, err error) {
	var clientset *runversioned.Clientset
	clientset, err = runversioned.NewForConfig(w.Config)
	if err != nil {
		return
	}

	var b, reqBody []byte
	reqBody, err = json.Marshal(&wcpLoginRequest{GuestClusterNamespace: namespace, GuestClusterName: name})
	if err != nil {
		return token, fmt.Errorf("unable to construct the login request: %v", err)
	}

	b, err = clientset.Discovery().RESTClient().Post().
		AbsPath("/wcp/login").
		Body(reqBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(w.LoginUser+":"+string(w.LoginPassword)))).
		Do(ctx).
		Raw()
	if err != nil {
		return token, fmt.Errorf("error submitting the login request: %v", err)
	}

	d := &wcpLoginResponse{}
	if err := json.Unmarshal(b, d); err != nil {
		return token, fmt.Errorf("unable to parse the login response: %v", err)
	}
	return d.SessionID, nil
}

// NewTanzuKubernetesGridService creates a new Tanzu Kubernetes supervisor instance.
func NewTanzuKubernetesGridService(endpoint, user string, password []byte, clientCfg rest.TLSClientConfig) (*TanzuKubernetesGridService, error) {
	k8s, err := NewKubernetesFromToken(endpoint, []byte{}, clientCfg)
	if err != nil {
		return nil, err
	}
	s := &TanzuKubernetesGridService{
		Kubernetes:    k8s,
		LoginUser:     user,
		LoginPassword: password,
	}
	return s, nil
}

func validateUnverifiedJWTToken(s string) (time.Duration, error) {
	t, _, err := new(jwt.Parser).ParseUnverified(s, &jwt.StandardClaims{})
	if err != nil {
		return 0, err
	}
	c := t.Claims.(*jwt.StandardClaims)
	return time.Second * time.Duration(c.ExpiresAt-time.Now().Unix()), t.Claims.Valid()
}

type wcpLoginRequest struct {
	GuestClusterName      string `json:"guest_cluster_name,omitempty"`
	GuestClusterNamespace string `json:"guest_cluster_namespace,omitempty"`
}

type wcpLoginResponse struct {
	SessionID          string `json:"session_id"`
	GuestClusterServer string `json:"guest_cluster_server"`
	GuestClusterCA     []byte `json:"guest_cluster_ca"`
}
