package admission

import (
	corev1 "k8s.io/api/core/v1"
)

type PodPatcher interface {
	PatchPodCreate(namespace string, pod corev1.Pod) ([]PatchOperation, error)
	PatchPodUpdate(namespace string, oldPod corev1.Pod, newPod corev1.Pod) ([]PatchOperation, error)
	PatchPodDelete(namespace string, pod corev1.Pod) ([]PatchOperation, error)
}
