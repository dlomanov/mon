package main

import (
	"errors"
	"fmt"
	"github.com/dlomanov/mon/internal/handlers"
	"github.com/dlomanov/mon/internal/metrics"
	"github.com/dlomanov/mon/internal/storage"
	"net/http"
	"strings"
)

const port = "8080"

func main() {
	fmt.Printf("server started, check: http://localhost:%s/alive\r\n", port)
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	db := storage.NewStorage()

	mux := http.NewServeMux()

	mux.HandleFunc("/alive", alive)

	// /update/<type>/<name>/<value>
	mux.HandleFunc("/update/", update("/update/", db))

	return http.ListenAndServe(":"+port, mux)
}

func alive(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("alive"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func update(path string, db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, invalid := validateRequest(r)
		if invalid {
			w.WriteHeader(status)
			return
		}

		valueString := strings.TrimLeft(r.RequestURI, path)
		values := strings.Split(valueString, "/")
		if len(values) > 3 { // unknown RequestURI
			w.WriteHeader(http.StatusNotFound)
			return
		}

		raw, err := parseRawMetric(values...)
		status, invalid = validateMetric(raw, err)
		if invalid {
			w.WriteHeader(status)
			return
		}

		metric := metrics.Metric{
			Kind:  metrics.Kind(raw.kind),
			Name:  raw.name,
			Value: raw.value,
		}

		var handlerError error = nil
		switch metric.Kind {
		case metrics.KindGauge:
			handlerError = handlers.HandleGauge(metric, db)
		case metrics.KindCounter:
			handlerError = handlers.HandleCounter(metric, db)
		default:
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		if handlerError == nil {
			return
		}
		var validationError handlers.ValidationError
		if errors.As(handlerError, &validationError) {
			fmt.Printf("error: %v\r\n", validationError)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("error: %v", handlerError)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func parseRawMetric(values ...string) (metric rawMetric, err error) {
	for i, value := range values {
		switch i {
		case 0:
			metric.kind = strings.ToLower(value)
		case 1:
			metric.name = strings.ToLower(value)
		case 2:
			metric.value = value
		default:
			err = fmt.Errorf("expected 3 values, but recieved %v", len(values))
		}
	}
	return
}

func validateRequest(r *http.Request) (status int, invalid bool) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, true
	}
	if header := r.Header.Get("Content-Type"); header != "text/plain" {
		return http.StatusUnsupportedMediaType, true
	}

	return 0, false
}

func validateMetric(raw rawMetric, err error) (status int, invalid bool) {
	if err != nil {
		fmt.Println("error: invalid number of values")
		return http.StatusBadRequest, true
	}
	_, ok := metrics.ParseKind(raw.kind)
	if !ok {
		fmt.Println("error: unknown type")
		return http.StatusBadRequest, true
	}
	if raw.name == "" {
		fmt.Println("error: empty name")
		return http.StatusNotFound, true
	}
	if raw.value == "" {
		fmt.Println("error: empty value")
		return http.StatusBadRequest, true
	}

	return 0, false
}

type rawMetric struct {
	kind  string
	name  string
	value string
}
