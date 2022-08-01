// Пакет handlers содержит методы обработчиков HTTP запросов.
package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/types"
)

const (
	GAUGE   = "gauge"   // значение метрики действительное число
	COUNTER = "counter" // значение метрики целое число
)

type Handler struct {
	config     *config.Config
	repository types.MetricRepo
}

// NewHandler возвращает структуру обработчика HTTP запросов.
func NewHandler(serverConfig *config.Config, repo types.MetricRepo) Handler {
	return Handler{config: serverConfig, repository: repo}
}

// Index обрабатывает GET / запрос.
// Отображаются актуальные значения метрик.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	result := strings.Builder{}
	result.WriteString(`
	<!doctype html>
	<html lang="en">
	<head>
	  <meta charset="utf-8">
	  <title>CMAS Index Page</title>
	</head>
	<body>
	<pre>`)

	metrics, err := h.repository.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)

		return
	}

	for name, value := range metrics {
		result.WriteString(name)
		result.WriteString(" = ")

		switch value.TValue {
		case GAUGE:
			result.WriteString(strconv.FormatFloat(float64(value.GValue), 'f', -1, 64))
		case COUNTER:
			result.WriteString(strconv.Itoa(int(value.CValue)))
		}

		result.WriteString("\n")
	}

	result.WriteString(`
	</pre>
	</body>
	</html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintln(w, result.String())
}

// UpdatePlain обновляет метрику из POST /update/{type}/{name}/{value} запроса.
//
// @Tags	Metrics
// @Summary	Обновляет значение метрики
// @Param	type path string true "Тип метрики"
// @Param	name path string true "Имя метрики"
// @Param	value path number true "Значение метрики"
// @Accept	plain
// @Produce	plain
// @Success	200
// @Failure	400 {string} string
// @Failure	500 {string} string
// @Failure	501 {string} string
// @Router /update/{type}/{name}/{value} [post]
func (h *Handler) UpdatePlain(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	mValue := chi.URLParam(r, "value")

	switch mType {
	case GAUGE:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			http.Error(w, "incorrect value", http.StatusBadRequest)

			return
		}

		err = h.repository.Save(r.Context(), mName, types.Value{GValue: types.Gauge(value), TValue: "gauge"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	case COUNTER:
		value, err := strconv.Atoi(mValue)
		if err != nil {
			http.Error(w, "incorrect value", http.StatusBadRequest)

			return
		}

		err = h.repository.Save(r.Context(), mName, types.Value{CValue: types.Counter(value), TValue: "counter"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	default:
		http.Error(w, "unknown data type", http.StatusNotImplemented)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

// ValuePlain возвращает значение метрики по GET /value/{type}/{name} запросу в текстовом виде.
//
// @Tags	Metrics
// @Summary	Возвращает значение метрики
// @Param	type path string true "Тип метрики"
// @Param	name path string true "Имя метрики"
// @Accept	plain
// @Produce	plain
// @Success	200
// @Failure	404
// @Router /update/{type}/{name} [get]
func (h *Handler) ValuePlain(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	value, err := h.repository.FindByName(r.Context(), mName)
	if err != nil || value.TValue != mType {
		http.Error(w, "", http.StatusNotFound)

		return
	}

	body := ""

	switch value.TValue {
	case GAUGE:
		body = strconv.FormatFloat(float64(value.GValue), 'f', -1, 64)
	case COUNTER:
		body = strconv.Itoa(int(value.CValue))
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintln(w, body)
}

// UpdateJSON обновляет метрику из POST /update запроса.
//  Пример JSON тела запроса:
//  {"id":"PollCount","type":"counter","delta":1}
//
// @Tags	Metrics
// @Summary	Обновляет значение метрики
// @Param	value body  types.ValueJSON true "значение метрики"
// @Accept	json
// @Produce	json
// @Success	200 {object} object
// @Failure	404 {object} object
// @Failure	500 {object} object
// @Failure	501 {object} object
// @Router /update [post]
func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	valueJSON := types.ValueJSON{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&valueJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	if strings.TrimSpace(valueJSON.ID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"metric name empity\"}")

		return
	}

	if h.config.Key != "" && !checkHash(valueJSON, h.config.Key) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{\"error\":\"unknown or bad hash value\"}")

		return
	}

	switch valueJSON.MType {
	case GAUGE:
		if valueJSON.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "{\"error\":\"unknown data value\"}")

			return
		}

		err = h.repository.Save(r.Context(), valueJSON.ID, types.Value{GValue: *valueJSON.Value, TValue: "gauge"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}
	case COUNTER:
		if valueJSON.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "{\"error\":\"unknown data value\"}")

			return
		}

		err = h.repository.Save(r.Context(), valueJSON.ID, types.Value{CValue: *valueJSON.Delta, TValue: "counter"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":%q}\n", err)

			return
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "{\"error\":\"unknown data type\"}")

		return
	}

	fmt.Fprintln(w, "{}")
}

// UpdateJSONBatch обновляет множество метрик за раз по POST /updates/ запросу.
//  Пример JSON тела запроса:
//  [
//	 {"id":"PollCount","type":"counter","delta":1},
//	 {"id":"RandomValue","type":"gauge","value":321.435}
//  ]
//
// @Tags	Metrics
// @Summary	Обновляет множество значений метрик
// @Param	value body  []types.ValueJSON true "значения метрик"
// @Accept	json
// @Produce	json
// @Success	200 {object} object
// @Failure	404 {object} object
// @Failure	500 {object} object
// @Failure	501 {object} object
// @Router /updates [post]
func (h *Handler) UpdateJSONBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	valuesJSON := []types.ValueJSON{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&valuesJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	for _, valueJSON := range valuesJSON {
		if strings.TrimSpace(valueJSON.ID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "{\"error\":\"metric name empity\"}")

			return
		}

		if h.config.Key != "" && !checkHash(valueJSON, h.config.Key) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"unknown or bad hash value for %s\"}\n", valueJSON.ID)

			return
		}

		switch valueJSON.MType {
		case GAUGE:
			if valueJSON.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "{\"error\":\"unknown data value for %s\"}\n", valueJSON.ID)

				return
			}

		case COUNTER:
			if valueJSON.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "{\"error\":\"unknown data value for %s\"}\n", valueJSON.ID)

				return
			}

		default:
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "{\"error\":\"unknown data type for %s\"}\n", valueJSON.ID)

			return
		}
	}

	err = h.repository.SaveAll(r.Context(), valuesJSON)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	fmt.Fprintln(w, "{}")
}

// ValueJSON возвращает значение метрикbи по POST /value запросу в JSON виде.
//  Пример JSON тела запроса:
//  {"id":"PollCount","type":"counter"}
//
// @Tags	Metrics
// @Summary	Возвращает значение метрики
// @Param	value body  types.RequestValueJSON true "параметры метрики"
// @Accept	json
// @Produce	json
// @Success	200 {object} types.ValueJSON
// @Failure	204 {object} object
// @Failure	400 {object} object
// @Failure	404 {object} object
// @Failure	501 {object} object
// @Router /value [post]
func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	valueJSON := types.ValueJSON{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&valueJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	value, err := h.repository.FindByName(r.Context(), valueJSON.ID)
	if err != nil || value.TValue != valueJSON.MType {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"error\":\"metric not found\"}\n")

		return
	}

	switch valueJSON.MType {
	case GAUGE:
		valueJSON.Value = &value.GValue
	case COUNTER:
		valueJSON.Delta = &value.CValue
	}

	if h.config.Key != "" {
		valueJSON.Hash = calcHash(valueJSON, h.config.Key)
	}

	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)

	err = encoder.Encode(valueJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	_, err = w.Write(body.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}
}

// Ping отвечает на GET /ping запрос.
// Возвращает HTTP статус 200 Ok, если сервер работает штатно, в противном случае статус 500 Internal Server Error.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := h.repository.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"error\":%q}\n", err)

		return
	}

	fmt.Fprintln(w, "{}")
}

func checkHash(valueJSON types.ValueJSON, key string) bool {
	hash, err := hex.DecodeString(valueJSON.Hash)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(key))

	switch valueJSON.MType {
	case GAUGE:
		fmt.Fprintf(h, "%s:gauge:%f", valueJSON.ID, *valueJSON.Value)

		return hmac.Equal(h.Sum(nil), hash)
	case COUNTER:
		fmt.Fprintf(h, "%s:counter:%d", valueJSON.ID, *valueJSON.Delta)

		return hmac.Equal(h.Sum(nil), hash)
	}

	return false
}

func calcHash(valueJSON types.ValueJSON, key string) string {
	h := hmac.New(sha256.New, []byte(key))

	switch valueJSON.MType {
	case GAUGE:
		fmt.Fprintf(h, "%s:gauge:%f", valueJSON.ID, *valueJSON.Value)

		return hex.EncodeToString(h.Sum(nil))
	case COUNTER:
		fmt.Fprintf(h, "%s:counter:%d", valueJSON.ID, *valueJSON.Delta)

		return hex.EncodeToString(h.Sum(nil))
	}

	return ""
}
