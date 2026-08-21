package httpapi

import (
	"encoding/json"
	"errors"
	"example.com/task149/tsnsched/internal/metrics"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/service"
	"example.com/task149/tsnsched/internal/webui"
	"io"
	"net/http"
	"strconv"
)

type Server struct {
	Service *service.Service
	Metrics *metrics.Counters
	Page    string
	mux     *http.ServeMux
}

func New(s *service.Service, m *metrics.Counters) *Server {
	h := &Server{Service: s, Metrics: m, Page: webui.Page, mux: http.NewServeMux()}
	h.routes()
	return h
}
func (h *Server) Handler() http.Handler { return h.mux }

func (h *Server) routes() {
	h.mux.HandleFunc("/", h.page)
	h.mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	h.mux.HandleFunc("/healthz", h.health)
	h.mux.HandleFunc("/readyz", h.ready)
	h.mux.HandleFunc("/version", h.version)
	h.mux.HandleFunc("/metrics", h.metrics)
	h.mux.HandleFunc("/api/nodes", h.nodes)
	h.mux.HandleFunc("/api/ports", h.ports)
	h.mux.HandleFunc("/api/links", h.links)
	h.mux.HandleFunc("/api/streams", h.streams)
	h.mux.HandleFunc("/api/topology", h.topology)
	h.mux.HandleFunc("/api/drafts", h.drafts)
	h.mux.HandleFunc("/api/drafts/create", h.createDraft)
	h.mux.HandleFunc("/api/drafts/validate", h.validate)
	h.mux.HandleFunc("/api/drafts/summary", h.summary)
	h.mux.HandleFunc("/api/drafts/allocations", h.allocations)
	h.mux.HandleFunc("/api/drafts/violations", h.violations)
	h.mux.HandleFunc("/api/drafts/export", h.export)
	h.mux.HandleFunc("/api/drafts/compare", h.compare)
	h.mux.HandleFunc("/api/active", h.active)
	h.mux.HandleFunc("/api/commit", h.commit)
	h.mux.HandleFunc("/api/rollback", h.rollback)
	h.mux.HandleFunc("/api/window", h.window)
	h.mux.HandleFunc("/api/reload", h.reload)
	h.mux.HandleFunc("/api/stats", h.stats)
	h.mux.HandleFunc("/api/capacity", h.capacity)
	h.mux.HandleFunc("/api/audit", h.audit)
	h.mux.HandleFunc("/api/history", h.history)
}

func (h *Server) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, h.Page)
}
func (h *Server) health(w http.ResponseWriter, r *http.Request) {
	h.write(w, http.StatusOK, h.Service.Health(r.Context()))
}
func (h *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.Data.Ping(r.Context()); err != nil {
		h.writeErr(w, http.StatusServiceUnavailable, err)
		return
	}
	h.write(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (h *Server) version(w http.ResponseWriter, r *http.Request) {
	h.write(w, http.StatusOK, map[string]string{"service": "tsnsched", "version": "149.1"})
}
func (h *Server) metrics(w http.ResponseWriter, r *http.Request) {
	h.write(w, http.StatusOK, h.Metrics.Snapshot())
}

func (h *Server) nodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.Service.Data.Nodes(r.Context())
		h.result(w, v, e)
	case http.MethodPost:
		var v model.Node
		if e := h.read(r, &v); e != nil {
			h.writeErr(w, 400, e)
			return
		}
		e := h.Service.AddNode(r.Context(), v)
		h.result(w, map[string]string{"status": "saved"}, e)
	default:
		h.method(w)
	}
}
func (h *Server) ports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.Service.Data.Ports(r.Context())
		h.result(w, v, e)
	case http.MethodPost:
		var v model.Port
		if e := h.read(r, &v); e != nil {
			h.writeErr(w, 400, e)
			return
		}
		e := h.Service.AddPort(r.Context(), v)
		h.result(w, map[string]string{"status": "saved"}, e)
	default:
		h.method(w)
	}
}
func (h *Server) links(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.Service.Data.Links(r.Context())
		h.result(w, v, e)
	case http.MethodPost:
		var v model.Link
		if e := h.read(r, &v); e != nil {
			h.writeErr(w, 400, e)
			return
		}
		e := h.Service.AddLink(r.Context(), v)
		h.result(w, map[string]string{"status": "saved"}, e)
	default:
		h.method(w)
	}
}
func (h *Server) streams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.Service.Streams(r.Context())
		h.result(w, v, e)
	case http.MethodPost:
		var v model.Stream
		if e := h.read(r, &v); e != nil {
			h.writeErr(w, 400, e)
			return
		}
		e := h.Service.AddStream(r.Context(), v)
		h.result(w, map[string]string{"status": "saved"}, e)
	default:
		h.method(w)
	}
}
func (h *Server) topology(w http.ResponseWriter, r *http.Request) {
	g := h.Service.GraphSnapshot()
	streams, e := h.Service.Streams(r.Context())
	if e != nil {
		h.writeErr(w, 500, e)
		return
	}
	h.write(w, 200, g.Snapshot(streams))
}
func (h *Server) drafts(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Drafts(r.Context())
	h.result(w, v, e)
}

