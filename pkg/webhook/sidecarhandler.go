package webhook

import (
	"context"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/ghodss/yaml"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type SideCar struct {
	Name             string                        `yaml:"name"`
	Containers       []corev1.Container            `yaml:"containers"`
	Volumes          []corev1.Volume               `yaml:"volumes"`
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
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

func (patcher *SidecarInjectorPatcher) configmapSidecarNames(namespace string, pod corev1.Pod) ([]string, bool) {
	annotations := map[string]string{}
	if pod.GetAnnotations() != nil {
		annotations = pod.GetAnnotations()
	}
	if sidecars, ok := annotations[patcher.sideCarInjectionAnnotation()]; ok {
		parts := lo.Map[string, string](strings.Split(sidecars, ","), func(part string, _ int) string {
			return strings.TrimSpace(part)
		})

		if len(parts) > 0 {
			name := pod.GetName()
			if name == "" {
				name = pod.GetGenerateName()
			}
			log.Infof("sideCar injection for %v/%v: sidecars: %v", namespace, name, sidecars)
			return parts, true
		}
	}
	log.Infof("Skipping mutation for [%v]. No action required", pod.GetName())
	return nil, false
}

func createPatches[T any](newCollection []T, existingCollection []T, path string) []admission.PatchOperation {
	var patches []admission.PatchOperation
	for index, item := range newCollection {
		first := index == 0 && len(existingCollection) == 0
		if !first {
			path = path + "/-"
		}
		var value interface{}
		if first {
			value = []T{item}
		}
		value = item
		patches = append(patches, admission.PatchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patches
}

func (patcher *SidecarInjectorPatcher) PatchPodCreate(namespace string, pod corev1.Pod) ([]admission.PatchOperation, error) {
	ctx := context.Background()
	var patches []admission.PatchOperation
	if configmapSidecarNames, ok := patcher.configmapSidecarNames(namespace, pod); ok {
		for _, configmapSidecarName := range configmapSidecarNames {
			configmapSidecar, err := patcher.K8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, configmapSidecarName, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				log.Warnf("sidecar configmap %s/%s was not found", namespace, configmapSidecarName)
			} else if err != nil {
				log.Errorf("error fetching sidecar configmap %s/%s - %v", namespace, configmapSidecarName, err)
			} else if sidecarsStr, ok := configmapSidecar.Data[patcher.SidecarDataKey]; ok {
				var sidecars []SideCar
				if err := yaml.Unmarshal([]byte(sidecarsStr), &sidecars); err != nil {
					log.Errorf("error unmarshalling %s from configmap %s/%s", patcher.SidecarDataKey, pod.GetNamespace(), configmapSidecarName)
				}
				if sidecars != nil {
					for _, sidecar := range sidecars {
						patches = append(patches, createPatches(sidecar.Containers, pod.Spec.Containers, "/spec/containers")...)
						patches = append(patches, createPatches(sidecar.Volumes, pod.Spec.Volumes, "/spec/volumes")...)
						patches = append(patches, createPatches(sidecar.ImagePullSecrets, pod.Spec.ImagePullSecrets, "/spec/imagePullSecrets")...)
					}
				}
			}
		}
	}
	return patches, nil
}

/*PatchPodUpdate not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodUpdate(_ string, _ corev1.Pod, _ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}

/*PatchPodDelete not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodDelete(_ string, _ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}
