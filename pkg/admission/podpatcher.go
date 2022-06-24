package admission

import (
	"context"
	corev1 "k8s.io/api/core/v1"
)

// PodPatcher Pod patching interface
type PodPatcher interface {
	PatchPodCreate(ctx context.Context, namespace string, pod corev1.Pod) ([]PatchOperation, error)
	PatchPodUpdate(ctx context.Context, namespace string, oldPod corev1.Pod, newPod corev1.Pod) ([]PatchOperation, error)
	PatchPodDelete(ctx context.Context, namespace string, pod corev1.Pod) ([]PatchOperation, error)
}
