package stats

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/falconfan123/Go-mall/services/search/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(cfg config.PrometheusExtConf, service *Service) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:    cfg.Host + ":" + strconv.Itoa(cfg.Port),
			Handler: NewHTTPHandler(cfg.Path, service, prometheus.DefaultGatherer),
		},
	}
}

func NewHTTPHandler(metricsPath string, service *Service, gatherer prometheus.Gatherer) http.Handler {
	mux := http.NewServeMux()
	path := metricsPath
	if path == "" {
		path = "/metrics"
	}
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}
	mux.Handle(path, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	mux.HandleFunc("/stats/rag/overview", makeJSONHandler(func(r *http.Request) (any, error) {
		return service.GetOverview(r.Context(), parseFilter(r))
	}))
	mux.HandleFunc("/stats/rag/trend", makeJSONHandler(func(r *http.Request) (any, error) {
		return service.GetTrend(r.Context(), parseFilter(r), r.URL.Query().Get("granularity"))
	}))
	mux.HandleFunc("/stats/rag/breakdown", makeJSONHandler(func(r *http.Request) (any, error) {
		return service.GetBreakdown(r.Context(), parseFilter(r), r.URL.Query().Get("dimension"))
	}))
	mux.HandleFunc("/stats/rag/events", makeJSONHandler(func(r *http.Request) (any, error) {
		return service.ListEvents(r.Context(), parseFilter(r))
	}))
	return mux
}

func (s *HTTPServer) Start() error {
	if s == nil || s.server == nil {
		return nil
	}
	go func() {
		_ = s.server.ListenAndServe()
	}()
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func makeJSONHandler(fn func(*http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := fn(r)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseFilter(r *http.Request) EventFilter {
	query := r.URL.Query()
	filter := EventFilter{
		AppID:           query.Get("app_id"),
		KnowledgeBaseID: query.Get("knowledge_base_id"),
		Channel:         query.Get("channel"),
		EventType:       query.Get("event_type"),
		Status:          query.Get("status"),
		Limit:           parseInt(query.Get("limit"), 100),
		Offset:          parseInt(query.Get("offset"), 0),
	}
	if startAt := parseTime(query.Get("start_at")); startAt != nil {
		filter.StartAt = startAt
	}
	if endAt := parseTime(query.Get("end_at")); endAt != nil {
		filter.EndAt = endAt
	}
	if rawIsRAG := query.Get("is_rag"); rawIsRAG != "" {
		value := rawIsRAG == "true" || rawIsRAG == "1"
		filter.IsRAG = &value
	}
	return filter
}

func parseTime(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
