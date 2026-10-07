package httpapi

import (
	"net/http"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/status"
)

func (s *Server) handleApiaryStatus(w http.ResponseWriter, r *http.Request) {
	asOf, err := parseOptionalTime("asOf", r.URL.Query().Get("asOf"))
	if err != nil {
		writeError(w, err)
		return
	}
	window, err := parseOptionalInspectionWindow("inspectionWindowDays", r.URL.Query().Get("inspectionWindowDays"))
	if err != nil {
		writeError(w, err)
		return
	}
	report, err := s.deps.Status.ApiaryStatus(r.PathValue("apiaryId"), status.Query{AsOf: asOf, Window: window})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, domain.NotFound("路径", r.URL.Path))
}
