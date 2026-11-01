package extension

import (
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func isPendingLoadBalancerService(service *corev1.Service) bool {
	return service != nil && service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) == 0
}

func (k *GenericInstaller) pendingLoadBalancerServices() []string {
	if !k.LoadBalancerEnabled {
		return nil
	}
	namespace := k.ServiceNamespace
	if namespace == "" {
		namespace = k.JobNamespace
	}
	var pending []string
	for _, name := range k.Services {
		service, err := getService(k.Kubernetes, namespace, name)
		if err == nil && isPendingLoadBalancerService(service) {
			pending = append(pending, namespace+"/"+name)
		}
	}
	return pending
}

func jobFailureMessage(extensionName, cluster, namespace, name string, job *batchv1.Job, reason string, pendingServices []string) string {
	if reason == "" {
		reason = "failed"
	}
	detail := reason
	if reason == "DeadlineExceeded" && job != nil && job.Spec.ActiveDeadlineSeconds != nil {
		detail = fmt.Sprintf("%s after %s", reason, (time.Duration(*job.Spec.ActiveDeadlineSeconds) * time.Second).String())
	}
	message := fmt.Sprintf("%s installation job %s/%s on cluster %q failed (%s)", extensionName, namespace, name, cluster, detail)
	if len(pendingServices) > 0 {
		message += fmt.Sprintf("; service %s is waiting for a LoadBalancer address; check EXTERNAL-IP and the AVI IPAM pool", strings.Join(pendingServices, ", "))
	} else {
		message += "; check the install job pod logs and Kubernetes events"
	}
	return message
}

func (k *GenericInstaller) failedJob(job *batchv1.Job, reason string) error {
	extensionName := k.Job
	if k.Extension != nil && k.Extension.Name != "" {
		extensionName = k.Extension.Name.String()
	}
	message := jobFailureMessage(extensionName, k.Cluster, k.JobNamespace, k.Job, job, reason, k.pendingLoadBalancerServices())
	k.log.Error(message)
	return &ErrCriticalFailure{fmt.Errorf("%s", message)}
}
