// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/rest"
)

func init() {
	Registry.Register("instance", DispatcherFactory(
		mapd.EventEmitterInstance,
		func() error {
			var instances []model.Instance
			if err := mapd.Database().Select("id").Where("`state` NOT IN (?, ?)", model.InstanceStateSuccess, model.InstanceStateFailure).Find(&instances).Error; err != nil {
				return fmt.Errorf("error obtaining existing instances: %v", err)
			}
			go func() {
				for _, each := range instances {
					mapd.EventEmitterInstance.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncInstance,
	))
}

// syncInstance handle Instance transactions.
func syncInstance(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	instance := &model.Instance{ID: id}
	finished := false
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(instance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finished = true
				return nil
			}
			return err
		}

		updates := make(map[string]interface{})
		switch instance.State {
		case model.InstanceStateSkipped, model.InstanceStateSuccess, model.InstanceStateFailure:
			finished = true
			return nil
		case model.InstanceStateDeleting:
		default:
			updates["state"] = model.InstanceStateProvisioning
			if err := tx.Model(instance).Updates(updates).Error; err != nil {
				return err
			}
			if instance.State == model.InstanceStatePending {
				// to commit state changes immediately.
				next = 0
				return nil
			}
		}

		c := &model.Cluster{ID: instance.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			return err
		}

		if c.Provisioned < model.ClusterProvisionFinish || c.Status != model.ClusterStatusHealthy {
			log.Infof("pending instance %s due to insufficient cluster condition => {status: %s, initialized: %t}", strconv.Quote(instance.Name), c.Status, c.Provisioned >= model.ClusterProvisionFinish)
			return nil
		}

		var registries []model.ImageRegistry
		if err := tx.Raw("SELECT `image_registry`.* FROM `cluster_image_registry` INNER JOIN `image_registry` ON `cluster_image_registry`.`image_registry_id` = `image_registry`.`id` WHERE `image_registry`.`deleted_at` IS NULL AND `cluster_image_registry`.`cluster_id` = ? ORDER BY `default` DESC", c.ID).Scan(&registries).Error; err != nil {
			return err
		}
		if len(registries) == 0 {
			log.Errorf("no available image registry for cluster %s.", strconv.Quote(c.Name))
			return nil
		}

		var tasks []*model.InstanceStatusTask
		if err := json.Unmarshal(instance.Status, &tasks); err != nil {
			panic(err)
		}
		block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
			Insecure: c.InsecureTLS,
			CAData:   c.CACert,
		})
		if err != nil {
			panic(err)
		}

		oldManifest, manifest := instance.PrevManifest, instance.Manifest
		if instance.State == model.InstanceStateDeleting {
			log.Infof("attempt to delete instance %s...", instance.Name)
			log.Sync()

			i := -1
			if err := deleteManifest(wc, manifest, k8s.NewPreBatchOption(func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error {
				defer log.Sync()

				i++

				name := obj.GetName()
				if namespace := obj.GetNamespace(); namespace != "" {
					name = namespace + "/" + name
				}

				switch obj.GetKind() {
				case "Namespace":
					log.Infof("skipped resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
					return k8s.ErrManifestInterceptorObjectAbort
				case "Secret":
					if obj.GetName() == defRegistrySecret {
						log.Infof("skipped resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
						return k8s.ErrManifestInterceptorObjectAbort
					}
				}
				tasks[len(tasks)-i-1].State = model.InstanceStateSuccess // assume it will be succeeded.
				tasks[len(tasks)-i-1].Message = ""
				log.Infof("deleting resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
				return nil
			})); err != nil {
				task := tasks[len(tasks)-i-1]
				task.State = model.InstanceStateFailure
				task.Message = err.Error()
				task.Retry++
				log.Errorf("failed to delete instance %s due to: %v. It will be retried later.", instance.Name, err)

				b, err := json.Marshal(tasks)
				if err != nil {
					panic(err)
				}
				updates["status"] = string(b)

				if err := tx.Model(instance).Updates(updates).Error; err != nil {
					return err
				}
				return nil
			}

			if err := tx.Delete(instance).Error; err != nil {
				return err
			}
			finished = true
			return nil
		}

		log.Infof("attempt to provision instance %s...", instance.Name)
		log.Sync()

		buf := bytes.NewBuffer(nil)
		ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		objsToDelete, _, err := compareManifest(oldManifest, manifest)
		if err != nil {
			log.Errorf("error occured while comparing manifest changes for instance %s: %v.", instance.Name, err)
			return nil
		}
		for _, each := range objsToDelete {
			if err := ser.Encode(each, buf); err != nil {
				// this is impossible in common cases.
				panic(err)
			}
		}

		// remove unless objects
		if err := deleteManifest(wc, buf.Bytes(), k8s.NewPreBatchOption(func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error {
			defer log.Sync()

			name := obj.GetName()
			if namespace := obj.GetNamespace(); namespace != "" {
				name = namespace + "/" + name
			}
			log.Infof("deleting instance resources (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
			return nil
		}), k8s.NewPostBatchOption(func(obj *unstructured.Unstructured, mapping *meta.RESTMapping, err error) error {
			defer log.Sync()

			name := obj.GetName()
			if namespace := obj.GetNamespace(); namespace != "" {
				name = namespace + "/" + name
			}

			if k8s.IsErrInvalid(err) {
				log.Infof("ignoring invalid resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
				return nil
			}
			return err
		})); err != nil {
			log.Errorf("failed to provision instance %s due to: %v. It will be retried later.", instance.Name, err)
			return nil
		}

		// create namespace
		if err := createNamespaceIfNotExist(wc, instance.Namespace, registries); err != nil {
			log.Errorf("failed to provision instance %s on creating namespace %s: %v. It will be retried later.", instance.Name, instance.Namespace, err)
			return nil
		}

		// install CRDs
		if err := applyManifest(wc, instance.CRDs, k8s.NewPreBatchOption(func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error {
			defer log.Sync()

			name := obj.GetName()
			if namespace := obj.GetNamespace(); namespace != "" {
				name = namespace + "/" + name
			}
			log.Infof("applying custom resource definitions (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
			return nil
		})); err != nil {
			log.Errorf("failed to provision instance %s on installing CRDs: %v. It will be retried later.", instance.Name, err)
			return nil
		}

		// apply manifest
		i := -1
		if err := applyManifest(wc, manifest, k8s.NewPreBatchOption(func(obj *unstructured.Unstructured, mapping *meta.RESTMapping) error {
			defer log.Sync()

			i++

			name := obj.GetName()
			if namespace := obj.GetNamespace(); namespace != "" {
				name = namespace + "/" + name
			}

			switch obj.GetKind() {
			case "Namespace":
				tasks[i].State = model.InstanceStateSkipped
				log.Warnf("skipped resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
				return k8s.ErrManifestInterceptorObjectAbort
			case "Secret":
				if obj.GetName() == defRegistrySecret {
					tasks[i].State = model.InstanceStateSkipped
					log.Warnf("skipped resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
					return k8s.ErrManifestInterceptorObjectAbort
				}
			case "Job":
				_, err := wc.GetJob(obj.GetNamespace(), obj.GetName())
				if err == nil {
					tasks[i].State = model.InstanceStateSkipped
					log.Warnf("skipped resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
					return k8s.ErrManifestInterceptorObjectAbort
				}
			}
			tasks[i].State = model.InstanceStateSuccess // assume it will be succeeded.
			tasks[i].Message = ""
			log.Infof("applying resource (%s :: %s) for instance %s.", obj.GetKind(), name, instance.Name)
			return nil
		})); err != nil {
			if i < 0 {
				i = 0
			}
			task := tasks[i]

			if task.Retry > 2 {
				task.State = model.InstanceStateFailure
				task.Message = err.Error()
				updates["state"] = model.InstanceStateFailure
				log.Errorf("failed to provision instance %s due to: %v. Retry threshold is exceeded.", instance.Name, err)
			} else {
				task.State = model.InstanceStateProvisioning
				log.Errorf("failed to provision instance %s due to: %v. It will be retried later.", instance.Name, err)
			}
			task.Retry++
			b, err := json.Marshal(tasks)
			if err != nil {
				panic(err)
			}
			updates["status"] = string(b)

			if err := tx.Model(instance).Updates(updates).Error; err != nil {
				return err
			}
			if task.State == model.InstanceStateFailure {
				finished = true
			}
			return nil
		}

		if instance.AutoRestart {
			if err := wc.DeletePods(instance.Namespace,
				k8s.LabelEqual(defLabelComponent, instance.ComponentID.String()),
				k8s.LabelEqual(defLabelManaged, "true"),
				k8s.LabelEqual(defLabelRestart, "true"),
			); err != nil {
				log.Errorf("failed to restart pods for instance %s due to: %v. It will be retried later.", instance.Name, err)
				return nil
			}
		}

		b, err := json.Marshal(tasks)
		if err != nil {
			panic(err)
		}
		updates["status"] = string(b)
		updates["state"] = model.InstanceStateSuccess
		if err := tx.Model(instance).Updates(updates).Error; err != nil {
			return err
		}
		log.Infof("instance %s is successfully provisioned.", instance.Name)
		finished = true
		return nil
	}); err != nil {
		finished = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if finished {
		mapd.EventEmitterInstance.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
		mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
			Type: mapd.EventTypeModify,
			OID:  instance.FederationID.UUID,
		})
	}
	return next
}

func createNamespaceIfNotExist(client *k8s.Kubernetes, namespace string, registries []model.ImageRegistry) error {
	if _, err := client.GetNamespace(namespace); err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			return err
		}

		ns := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					defLabelManaged: "true",
				},
			},
		}
		if _, err := client.CreateNamespace(ns); err != nil {
			return err
		}
		return createRegistrySecret(client, namespace, registries)
	}

	if _, err := client.GetSecret(namespace, defRegistrySecret); err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			return err
		}
		return createRegistrySecret(client, namespace, registries)
	}
	return nil
}

