// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cryptoutil_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
)

func TestAESGCMCipher(t *testing.T) {
	id := uuid.New()
	cipher, err := cryptoutil.NewAESGCMCipher(id[:])
	assert.NoError(t, err)
	plain := []byte("hello world, this is my super secret code stuff.")
	encrypted := cipher.Encrypt(plain)
	assert.Equal(t, plain, cipher.Decrypt(encrypted))
}
