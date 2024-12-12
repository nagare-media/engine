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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func IsInDeletion[T interface{ GetDeletionTimestamp() *metav1.Time }](obj T) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

func WorkflowHasTerminated(wf *enginev1.Workflow) bool {
	return !WorkflowIsActive(wf)
}

func WorkflowIsActive(wf *enginev1.Workflow) bool {
	return wf != nil &&
		// we consider any other status as active, i.e. also "" before reconciliation
		wf.Status.Phase != enginev1.SucceededWorkflowPhase &&
		wf.Status.Phase != enginev1.FailedWorkflowPhase
}

func TaskHasTerminated(task *enginev1.Task) bool {
	return !TaskIsActive(task)
}

func TaskIsActive(task *enginev1.Task) bool {
	return task != nil &&
		// we consider any other status as active, i.e. also "" before reconciliation
		task.Status.Phase != enginev1.SucceededTaskPhase &&
		task.Status.Phase != enginev1.FailedTaskPhase
}

func JobHasTerminated(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Status == corev1.ConditionTrue &&
			(c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed || c.Type == batchv1.JobFailureTarget) {
			return true
		}
	}
	return false
}

func JobIsActive(job *batchv1.Job) bool {
	return !JobHasTerminated(job)
}

func JobWasSuccessful(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