func (h *Server) createDraft(w http.ResponseWriter, r *http.Request) {
	network := r.URL.Query().Get("network")
	if network == "" {
		network = "default"
	}
	version, _ := strconv.ParseInt(r.URL.Query().Get("version"), 10, 64)
	if version == 0 {
		version = 1
	}
	v, e := h.Service.CreateDraft(r.Context(), network, version)
	if e == nil {
		h.Metrics.DraftCreated()
	}
	h.result(w, v, e)
}
func (h *Server) validate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	v, e := h.Service.Validate(r.Context(), id)
	h.result(w, v, e)
}
func (h *Server) summary(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Summary(r.Context(), r.URL.Query().Get("id"))
	h.result(w, v, e)
}
func (h *Server) allocations(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Data.Allocations(r.Context(), r.URL.Query().Get("id"))
	h.result(w, v, e)
}
func (h *Server) violations(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Data.Violations(r.Context(), r.URL.Query().Get("id"))
	h.result(w, v, e)
}
func (h *Server) export(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Data.Export(r.Context(), r.URL.Query().Get("id"))
	if e != nil {
		h.writeErr(w, 500, e)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(v)
}
func (h *Server) compare(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Compare(r.Context(), r.URL.Query().Get("left"), r.URL.Query().Get("right"))
	h.result(w, v, e)
}
func (h *Server) active(w http.ResponseWriter, r *http.Request) {
	network := r.URL.Query().Get("network")
	v, e := h.Service.Active(r.Context(), network)
	h.result(w, v, e)
}
func (h *Server) commit(w http.ResponseWriter, r *http.Request) {
	network := r.URL.Query().Get("network")
	id := r.URL.Query().Get("id")
	e := h.Service.Commit(r.Context(), network, id)
	if e == nil {
		h.Metrics.Committed()
	}
	h.result(w, map[string]string{"status": "committed"}, e)
}
func (h *Server) rollback(w http.ResponseWriter, r *http.Request) {
	e := h.Service.Rollback(r.Context(), r.URL.Query().Get("network"))
	h.result(w, map[string]string{"status": "rolled_back"}, e)
}
func (h *Server) window(w http.ResponseWriter, r *http.Request) {
	from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	v, e := h.Service.Window(r.Context(), r.URL.Query().Get("id"), from, to)
	h.result(w, v, e)
}
func (h *Server) reload(w http.ResponseWriter, r *http.Request) {
	e := h.Service.Reload(r.Context())
	h.result(w, map[string]string{"status": "reloaded"}, e)
}
func (h *Server) stats(w http.ResponseWriter, r *http.Request) {
	h.write(w, 200, h.Service.Stats(r.Context()))
}
func (h *Server) capacity(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Capacity(r.Context(), r.URL.Query().Get("id"))
	h.result(w, v, e)
}
func (h *Server) audit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	v, e := h.Service.Audit(r.Context(), limit)
	h.result(w, v, e)
}
func (h *Server) history(w http.ResponseWriter, r *http.Request) {
	v, e := h.Service.Data.DraftHistory(r.Context(), r.URL.Query().Get("network"))
	h.result(w, v, e)
}

func (h *Server) read(r *http.Request, dst any) error {
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(dst)
}
func (h *Server) result(w http.ResponseWriter, v any, err error) {
	if err != nil {
		h.writeErr(w, status(err), err)
		return
	}
	h.write(w, 200, v)
}
func (h *Server) write(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func (h *Server) writeErr(w http.ResponseWriter, code int, err error) {
	h.Metrics.Request(false)
	h.write(w, code, map[string]string{"error": err.Error()})
}
func (h *Server) method(w http.ResponseWriter) { w.WriteHeader(http.StatusMethodNotAllowed) }
func status(err error) int {
	if errors.Is(err, model.ErrNotFound) {
		return 404
	}
	if errors.Is(err, model.ErrInvalid) {
		return 400
	}
	if errors.Is(err, model.ErrNotReady) {
		return 409
	}
	return 500
}
