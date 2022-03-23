package webhook

import (
	"context"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type SideCars struct {
	Sidecars []InjectSideCar `yaml:"sidecars"`
}

/*InjectSideCar is a named sidecar to be injected*/
type InjectSideCar struct {
	Name    string  `yaml:"name"`
	Sidecar SideCar `yaml:"sidecar"`
}

type SidecarInjectorPatcher struct {
	K8sClient      kubernetes.Interface
	InjectPrefix   string
	InjectName     string
	SidecarDataKey string
}

func (patcher *SidecarInjectorPatcher) sideCarInjectionAnnotation() string {
	return patcher.InjectPrefix + "/" + patcher.InjectName
}

func (patcher *SidecarInjectorPatcher) configmapSidecarNames(pod corev1.Pod) ([]string, bool) {
	annotations := map[string]string{}
	if pod.GetAnnotations() != nil {
		annotations = pod.GetAnnotations()
	}
	if sidecars, ok := annotations[patcher.sideCarInjectionAnnotation()]; ok {
		parts := strings.Split(sidecars, ",")
		for i := range parts {
			parts[i] = strings.Trim(parts[i], " ")
		}

		if len(parts) > 0 {
			log.Infof("sideCar injection for %v/%v: sidecars: %v", pod.GetNamespace(), pod.GetName(), sidecars)
			return parts, true
		}
	}
	log.Infof("Skipping mutation for [%v]. No action required", pod.GetName())
	return nil, false
}

func createPatchOperation(path string, index int, targetLength int, value func(bool) interface{}) admission.PatchOperation {
	first := index == 0 && targetLength == 0
	if !first {
		path = path + "/-"
	}
	return admission.PatchOperation{
		Op:    "add",
		Path:  path,
		Value: value(first),
	}
}

func (patcher *SidecarInjectorPatcher) PatchPodCreate(pod corev1.Pod) ([]admission.PatchOperation, error) {
	ctx := context.Background()
	if configmapSidecarNames, ok := patcher.configmapSidecarNames(pod); ok {
		var patches []admission.PatchOperation
		for _, configmapSidecarName := range configmapSidecarNames {
			configmapSidecar, err := patcher.K8sClient.CoreV1().ConfigMaps(pod.GetNamespace()).Get(ctx, configmapSidecarName, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				log.Warnf("sidecar configmap %s/%s was not found", pod.GetNamespace(), configmapSidecarName)
			} else if err != nil {
				log.Errorf("error fetching sidecar configmap %s/%s", pod.GetNamespace(), configmapSidecarName)
			} else if sidecarsStr, ok := configmapSidecar.Data[patcher.SidecarDataKey]; ok {
				var sidecars []SideCar
				if err := yaml.Unmarshal([]byte(sidecarsStr), &sidecars); err != nil {
					log.Errorf("error unmarshalling %s from configmap %s/%s", patcher.SidecarDataKey, pod.GetNamespace(), configmapSidecarName)
				}
				if sidecars != nil {
					for _, sidecar := range sidecars {
						for index, container := range sidecar.Containers {
							patchOperation := createPatchOperation("/spec/containers", index, len(pod.Spec.Containers), func(first bool) interface{} {
								if first {
									return []corev1.Container{container}
								}
								return container
							})
							patches = append(patches, patchOperation)
						}
						for index, volume := range sidecar.Volumes {
							patchOperation := createPatchOperation("/spec/containers", index, len(pod.Spec.Volumes), func(first bool) interface{} {
								if first {
									return []corev1.Volume{volume}
								}
								return volume
							})
							patches = append(patches, patchOperation)
						}
						for index, imagePullSecret := range sidecar.ImagePullSecrets {
							patchOperation := createPatchOperation("/spec/containers", index, len(pod.Spec.ImagePullSecrets), func(first bool) interface{} {
								if first {
									return []corev1.LocalObjectReference{imagePullSecret}
								}
								return imagePullSecret
							})
							patches = append(patches, patchOperation)
						}
					}
				}
			}
		}
		return patches, nil
	} else {
		return []admission.PatchOperation{}, nil
	}
	return nil, nil
}

/*PatchPodUpdate not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodUpdate(_ corev1.Pod, _ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}

/*PatchPodDelete not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodDelete(_ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}
