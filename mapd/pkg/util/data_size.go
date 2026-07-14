// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"strconv"
)

const (
	Ki = 1 << 10
	Mi = Ki << 10
	Gi = Mi << 10
	Ti = Gi << 10
	Pi = Ti << 10
	Ei = Pi << 10

	K = 1000
	M = K * 1000
	G = M * 1000
	T = G * 1000
	P = T * 1000
	E = P * 1000
)

type DataSize uint64

func (d DataSize) Ki() float64 {
	return float64(d) / Ki
}

func (d DataSize) Mi() float64 {
	return float64(d) / Mi
}

func (d DataSize) Gi() float64 {
	return float64(d) / Gi
}

func (d DataSize) Ti() float64 {
	return float64(d) / Ti
}

func (d DataSize) Pi() float64 {
	return float64(d) / Pi
}

func (d DataSize) Ei() float64 {
	return float64(d) / Ei
}

func (d DataSize) K() float64 {
	return float64(d) / K
}

func (d DataSize) M() float64 {
	return float64(d) / M
}

func (d DataSize) G() float64 {
	return float64(d) / G
}

func (d DataSize) T() float64 {
	return float64(d) / T
}

func (d DataSize) P() float64 {
	return float64(d) / P
}

func (d DataSize) E() float64 {
	return float64(d) / E
}

func (d DataSize) String() string {
	return d.CanonicalString('b')
}

func (d DataSize) CanonicalString(fmt rune) string {
	if d == 0 {
		return "0"
	}

	var buf bytes.Buffer
	switch fmt {
	case 'b':
		switch {
		case d%Ei == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Ei, 10) + "Ei")
		case d%Pi == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Pi, 10) + "Pi")
		case d%Ti == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Ti, 10) + "Ti")
		case d%Gi == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Gi, 10) + "Gi")
		case d%Mi == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Mi, 10) + "Mi")
		case d%Ki == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/Ki, 10) + "Ki")
		default:
			buf.WriteString(strconv.FormatUint(uint64(d), 10))
		}
	case 'd':
		switch {
		case d%E == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/E, 10) + "E")
		case d%P == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/P, 10) + "P")
		case d%T == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/T, 10) + "T")
		case d%G == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/G, 10) + "G")
		case d%M == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/M, 10) + "M")
		case d%K == 0:
			buf.WriteString(strconv.FormatUint(uint64(d)/K, 10) + "k")
		default:
			buf.WriteString(strconv.FormatUint(uint64(d), 10))
		}
	default:
		buf.WriteString(strconv.FormatUint(uint64(d), 10))
	}
	return buf.String()
}

func (d DataSize) HumanReadable(fmt rune, prec int) string {
	if d == 0 {
		return "0"
	}

	var buf bytes.Buffer
	switch fmt {
	case 'b':
		switch {
		case d > Ei:
			buf.WriteString(strconv.FormatFloat(float64(d)/Ei, 'f', prec, 64) + "Ei")
		case d > Pi:
			buf.WriteString(strconv.FormatFloat(float64(d)/Pi, 'f', prec, 64) + "Pi")
		case d > Ti:
			buf.WriteString(strconv.FormatFloat(float64(d)/Ti, 'f', prec, 64) + "Ti")
		case d > Gi:
			buf.WriteString(strconv.FormatFloat(float64(d)/Gi, 'f', prec, 64) + "Gi")
		case d > Mi:
			buf.WriteString(strconv.FormatFloat(float64(d)/Mi, 'f', prec, 64) + "Mi")
		case d > Ki:
			buf.WriteString(strconv.FormatFloat(float64(d)/Ki, 'f', prec, 64) + "Ki")
		default:
			buf.WriteString(strconv.FormatUint(uint64(d), 10))
		}
	case 'd':
		switch {
		case d > E:
			buf.WriteString(strconv.FormatFloat(float64(d)/E, 'f', prec, 64) + "E")
		case d > P:
			buf.WriteString(strconv.FormatFloat(float64(d)/P, 'f', prec, 64) + "P")
		case d > T:
			buf.WriteString(strconv.FormatFloat(float64(d)/T, 'f', prec, 64) + "T")
		case d > G:
			buf.WriteString(strconv.FormatFloat(float64(d)/G, 'f', prec, 64) + "G")
		case d > M:
			buf.WriteString(strconv.FormatFloat(float64(d)/M, 'f', prec, 64) + "M")
		case d > K:
			buf.WriteString(strconv.FormatFloat(float64(d)/K, 'f', prec, 64) + "k")
		default:
			buf.WriteString(strconv.FormatUint(uint64(d), 10))
		}
	default:
		buf.WriteString(strconv.FormatUint(uint64(d), 10))
	}
	return buf.String()
}
