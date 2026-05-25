package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/tools/rag/internal/app"
	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	cli       *app.CLI
	static    http.Handler
	repoRoot  string
	indexTO   time.Duration
	requestTO time.Duration
}

type askRequest struct {
	Query   string `json:"query"`
	Backend string `json:"backend"`
	TopK    int    `json:"top_k"`
	Refresh bool   `json:"refresh"`
}

type askResponse struct {
	Answer   string             `json:"answer"`
	Backend  string             `json:"backend"`
	Results  []searchResultView `json:"results"`
	RepoRoot string             `json:"repo_root"`
}

type doctorResponse struct {
	RepoRoot           string        `json:"repo_root"`
	DBPath             string        `json:"db_path"`
	Writable           bool          `json:"writable"`
	GitReady           bool          `json:"git_ready"`
	AnthropicAuthToken bool          `json:"anthropic_auth_token"`
	AnthropicAPIKey    bool          `json:"anthropic_api_key"`
	AuthReady          bool          `json:"auth_ready"`
	IndexedSources     []sourceCount `json:"indexed_sources"`
	AllowedChecks      []checkResult `json:"allowed_checks"`
}

type sourceCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type checkResult struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type indexResponse struct {
	FilesIndexed int `json:"files_indexed"`
	FilesSkipped int `json:"files_skipped"`
	ChunkCount   int `json:"chunk_count"`
}

type searchResultView struct {
	Path      string  `json:"path"`
	LineStart int     `json:"line_start"`
	LineEnd   int     `json:"line_end"`
	Citation  string  `json:"citation"`
	Score     float64 `json:"score"`
	Snippet   string  `json:"snippet"`
}

func NewServer(repoRoot string) (*Server, error) {
	cli, err := app.NewCLI(repoRoot)
	if err != nil {
		return nil, err
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	return &Server{
		cli:       cli,
		repoRoot:  cli.RepoRoot,
		static:    http.FileServer(http.FS(sub)),
		indexTO:   10 * time.Minute,
		requestTO: 5 * time.Minute,
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", s.handleHealthz)
	mux.HandleFunc("/api/doctor", s.handleDoctor)
	mux.HandleFunc("/api/index", s.handleIndex)
	mux.HandleFunc("/api/ask", s.handleAsk)
	mux.Handle("/", s.static)
	return withJSONLogging(mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	report, err := s.cli.Doctor(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := doctorResponse{
		RepoRoot:           report.RepoRoot,
		DBPath:             report.DBPath,
		Writable:           report.Writable,
		GitReady:           report.GitReady,
		AnthropicAuthToken: report.AuthTokenReady,
		AnthropicAPIKey:    report.APIKeyReady,
		AuthReady:          report.AuthReady,
	}
	for name, count := range report.IndexedSources {
		resp.IndexedSources = append(resp.IndexedSources, sourceCount{Name: name, Count: count})
	}
	for name, err := range report.AllowedChecks {
		item := checkResult{Name: name, OK: err == nil}
		if err != nil {
			item.Detail = err.Error()
		}
		resp.AllowedChecks = append(resp.AllowedChecks, item)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.indexTO)
	defer cancel()

	result, err := s.cli.Index(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, indexResponse{
		FilesIndexed: result.FilesIndexed,
		FilesSkipped: result.FilesSkipped,
		ChunkCount:   len(result.Chunks),
	})
}

func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	client, backendName, err := loadModel(req.Backend)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errBackendAuth) {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.requestTO)
	defer cancel()
	answer, results, err := s.cli.Ask(ctx, app.AskOptions{
		Query:   req.Query,
		TopK:    req.TopK,
		Model:   client,
		Refresh: req.Refresh,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := askResponse{
		Answer:   answer,
		Backend:  backendName,
		RepoRoot: s.repoRoot,
	}
	for _, result := range results {
		resp.Results = append(resp.Results, searchResultView{
			Path:      result.Chunk.Path,
			LineStart: result.Chunk.LineStart,
			LineEnd:   result.Chunk.LineEnd,
			Citation:  result.Citation,
			Score:     result.Score,
			Snippet:   result.Chunk.Content,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

var errBackendAuth = errors.New("backend auth is not configured")

func loadModel(name string) (model.Client, string, error) {
	backend := strings.TrimSpace(name)
	switch backend {
	case "", "anthropic":
		client, err := app.NewAnthropicClient()
		if err != nil {
			return nil, "anthropic", fmt.Errorf("%w: %v", errBackendAuth, err)
		}
		return client, "anthropic", nil
	case "none":
		return nil, "none", nil
	default:
		return nil, backend, fmt.Errorf("unsupported backend: %s", backend)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withJSONLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func SearchResultsForTests(results []indexer.SearchResult) []searchResultView {
	out := make([]searchResultView, 0, len(results))
	for _, result := range results {
		out = append(out, searchResultView{
			Path:      result.Chunk.Path,
			LineStart: result.Chunk.LineStart,
			LineEnd:   result.Chunk.LineEnd,
			Citation:  result.Citation,
			Score:     result.Score,
			Snippet:   result.Chunk.Content,
		})
	}
	return out
}
