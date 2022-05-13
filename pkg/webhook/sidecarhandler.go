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
	InitContainers   []corev1.Container            `yaml:"initContainers"`
	Containers       []corev1.Container            `yaml:"containers"`
	Volumes          []corev1.Volume               `yaml:"volumes"`
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
	Annotations      map[string]string             `yaml:"annotations"`
}

type SidecarInjectorPatcher struct {
	K8sClient                kubernetes.Interface
	InjectPrefix             string
	InjectName               string
	SidecarDataKey           string
	AllowAnnotationOverrides bool
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

func createArrayPatches[T any](newCollection []T, existingCollection []T, path string) []admission.PatchOperation {
	var patches []admission.PatchOperation
	for index, item := range newCollection {
		var value interface{}
		first := index == 0 && len(existingCollection) == 0
		if !first {
			path = path + "/-"
			value = item
		} else {
			value = []T{item}
		}
		patches = append(patches, admission.PatchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patches
}

func createAnnotationPatches(newMap map[string]string, existingMap map[string]string, override bool) []admission.PatchOperation {
	var patches []admission.PatchOperation
	for key, value := range newMap {
		if _, ok := existingMap[key]; ok && override {
			if override {
				patches = append(patches, admission.PatchOperation{
					Op:    "replace",
					Path:  "/metadata/annotations/" + key,
					Value: value,
				})
			}
		} else {
			patches = append(patches, admission.PatchOperation{
				Op:    "add",
				Path:  "/metadata/annotations",
				Value: map[string]string{key: value},
			})
		}
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
						patches = append(patches, createArrayPatches(sidecar.InitContainers, pod.Spec.InitContainers, "/spec/initContainers")...)
						patches = append(patches, createArrayPatches(sidecar.Containers, pod.Spec.Containers, "/spec/containers")...)
						patches = append(patches, createArrayPatches(sidecar.Volumes, pod.Spec.Volumes, "/spec/volumes")...)
						patches = append(patches, createArrayPatches(sidecar.ImagePullSecrets, pod.Spec.ImagePullSecrets, "/spec/imagePullSecrets")...)
						patches = append(patches, createAnnotationPatches(sidecar.Annotations, pod.Annotations, patcher.AllowAnnotationOverrides)...)
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
