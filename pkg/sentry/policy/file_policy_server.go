package policy

import (
	"encoding/json"
	"net/http"
)

type PolicyUpdateRequest struct {
	Path       string `json:"path"`
	Permission string `json:"permission"` // "ro", "rw", "deny"
}

func StartFilePolicyServer(mgr *FilePolicyManager, addr string) error {
	http.HandleFunc("/update_permission", func(w http.ResponseWriter, r *http.Request) {
		var req PolicyUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		perm := FilePermission(req.Permission)
		mgr.SetPermission(req.Path, perm)
		w.WriteHeader(http.StatusOK)
	})
	return http.ListenAndServe(addr, nil)
}