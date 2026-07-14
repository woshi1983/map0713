// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
)

func TestContainsCJK(t *testing.T) {
	assert.False(t, textutil.ContainsCJK([]byte("English (U.S.)")))
	assert.True(t, textutil.ContainsCJK([]byte("にほんご")))
	assert.True(t, textutil.ContainsCJK([]byte("中文（简体）")))
	assert.True(t, textutil.ContainsCJK([]byte("日本語")))
	assert.True(t, textutil.ContainsCJK([]byte("한국어")))
	assert.True(t, textutil.ContainsCJK([]byte("ﾆﾎﾝｺﾞ")))
}