func createRegistrySecret(client *k8s.Kubernetes, namespace string, registries []model.ImageRegistry) error {
	// create default dockerconfigjson secret.
	cred := &k8s.DockerConfigJSON{
		Auths: make(map[string]*k8s.DockerConfigJSONAuth),
	}

	for _, registry := range registries {
		block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
		if err != nil {
			return err
		}

		if registry.AuthUser != "" {
			password := block.Decrypt(registry.AuthPassword)
			cred.Auths[registry.Endpoint] = &k8s.DockerConfigJSONAuth{
				Username: registry.AuthUser,
				Password: string(password),
				Auth:     []byte(strings.Join([]string{registry.AuthUser, string(password)}, ":")),
			}
		}
	}
	b, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("error constructing .dockerconfigjson: %v", err)
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defRegistrySecret,
			Namespace: namespace,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{".dockerconfigjson": b},
	}
	_, err = client.CreateSecret(secret)
	return err
}

func applyManifest(client *k8s.Kubernetes, manifest []byte, options ...k8s.BatchOption) error {
	return client.Apply(manifest, mapd.GlobalConfig.HTTPClientTimeout, options...)
}

func deleteManifest(client *k8s.Kubernetes, manifest []byte, options ...k8s.BatchOption) error {
	return client.Delete(manifest, mapd.GlobalConfig.HTTPClientTimeout, options...)
}

// compareManifest only compares kind, namespace, and name, returns their difference.
func compareManifest(p1, p2 []byte) (s1, s2 []*unstructured.Unstructured, err error) {
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
		index := fmt.Sprintf("%s-%s", obj.GetKind(), obj.GetName())
		m2[index] = obj
		i2[index] = len(i2) + 1
		return nil
	}); err != nil {
		return
	}

	for k, v := range m1 {
		if _, ok := m2[k]; !ok {
			s1 = append(s1, v)
		}
	}
	for k, v := range m2 {
		if _, ok := m1[k]; !ok {
			s2 = append(s2, v)
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

type ErrAutoCompleteObject struct {
	Err error
	Obj *unstructured.Unstructured
}

func (e *ErrAutoCompleteObject) Error() string {
	name := e.Obj.GetName()
	if ns := e.Obj.GetNamespace(); ns != "" {
		name = ns + "/" + name
	}
	return fmt.Sprintf("failed to autocomplete object %s for kind %s: %v", name, e.Obj.GetKind(), e.Err)
}

func IsErrAutoCompleteObject(err error) bool {
	_, ok := err.(*ErrAutoCompleteObject)
	return ok
}
