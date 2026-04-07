package api

import (
	"encoding/json"
	"net/http"

	v1 "chaosOperator/api/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CheckResponse struct {
	Allowed   bool    `json:"allowed"`
	Remaining float64 `json:"remaining"`
	Reason    string  `json:"reason"`
}

type Server struct {
	Client client.Client
}

func (s *Server) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	var cbList v1.ChaosBudgetList
	if err := s.Client.List(ctx, &cbList); err != nil {
		l.Error(err, "failed to list ChaosBudgets")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var found *v1.ChaosBudget
	for i := range cbList.Items {
		if cbList.Items[i].Name == target {
			found = &cbList.Items[i]
			break
		}
	}

	if found == nil {
		resp := CheckResponse{
			Allowed: false,
			Reason:  "ChaosBudget not found for target",
		}
		jsonResponse(w, resp)
		return
	}

	resp := CheckResponse{
		Allowed:   found.Status.Allowed,
		Remaining: found.Status.Remaining,
		Reason:    "within budget",
	}
	if !resp.Allowed {
		resp.Reason = "budget exceeded or denied by policy"
	}

	jsonResponse(w, resp)
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
