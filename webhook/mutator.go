package webhook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

var (
	runtimeScheme    = runtime.NewScheme()
	codecs           = serializer.NewCodecFactory(runtimeScheme)
	deserializer     = codecs.UniversalDeserializer()
	systemNameSpaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

const (
	sideCarNameSpace                 = "sidecar-injector.expedia.com/"
	injectAnnotation                 = "inject"
	statusAnnotation                 = "status"
	sideCarInjectionAnnotation       = sideCarNameSpace + injectAnnotation
	sideCarInjectionStatusAnnotation = sideCarNameSpace + statusAnnotation
	injectedValue                    = "injected"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

/*SideCar is the template of the sidecar to be implemented*/
type SideCar struct {
	Containers       []corev1.Container            `yaml:"containers"`
	Volumes          []corev1.Volume               `yaml:"volumes"`
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
}

/*Mutator is the interface for mutating webhook*/
type Mutator struct {
	SideCars map[string]*SideCar
}

/*Mutate function performs the actual mutation of pod spec*/
func (mutator Mutator) Mutate(req []byte) ([]byte, error) {
	admissionReviewResp := v1beta1.AdmissionReview{}
	admissionReviewReq := v1beta1.AdmissionReview{}
	var admissionResponse *v1beta1.AdmissionResponse

	_, _, err := deserializer.Decode(req, nil, &admissionReviewReq)

	if err == nil && admissionReviewReq.Request != nil {
		admissionResponse = mutate(&admissionReviewReq, mutator.SideCars)
	} else {
		message := "Failed to decode request"

		if err != nil {
			message = fmt.Sprintf("message: %s err: %v", message, err)
		}

		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: message,
			},
		}
	}

	admissionReviewResp.Response = admissionResponse
	return json.Marshal(admissionReviewResp)
}

func mutate(ar *v1beta1.AdmissionReview, sideCars map[string]*SideCar) *v1beta1.AdmissionResponse {
	req := ar.Request

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	pod, err := unMarshall(req)
	if err != nil {
		return errorResponse(ar.Request.UID, err)
	}

	if sideCarNames, ok := shouldMutate(systemNameSpaces, &pod.ObjectMeta); ok {
		annotations := map[string]string{sideCarInjectionStatusAnnotation: injectedValue}
		patchBytes, err := createPatch(&pod, sideCarNames, sideCars, annotations)
		if err != nil {
			return errorResponse(req.UID, err)
		}

		glog.Infof("AdmissionResponse: Patch: %v\n", string(patchBytes))
		pt := v1beta1.PatchTypeJSONPatch
		return &v1beta1.AdmissionResponse{
			UID:       req.UID,
			Allowed:   true,
			Patch:     patchBytes,
			PatchType: &pt,
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	glog.Errorf("AdmissionReview failed : [%v] %s", uid, err)
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func unMarshall(req *v1beta1.AdmissionRequest) (corev1.Pod, error) {
	var pod corev1.Pod
	err := json.Unmarshal(req.Object.Raw, &pod)
	return pod, err
}

func shouldMutate(ignoredList []string, metadata *metav1.ObjectMeta) ([]string, bool) {
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skipping mutation for [%v] in special namespace: [%v]", metadata.Name, metadata.Namespace)
			return nil, false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if status, ok := annotations[sideCarInjectionStatusAnnotation]; ok && strings.ToLower(status) == injectedValue {
		glog.Infof("Skipping mutation for [%v]. Has been mutated already", metadata.Name)
		return nil, false
	}

	if sidecars, ok := annotations[sideCarInjectionAnnotation]; ok {
		parts := strings.Split(sidecars, ",")
		for i := range parts {
			parts[i] = strings.Trim(parts[i], " ")
		}

		if len(parts) > 0 {
			glog.Infof("sideCar injection for %v/%v: sidecars: %v", metadata.Namespace, metadata.Name, sidecars)
			return parts, true
		}
	}

	glog.Infof("Skipping mutation for [%v]. No action required", metadata.Name)
	return nil, false
}

func createPatch(pod *corev1.Pod, sideCarNames []string, sideCars map[string]*SideCar, annotations map[string]string) ([]byte, error) {

	var patch []patchOperation
	var containers       []corev1.Container
	var volumes          []corev1.Volume
	var imagePullSecrets []corev1.LocalObjectReference
	count := 0

	for _, name := range sideCarNames {
		if sideCar, ok := sideCars[name]; ok {
			sideCarCopy := sideCar

			// copies all annotations in the pod with sideCarNameSpace as env
			// in the injected sidecar containers
			envVariables := getEnvToInject(pod.Annotations)
			if len(envVariables) > 0 {
				for i := range sideCarCopy.Containers {
					sideCarCopy.Containers[i].Env = append(sideCarCopy.Containers[i].Env, envVariables...)
				}
			}

			containers = append(containers, sideCarCopy.Containers...)
			volumes = append(volumes, sideCarCopy.Volumes...)
			imagePullSecrets = append(imagePullSecrets, sideCarCopy.ImagePullSecrets...)

			count++
		}
	}

	if len(sideCarNames) == count {
		patch = append(patch, addContainer(pod.Spec.Containers, containers, "/spec/containers")...)
		patch = append(patch, addVolume(pod.Spec.Volumes, volumes, "/spec/volumes")...)
		patch = append(patch, addImagePullSecrets(pod.Spec.ImagePullSecrets, imagePullSecrets, "/spec/imagePullSecrets")...)
		patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

		return json.Marshal(patch)
	}

	return nil, fmt.Errorf("Did not find one or more sidecars to inject %v", sideCarNames)
}

func addContainer(target, added []corev1.Container, basePath string) []patchOperation {
	var patch []patchOperation
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func addVolume(target, added []corev1.Volume, basePath string) []patchOperation {
	var patch []patchOperation
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Volume{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func addImagePullSecrets(target, added []corev1.LocalObjectReference, basePath string) []patchOperation {
	var patch []patchOperation
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.LocalObjectReference{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func updateAnnotation(target map[string]string, added map[string]string) []patchOperation {
	var patch []patchOperation
	if target == nil {
		target = map[string]string{}
	}
	for key, value := range added {
		_, ok := target[key]
		if ok {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		} else {
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		}
	}
	return patch
}

func getEnvToInject(annotations map[string]string) []corev1.EnvVar {
	var env []corev1.EnvVar
	sz := len(sideCarNameSpace)

	for key, value := range annotations {
		if len(key) > sz && strings.HasPrefix(key, sideCarNameSpace) {
			parts := strings.Split(key, "/")

			if parts[1] != injectAnnotation && parts[1] != statusAnnotation {
				env = append(env, corev1.EnvVar{Name: parts[1], Value: value})
			}
		}
	}

	return env
}
