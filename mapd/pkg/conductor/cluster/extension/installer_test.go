// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestUnmarshalEnvFromExtensionTreatsEmptySpecAsEmptyEnv(t *testing.T) {
	tests := []struct {
		name string
		spec []byte
	}{
		{name: "nil", spec: nil},
		{name: "empty", spec: []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			ext := &model.ClusterExtension{Spec: tt.spec}

			if err := unmarshalEnvFromExtension(ext, env); err != nil {
				t.Fatalf("unmarshalEnvFromExtension() error = %v", err)
			}
			if len(env) != 0 {
				t.Fatalf("unmarshalEnvFromExtension() env length = %d, want 0", len(env))
			}
		})
	}
}
