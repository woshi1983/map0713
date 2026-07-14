// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil

import "regexp"

var (
	// Han (89513 characters):
	//
	// U+02E80 … U+02E99
	// U+02E9B … U+02EF3
	// U+02F00 … U+02FD5
	// U+03001 … U+03003
	// U+03005 … U+03011
	// U+03013 … U+0301F
	// U+03021 … U+0302D
	// U+03030
	// U+03037 … U+0303F
	// U+030FB
	// U+03190 … U+0319F
	// U+031C0 … U+031E3
	// U+03220 … U+03247
	// U+03280 … U+032B0
	// U+032C0 … U+032CB
	// U+032FF
	// U+03358 … U+03370
	// U+0337B … U+0337F
	// U+033E0 … U+033FE
	// U+03400 … U+04DB5
	// U+04E00 … U+09FEF
	// U+0F900 … U+0FA6D
	// U+0FA70 … U+0FAD9
	// U+0FE45 … U+0FE46
	// U+0FF61 … U+0FF65
	// U+1D360 … U+1D371
	// U+1F250 … U+1F251
	// U+20000 … U+2A6D6
	// U+2A700 … U+2B734
	// U+2B740 … U+2B81D
	// U+2B820 … U+2CEA1
	// U+2CEB0 … U+2EBE0
	// U+2F800 … U+2FA1D
	//
	// Hangul (11775 characters):
	//
	// U+01100 … U+011FF
	// U+03001 … U+03003
	// U+03008 … U+03011
	// U+03013 … U+0301F
	// U+0302E … U+03030
	// U+03037
	// U+030FB
	// U+03131 … U+0318E
	// U+03200 … U+0321E
	// U+03260 … U+0327E
	// U+0A960 … U+0A97C
	// U+0AC00 … U+0D7A3
	// U+0D7B0 … U+0D7C6
	// U+0D7CB … U+0D7FB
	// U+0FE45 … U+0FE46
	// U+0FF61 … U+0FF65
	// U+0FFA0 … U+0FFBE
	// U+0FFC2 … U+0FFC7
	// U+0FFCA … U+0FFCF
	// U+0FFD2 … U+0FFD7
	// U+0FFDA … U+0FFDC
	//
	// Hiragana (431 characters):
	//
	// U+03001 … U+03003
	// U+03008 … U+03011
	// U+03013 … U+0301F
	// U+03030 … U+03035
	// U+03037
	// U+0303C … U+0303D
	// U+03041 … U+03096
	// U+03099 … U+030A0
	// U+030FB … U+030FC
	// U+0FE45 … U+0FE46
	// U+0FF61 … U+0FF65
	// U+0FF70
	// U+0FF9E … U+0FF9F
	// U+1B001 … U+1B11E
	// U+1B150 … U+1B152
	// U+1F200
	//
	// Katakana (356 characters)
	//
	// U+03001 … U+03003
	// U+03008 … U+03011
	// U+03013 … U+0301F
	// U+03030 … U+03035
	// U+03037
	// U+0303C … U+0303D
	// U+03099 … U+0309C
	// U+030A0 … U+030FF
	// U+031F0 … U+031FF
	// U+032D0 … U+032FE
	// U+03300 … U+03357
	// U+0FE45 … U+0FE46
	// U+0FF61 … U+0FF9F
	// U+1B000
	// U+1B164 … U+1B167
	//
	// but only 4-byte characters are matched. (The others are emoji.)
	cjk = regexp.MustCompile("[\u2E80-\u2E99\u2E9B-\u2EF3\u2F00-\u2FD5\u3001-\u3003\u3005-\u3011\u3013-\u301F\u3021-\u302D\u3030\u3037-\u303F\u30FB\u3190-\u319F\u31C0-\u31E3\u3220-\u3247\u3280-\u32B0\u32C0-\u32CB\u32FF\u3358-\u3370\u337B-\u337F\u33E0-\u33FE\u3400-\u4DB5\u4E00-\u9FEF\uF900-\uFA6D\uFA70-\uFAD9\uFE45-\uFE46\uFF61-\uFF65\u1100-\u11FF\u3001-\u3003\u3008-\u3011\u3013-\u301F\u302E-\u3030\u3037\u30FB\u3131-\u318E\u3200-\u321E\u3260-\u327E\uA960-\uA97C\uAC00-\uD7A3\uD7B0-\uD7C6\uD7CB-\uD7FB\uFE45-\uFE46\uFF61-\uFF65\uFFA0-\uFFBE\uFFC2-\uFFC7\uFFCA-\uFFCF\uFFD2-\uFFD7\uFFDA-\uFFDC\u3001-\u3003\u3008-\u3011\u3013-\u301F\u3030-\u3035\u3037\u303C-\u303D\u3041-\u3096\u3099-\u30A0\u30FB-\u30FC\uFE45-\uFE46\uFF61-\uFF65\uFF70\uFF9E-\uFF9F\u3001-\u3003\u3008-\u3011\u3013-\u301F\u3030-\u3035\u3037\u303C-\u303D\u3099-\u309C\u30A0-\u30FF\u31F0-\u31FF\u32D0-\u32FE\u3300-\u3357\uFE45-\uFE46\uFF61-\uFF9F]+")
)

// ContainsCJK returns true if given block b contains CJK characters.
func ContainsCJK(b []byte) bool {
	return cjk.Match(b)
}
