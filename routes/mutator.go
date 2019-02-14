package routes

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/mchandramouli/haystack-kube-sidecar-injector/webhook"
	"io/ioutil"
	"net/http"
)

func loadConfig(sideCarConfigFile string) (*webhook.SideCar, error) {
	data, err := ioutil.ReadFile(sideCarConfigFile)
	if err != nil {
		return nil, err
	}
	glog.Infof("New sideCar configuration: %s", data)

	var cfg webhook.SideCar
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type MutatorController interface {
	Mutate(http.ResponseWriter, *http.Request)
}

func NewMutatorController(sideCarConfigFile string) (MutatorController, error) {
	sideCarConfig, err := loadConfig(sideCarConfigFile)
	if sideCarConfig != nil {
		return mutatorController{mutator: webhook.Mutator{SideCar: sideCarConfig}}, nil
	}
	return nil, err
}

type mutatorController struct {
	mutator webhook.Mutator
}

func (m mutatorController) Mutate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(m.mutator.Mutate())); err != nil {
		glog.Errorf("Failed writing response: %v", err)
		http.Error(w, fmt.Sprintf("Failed writing response: %v", err), http.StatusInternalServerError)
	}
}

