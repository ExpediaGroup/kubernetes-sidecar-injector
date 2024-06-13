package pkg

import (
	"context"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// Sidecar Kubernetes Sidecar Injector schema
type Sidecar struct {
	Name             string                        `yaml:"name"`
	InitContainers   []corev1.Container            `yaml:"initContainers"`
	Containers       []corev1.Container            `yaml:"containers"`
	Volumes          []corev1.Volume               `yaml:"volumes"`
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
	Annotations      map[string]string             `yaml:"annotations"`
	Labels           map[string]string             `yaml:"labels"`
}

type SidecarInjector struct {
	decoder                  *admission.Decoder
	client                   *kubernetes.Clientset
	InjectPrefix             string
	InjectName               string
	SidecarDataKey           string
	AllowAnnotationOverrides bool
	AllowLabelOverrides      bool
}

func (sidecarInjector *SidecarInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create:
		return sidecarInjector.handleCreate(ctx, req)
	case admissionv1.Update:
		return sidecarInjector.handleUpdate(ctx, req)
	case admissionv1.Delete:
		return sidecarInjector.handleDelete(ctx, req)
	default:
		return admission.Allowed("operation not supported")
	}
}

func (sidecarInjector *SidecarInjector) InjectDecoder(d *admission.Decoder) error {
	sidecarInjector.decoder = d
	return nil
}

func (sidecarInjector *SidecarInjector) InjectClient(c *kubernetes.Clientset) error {
	sidecarInjector.client = c
	return nil
}

func (sidecarInjector *SidecarInjector) sideCarInjectionAnnotation() string {
	return sidecarInjector.InjectPrefix + "/" + sidecarInjector.InjectName
}

func (sidecarInjector *SidecarInjector) configmapSidecarNames(pod *corev1.Pod) []string {
	annotations := map[string]string{}
	if pod.Annotations != nil {
		annotations = pod.Annotations
	}
	if sidecars, ok := annotations[sidecarInjector.sideCarInjectionAnnotation()]; ok {
		parts := lo.Map[string, string](strings.Split(sidecars, ","), func(part string, _ int) string {
			return strings.TrimSpace(part)
		})

		if len(parts) > 0 {
			log.Infof("sideCar injection for %v/%v: sidecars: %v", pod.Namespace, pod.Name, sidecars)
			return parts
		}
	}
	log.Infof("Skipping mutation for [%v]. No action required", pod.GetName())
	return nil
}

// HandleCreate Handle Create Request
func (sidecarInjector *SidecarInjector) handleCreate(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	if err := sidecarInjector.decoder.Decode(req, pod); err != nil {
		return webhook.Errored(http.StatusBadRequest, err)
	}
	if configmapSidecarNames := sidecarInjector.configmapSidecarNames(pod); configmapSidecarNames != nil {
		for _, configmapSidecarName := range configmapSidecarNames {
			configmapSidecar, err := sidecarInjector.client.CoreV1().ConfigMaps(pod.Namespace).Get(ctx, configmapSidecarName, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				log.Warnf("sidecar configmap %s/%s was not found", pod.Namespace, configmapSidecarName)
			} else if err != nil {
				log.Errorf("error fetching sidecar configmap %s/%s - %v", pod.Namespace, configmapSidecarName, err)
			} else if sidecarsStr, ok := configmapSidecar.Data[sidecarInjector.SidecarDataKey]; ok {
				var sidecars []Sidecar
				if err := yaml.Unmarshal([]byte(sidecarsStr), &sidecars); err != nil {
					log.Errorf("error unmarshalling %s from configmap %s/%s", sidecarInjector.SidecarDataKey, pod.GetNamespace(), configmapSidecarName)
				}
				if sidecars != nil {
					for _, sidecar := range sidecars {
						pod.Spec.InitContainers = append(pod.Spec.InitContainers, sidecar.InitContainers...)
						pod.Spec.Containers = append(pod.Spec.Containers, sidecar.Containers...)
						pod.Spec.Volumes = append(pod.Spec.Volumes, sidecar.Volumes...)
						pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, sidecar.ImagePullSecrets...)
						for k, v := range sidecar.Annotations {
							if _, exists := pod.Annotations[k]; exists {
								if sidecarInjector.AllowAnnotationOverrides {
									pod.Annotations[k] = v
								}
							} else {
								pod.Annotations[k] = v
							}
						}
						for k, v := range sidecar.Labels {
							if _, exists := pod.Labels[k]; exists {
								if sidecarInjector.AllowLabelOverrides {
									pod.Labels[k] = v
								}
							} else {
								pod.Labels[k] = v
							}
						}
					}
				}
			}
		}
	}
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	patchResponse := admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
	log.Debugf("sidecar patches being applied for %v/%v: patches: %v", pod.Namespace, pod.Name, patchResponse.Patches)
	return patchResponse
}

/*HandleUpdate not supported, only support create */
func (sidecarInjector *SidecarInjector) handleUpdate(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("update handled")
}

/*HandleDelete not supported, only support create */
func (sidecarInjector *SidecarInjector) handleDelete(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("delete handled")
}
