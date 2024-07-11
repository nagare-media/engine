/*
Copyright 2022-2024 The nagare media authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func SetCondition(conditions []enginev1.Condition, cNew enginev1.Condition) []enginev1.Condition {
	for i, cOld := range conditions {
		if cOld.Type == cNew.Type {
			if cOld.Status != cNew.Status || cOld.Message != cNew.Message || cOld.Reason != cNew.Reason {
				cNew.LastTransitionTime = metav1.Time{Time: time.Now()}
				conditions[i] = cNew
			}
			return conditions
		}
	}
	cNew.LastTransitionTime = metav1.Time{Time: time.Now()}
	return append(conditions, cNew)
}

func MarkCondition(conditions []enginev1.Condition, t enginev1.ConditionType, s corev1.ConditionStatus) []enginev1.Condition {
	return SetCondition(conditions, enginev1.Condition{
		Type:   t,
		Status: s,
	})
}

func MarkConditionTrue(conditions []enginev1.Condition, t enginev1.ConditionType) []enginev1.Condition {
	return MarkCondition(conditions, t, corev1.ConditionTrue)
}

func MarkConditionFalse(conditions []enginev1.Condition, t enginev1.ConditionType) []enginev1.Condition {
	return MarkCondition(conditions, t, corev1.ConditionFalse)
}

func MarkConditionUnknown(conditions []enginev1.Condition, t enginev1.ConditionType) []enginev1.Condition {
	return MarkCondition(conditions, t, corev1.ConditionUnknown)
}

func DeleteCondition(conditions []enginev1.Condition, t enginev1.ConditionType) []enginev1.Condition {
	newConditions := make([]enginev1.Condition, 0, len(conditions))
	for _, c := range conditions {
		if c.Type != t {
			newConditions = append(newConditions, c)
		}
	}
	return newConditions
}
