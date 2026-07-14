// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNamespaces(client *k8s.Kubernetes, selectors ...k8s.Selector) ([]corev1.Namespace, error) {
	list, err := client.GetNamespaces(selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertNamespace(namespace *model.Namespace, obj *corev1.Namespace) {
	namespace.Name = obj.ObjectMeta.Name
	namespace.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		namespace.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		namespace.CreatedAt = time.Now()
	}
	if namespace.CreatedAt.IsZero() {
		namespace.CreatedAt = time.Unix(0, 0)
	}
	if namespace.LastModified.IsZero() {
		namespace.LastModified = namespace.CreatedAt
	}
}

func createDefaultDockerConfigJSONSecretForNamespaceIfNotExists(c *k8s.Kubernetes, name string, registries []model1.ImageRegistry) error {
	_, err := c.GetSecret(name, defRegistrySecret)
	if err == nil {
		return nil
	}
	if !k8s.IsErrObjectNotExists(err) {
		return err
	}

	cred := &k8s.DockerConfigJSON{
		Auths: make(map[string]*k8s.DockerConfigJSONAuth),
	}

	for _, registry := range registries {
		block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
		if err != nil {
			panic(err)
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
		panic(fmt.Errorf("error constructing .dockerconfigjson: %v", err))
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defRegistrySecret,
			Namespace: name,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{".dockerconfigjson": b},
	}
	if _, err = c.CreateSecret(secret); err != nil {
		return err
	}
	return nil
}
