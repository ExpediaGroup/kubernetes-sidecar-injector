package admission

import (
	corev1 "k8s.io/api/core/v1"
)

type PodPatcher interface {
	PatchPodCreate(pod corev1.Pod) ([]PatchOperation, error)
	PatchPodUpdate(oldPod corev1.Pod, newPod corev1.Pod) ([]PatchOperation, error)
	PatchPodDelete(pod corev1.Pod) ([]PatchOperation, error)
}
