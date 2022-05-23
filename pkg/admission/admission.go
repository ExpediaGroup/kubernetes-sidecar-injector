package admission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// PatchOperation JsonPatch struct http://jsonpatch.com/
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// RequestHandler AdmissionRequest handler
type RequestHandler interface {
	handleAdmissionCreate(ctx context.Context, request *admissionv1.AdmissionRequest) ([]PatchOperation, error)
	handleAdmissionUpdate(ctx context.Context, request *admissionv1.AdmissionRequest) ([]PatchOperation, error)
	handleAdmissionDelete(ctx context.Context, request *admissionv1.AdmissionRequest) ([]PatchOperation, error)
}

// Handler Generic handler for Admission
type Handler struct {
	Handler RequestHandler
}

// HandleAdmission HttpServer function to handle Admissions
func (handler *Handler) HandleAdmission(writer http.ResponseWriter, request *http.Request) {
	if err := validateRequest(request); err != nil {
		log.Error(err.Error())
		handler.writeErrorAdmissionReview(http.StatusBadRequest, err.Error(), writer)
		return
	}

	body, err := readRequestBody(request)
	if err != nil {
		log.Error(err.Error())
		handler.writeErrorAdmissionReview(http.StatusInternalServerError, err.Error(), writer)
		return
	}

	admReview := admissionv1.AdmissionReview{}

	err = json.Unmarshal(body, &admReview)
	if err != nil {
		message := fmt.Sprintf("Could not decode body: %v", err)
		log.Error(message)
		handler.writeErrorAdmissionReview(http.StatusInternalServerError, message, writer)
		return
	}

	ctx := context.Background()

	req := admReview.Request
	log.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v", req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
	if patchOperations, err := handler.Process(ctx, req); err != nil {
		message := fmt.Sprintf("request for object '%s' with name '%s' in namespace '%s' denied: %v", req.Kind.String(), req.Name, req.Namespace, err)
		log.Error(message)
		handler.writeDeniedAdmissionResponse(&admReview, message, writer)
	} else if patchBytes, err := json.Marshal(patchOperations); err != nil {
		message := fmt.Sprintf("request for object '%s' with name '%s' in namespace '%s' denied: %v", req.Kind.String(), req.Name, req.Namespace, err)
		log.Error(message)
		handler.writeDeniedAdmissionResponse(&admReview, message, writer)
	} else {
		handler.writeAllowedAdmissionReview(&admReview, patchBytes, writer)
	}
}

// Process Handles the AdmissionRequest via the handler
func (handler *Handler) Process(ctx context.Context, request *admissionv1.AdmissionRequest) ([]PatchOperation, error) {
	switch request.Operation {
	case admissionv1.Create:
		return handler.Handler.handleAdmissionCreate(ctx, request)
	case admissionv1.Update:
		return handler.Handler.handleAdmissionUpdate(ctx, request)
	case admissionv1.Delete:
		return handler.Handler.handleAdmissionDelete(ctx, request)
	default:
		return nil, fmt.Errorf("unhandled request operations type %s", request.Operation)
	}
}

func validateRequest(req *http.Request) error {
	if req.Method != http.MethodPost {
		return fmt.Errorf("wrong http verb. got %s", req.Method)
	}
	if req.Body == nil {
		return errors.New("empty body")
	}
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("wrong content type. expected 'application/json', got: '%s'", contentType)
	}
	return nil
}

func readRequestBody(req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read Request Body: %v", err)
	}
	return body, nil
}

func (handler *Handler) writeAllowedAdmissionReview(ar *admissionv1.AdmissionReview, patch []byte, res http.ResponseWriter) {
	ar.Response = handler.admissionResponse(http.StatusOK, "")
	ar.Response.Allowed = true
	ar.Response.UID = ar.Request.UID
	if patch != nil {
		pt := admissionv1.PatchTypeJSONPatch
		ar.Response.Patch = patch
		ar.Response.PatchType = &pt
	}
	handler.write(ar, res)
}

func (handler *Handler) writeDeniedAdmissionResponse(ar *admissionv1.AdmissionReview, message string, res http.ResponseWriter) {
	ar.Response = handler.admissionResponse(http.StatusForbidden, message)
	ar.Response.UID = ar.Request.UID
	handler.write(ar, res)
}

func (handler *Handler) writeErrorAdmissionReview(status int, message string, res http.ResponseWriter) {
	admResp := handler.errorAdmissionReview(status, message)
	handler.write(admResp, res)
	return
}

func (handler *Handler) errorAdmissionReview(httpErrorCode int, message string) *admissionv1.AdmissionReview {
	r := baseAdmissionReview()
	r.Response = handler.admissionResponse(httpErrorCode, message)
	return r
}

func (handler *Handler) admissionResponse(httpErrorCode int, message string) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Code:    int32(httpErrorCode),
			Message: message,
		},
	}
}

func baseAdmissionReview() *admissionv1.AdmissionReview {
	gvk := admissionv1.SchemeGroupVersion.WithKind("AdmissionReview")
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
	}
}

func (handler *Handler) write(r *admissionv1.AdmissionReview, res http.ResponseWriter) {
	resp, err := json.Marshal(r)
	if err != nil {
		log.Errorf("Error marshalling decision: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
