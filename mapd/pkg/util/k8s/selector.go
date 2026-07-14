// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Selector interface {
	String() string

	isSelector()
}

type FieldSelector interface {
	Selector

	isFieldSelector()
}

type LabelSelector interface {
	Selector

	isLabelSelector()
}

type EqualityBasedField struct {
	Name    string
	Reverse bool
	Value   string
}

func (l *EqualityBasedField) String() string {
	if l.Reverse {
		return l.Name + "!=" + l.Value
	}
	return l.Name + "=" + l.Value
}

func (l *EqualityBasedField) isSelector()      {}
func (l *EqualityBasedField) isFieldSelector() {}

func FieldEqual(name, value string) *EqualityBasedField {
	return &EqualityBasedField{Name: name, Value: value}
}

func FieldNotEqual(name, value string) *EqualityBasedField {
	return &EqualityBasedField{Name: name, Value: value, Reverse: true}
}

type EqualityBasedLabel struct {
	Name    string
	Reverse bool
	Value   string
}

func (l *EqualityBasedLabel) String() string {
	if l.Reverse {
		return l.Name + "!=" + l.Value
	}
	return l.Name + "=" + l.Value
}

func (l *EqualityBasedLabel) isSelector()      {}
func (l *EqualityBasedLabel) isLabelSelector() {}

func LabelEqual(name, value string) *EqualityBasedLabel {
	return &EqualityBasedLabel{Name: name, Value: value}
}

func LabelNotEqual(name, value string) *EqualityBasedLabel {
	return &EqualityBasedLabel{Name: name, Value: value, Reverse: true}
}

func FromLabelSelector(labelSelector *metav1.LabelSelector) (selectors []Selector) {
	if labelSelector == nil {
		return
	}

	for k, v := range labelSelector.MatchLabels {
		selectors = append(selectors, LabelEqual(k, v))
	}
	for _, exp := range labelSelector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			selectors = append(selectors, LabelIn(exp.Key, exp.Values...))
		case metav1.LabelSelectorOpNotIn:
			selectors = append(selectors, LabelNotIn(exp.Key, exp.Values...))
		case metav1.LabelSelectorOpExists:
			selectors = append(selectors, LabelExists(exp.Key))
		case metav1.LabelSelectorOpDoesNotExist:
			selectors = append(selectors, LabelNotExists(exp.Key))
		}
	}
	return
}

type SetBasedLabel struct {
	Name    string
	Reverse bool
	Values  []string
}

func (l *SetBasedLabel) String() string {
	if len(l.Values) == 0 {
		if l.Reverse {
			return "!" + l.Name
		}
		return l.Name
	}

	ss := []string{l.Name}
	if l.Reverse {
		ss = append(ss, "notin")
	} else {
		ss = append(ss, "in")
	}
	ss = append(ss, "("+strings.Join(l.Values, ", ")+")")
	return strings.Join(ss, " ")
}

func (l *SetBasedLabel) isSelector()      {}
func (l *SetBasedLabel) isLabelSelector() {}

func LabelExists(name string) *SetBasedLabel {
	return &SetBasedLabel{Name: name}
}

func LabelNotExists(name string) *SetBasedLabel {
	return &SetBasedLabel{Name: name, Reverse: true}
}

func LabelIn(name string, values ...string) *SetBasedLabel {
	return &SetBasedLabel{Name: name, Values: values}
}

func LabelNotIn(name string, values ...string) *SetBasedLabel {
	return &SetBasedLabel{Name: name, Reverse: true, Values: values}
}
