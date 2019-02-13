package webhook

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

type Mutator struct {
}

func (mutator Mutator) Mutate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello World! You are mutated")); err != nil {
		glog.Errorf("Failed writing response: %v", err)
		http.Error(w, fmt.Sprintf("Failed writing response: %v", err), http.StatusInternalServerError)
	}
}
