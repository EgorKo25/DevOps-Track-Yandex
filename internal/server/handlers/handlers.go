package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/EgorKo25/DevOps-Track-Yandex/internal/middleware"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage    *storage.MetricStorage
	compressor *middleware.Compressor
}

// NewHandler handler type constructor
func NewHandler(storage *storage.MetricStorage, compressor *middleware.Compressor) *Handler {
	return &Handler{
		storage:    storage,
		compressor: compressor,
	}
}

// GetValueStat a handler that returns the value of a specific metric
func (h Handler) GetValueStat(w http.ResponseWriter, r *http.Request) {
	res := h.storage.StatStatusM(chi.URLParam(r, "name"), chi.URLParam(r, "type"))
	if res == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := w.Write([]byte(fmt.Sprintf("%v\n", res)))
	if err != nil {
		log.Printf("Cannot write reqeust: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	return

}

// GetJSONValue go dock
func (h Handler) GetJSONValue(w http.ResponseWriter, r *http.Request) {

	var metric storage.Metric

	b, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(b, &metric); err != nil {
		fmt.Printf("Unmarshal went wrong:  %s\n", err)
	}

	stat := h.storage.StatStatusM(metric.ID, metric.MType)

	switch metric.MType {

	case "gauge":
		if stat == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if stat != nil {

			tmp := stat.(storage.Gauge)

			metric.Value = &tmp
			metric.Delta = nil

		}
	case "counter":
		if stat == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if stat != nil {

			tmp := stat.(storage.Counter)

			metric.Delta = &tmp
			metric.Value = nil

		}
	}

	dataJSON, err := json.Marshal(metric)
	if err != nil {
		log.Println("Failed to serialize")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept-Encoding") == "gzip" {

		dataJSON, err = h.compressor.Compress(dataJSON)
		if err != nil {
			log.Println("Failed to compress")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Encoding", "gzip")
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dataJSON)
	return

}

// SetJSONValue go dock
func (h Handler) SetJSONValue(w http.ResponseWriter, r *http.Request) {

	var metric storage.Metric

	b, _ := io.ReadAll(r.Body)

	if r.Header.Get("Content-Encoding") == "gzip" {
		b, _ = h.compressor.Decompress(b)
	}

	if err := json.Unmarshal(b, &metric); err != nil {
		fmt.Printf("Unmarshal went wrong:  %s\n", err)
	}

	if metric.Value == nil && metric.Delta == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case "gauge":

		if metric.Value == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if metric.Value != nil {
			h.storage.SetStat(&metric)
		}

		if stat := h.storage.StatStatusM(metric.ID, metric.MType); stat != nil {
			tmp := stat.(storage.Gauge)
			metric.Value = &tmp
		}

	case "counter":

		if metric.Delta == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if metric.Delta != nil {
			h.storage.SetStat(&metric)
		}

		if stat := h.storage.StatStatusM(metric.ID, metric.MType); stat != nil && stat.(storage.Counter) != 0 {
			tmp := stat.(storage.Counter)
			metric.Delta = &tmp
		}

	}

	dataJSON, err := json.Marshal(metric)
	if err != nil {
		log.Println("Failed to serialize")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept-Encoding") == "gzip" {

		dataJSON, err = h.compressor.Compress(dataJSON)
		if err != nil {
			log.Println("Failed to middleware")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Encoding", "gzip")
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(dataJSON)
	return

}

// GetAllStats returns the values of all metrics
func (h Handler) GetAllStats(w http.ResponseWriter, r *http.Request) {

	var res string
	var err error

	for k, v := range h.storage.Metrics {

		if v.MType == "gauge" {
			res += "> " + k + ":  " + fmt.Sprintf("%f", *v.Value) + "\n"
		}
		if v.MType == "counter" {
			res += "> " + k + ":  " + fmt.Sprintf("%d", *v.Delta) + "\n"
		}

	}

	tmp := []byte(res)

	if r.Header.Get("Accept-Encoding") == "gzip" {
		tmp, err = h.compressor.Compress(tmp)
		if err != nil {
			log.Println("Failed to compress")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Encoding", "gzip")
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(tmp)
}

// SetMetricValue sets the value of the specified metric
func (h Handler) SetMetricValue(w http.ResponseWriter, r *http.Request) {

	var metric storage.Metric

	if mType := chi.URLParam(r, "type"); mType == "gauge" {

		tmp, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Somethings went wrong: %s", err)
			return
		}
		value := storage.Gauge(tmp)

		metric.ID = chi.URLParam(r, "name")
		metric.MType = mType
		metric.Value = &value
		metric.Delta = nil

		h.storage.SetStat(&metric)
		w.WriteHeader(http.StatusOK)
	}

	if mType := chi.URLParam(r, "type"); mType == "counter" {

		tmp, err := strconv.Atoi(chi.URLParam(r, "value"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Somethings went wrong: %s", err)
			return
		}
		value := storage.Counter(tmp)

		metric.ID = chi.URLParam(r, "name")
		metric.MType = mType
		metric.Value = nil
		metric.Delta = &value

		h.storage.SetStat(&metric)
		w.WriteHeader(http.StatusOK)
	}

	w.WriteHeader(http.StatusNotImplemented)
	return

}
