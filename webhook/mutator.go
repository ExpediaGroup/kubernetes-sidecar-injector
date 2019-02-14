package webhook

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme    = runtime.NewScheme()
	codecs           = serializer.NewCodecFactory(runtimeScheme)
	deserializer     = codecs.UniversalDeserializer()
	systemNameSpaces = []string {
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

const (
	sideCarInjectionKey = "haystack-kube-sidecar-injector.expedia.com/inject"
	sideCarInjectionStatusKey = "haystack-kube-sidecar-injector.expedia.com/status"
)

type SideCar struct {
	Containers  []corev1.Container  `yaml:"containers"`
	Volumes     []corev1.Volume     `yaml:"volumes"`
}

type Mutator struct {
	SideCar *SideCar
}

func (mutator Mutator) Mutate() string {
	return "Hello World! You are mutated"
}


