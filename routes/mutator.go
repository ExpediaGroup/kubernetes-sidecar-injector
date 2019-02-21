package routes

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ExpediaDotCom/haystack-kube-sidecar-injector/webhook"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
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

func (controller mutatorController) Mutate(writer http.ResponseWriter, request *http.Request) {
	body, err := readRequestBody(request)
	if err != nil {
		writeError(writer, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := controller.mutator.Mutate(body)
	if err != nil {
		writeError(writer, fmt.Sprintf("Failed to process request: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(resp); err != nil {
		writeError(writer, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

func readRequestBody(r *http.Request) ([]byte, error) {
	var body []byte

	if r.Body != nil {
		defer r.Body.Close()
		if data, err := ioutil.ReadAll(r.Body); err != nil {
			io.Copy(ioutil.Discard, r.Body)
		} else {
			body = data
		}
	}

	if len(body) == 0 {
		return nil, errors.New("body of the request is empty")
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("received Content-Type=%s, Expected Content-Type is 'application/json'", contentType)
	}

	glog.Infof("Request received  : \n %s \n", string(body))
	return body, nil
}

func writeError(writer http.ResponseWriter, message string, status int) {
	glog.Errorf(message)
	http.Error(writer, message, status)
}
