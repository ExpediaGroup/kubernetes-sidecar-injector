package webhook

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
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
	sideCarNameSpace			     = "haystack-kube-sidecar-injector.expedia.com/"
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

type SideCar struct {
	Containers       []corev1.Container            `yaml:"containers"`
	Volumes          []corev1.Volume               `yaml:"volumes"`
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
}

type Mutator struct {
	SideCar *SideCar
}

func (mutator Mutator) Mutate(req []byte) ([]byte, error) {
	admissionReviewResp := v1beta1.AdmissionReview{}
	admissionReviewReq := v1beta1.AdmissionReview{}
	var admissionResponse *v1beta1.AdmissionResponse

	_, _, err := deserializer.Decode(req, nil, &admissionReviewReq)

	if err == nil && admissionReviewReq.Request != nil {
		admissionResponse = mutate(&admissionReviewReq, mutator.SideCar)
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

func mutate(ar *v1beta1.AdmissionReview, sideCar *SideCar) *v1beta1.AdmissionResponse {
	req := ar.Request

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	pod, err := unMarshall(req)
	if err != nil {
		return errorResponse(ar.Request.UID, err)
	}

	if !shouldMutate(systemNameSpaces, &pod.ObjectMeta) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{sideCarInjectionStatusAnnotation: injectedValue}
	patchBytes, err := createPatch(&pod, sideCar, annotations)
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

func shouldMutate(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skipping mutation for [%v] in special namespace: [%v]", metadata.Name, metadata.Namespace)
			return false
		}
	}

	required := false
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status := annotations[sideCarInjectionStatusAnnotation]
	if strings.ToLower(status) != injectedValue {
		switch strings.ToLower(annotations[sideCarInjectionAnnotation]) {
		case "y", "yes", "true", "on":
			required = true
		}
	}

	glog.Infof("sideCar injection for %v/%v: status: %q required:%v", metadata.Namespace, metadata.Name, status, required)
	return required
}

func createPatch(pod *corev1.Pod, inSidecarConfig *SideCar, annotations map[string]string) ([]byte, error) {
	sideCar, err := makeCopy(inSidecarConfig)
	if err != nil {
		return nil, err
	}

	// copies all annotations in the pod with sideCarNameSpace as env
	// in the injected sidecar containers
	envVariables := getEnvToIjnect(pod.Annotations)
	for key, value := range envVariables {
		for _, c := range sideCar.Containers {
			c.Env = append(c.Env, corev1.EnvVar{Name: key, Value: value})
		}
	}

	var patch []patchOperation
	patch = append(patch, addContainer(pod.Spec.Containers, sideCar.Containers, "/spec/containers")...)
	patch = append(patch, addVolume(pod.Spec.Volumes, sideCar.Volumes, "/spec/volumes")...)
	patch = append(patch, addImagePullSecrets(pod.Spec.ImagePullSecrets, sideCar.ImagePullSecrets, "/spec/imagePullSecrets")...)

	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
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

func makeCopy(src *SideCar) (*SideCar, error) {
	data, err := yaml.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst SideCar
	err = yaml.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

func getEnvToIjnect(annotations map[string]string) map[string]string {
	var env map[string]string
	env = make(map[string]string)

	sz := len(sideCarNameSpace)

	for key, value := range annotations {
		if strings.HasPrefix(key, sideCarNameSpace) && len(key) > sz {
			parts := strings.Split(key, "/")

			if parts[1] != injectAnnotation && parts[1] != statusAnnotation {
				env[parts[1]] = value
			}
		}
	}
}
