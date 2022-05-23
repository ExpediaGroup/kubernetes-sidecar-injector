package webhook

import "net/http"

// HealthCheckHandler HttpServer function to handle Health check
func HealthCheckHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
