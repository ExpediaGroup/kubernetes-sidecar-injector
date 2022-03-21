package webhook

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
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
	sideCars map[string]*SideCar
}

func NewSidecarInjectorPatcher(sideCarConfigFile string) (*SidecarInjectorPatcher, error) {
	mapOfSideCars, err := loadConfig(sideCarConfigFile)
	if mapOfSideCars != nil {
		return &SidecarInjectorPatcher{sideCars: mapOfSideCars}, nil
	}
	return nil, err
}

func (patcher *SidecarInjectorPatcher) PatchPodCreate(pod corev1.Pod) ([]admission.PatchOperation, error) {
	if sideCarNames, ok := shouldMutate(systemNameSpaces, &pod.ObjectMeta); ok {
		annotations := map[string]string{sideCarInjectionStatusAnnotation: injectedValue}
		return CreateRawPatch(&pod, sideCarNames, patcher.sideCars, annotations)
	}
	return nil, nil
}

func loadConfig(sideCarConfigFile string) (map[string]*SideCar, error) {
	data, err := ioutil.ReadFile(sideCarConfigFile)
	if err != nil {
		return nil, err
	}
	log.Infof("New sideCar configuration: %s", data)

	var cfg SideCars
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	mapOfSideCar := make(map[string]*SideCar)
	for _, configuration := range cfg.Sidecars {
		mapOfSideCar[configuration.Name] = &configuration.Sidecar
	}

	return mapOfSideCar, nil
}

/*PatchPodUpdate not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodUpdate(_ corev1.Pod, _ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}

/*PatchPodDelete not supported, only support create */
func (patcher *SidecarInjectorPatcher) PatchPodDelete(_ corev1.Pod) ([]admission.PatchOperation, error) {
	return nil, nil
}
