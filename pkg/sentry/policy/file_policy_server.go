package policy

import (
	"encoding/json"
	"net/http"
)

type PolicyUpdateRequest struct {
	Path       string `json:"path"`
	Permission string `json:"permission"` // "ro", "rw", "deny"
}

// StartFilePolicyServer starts an HTTP server that updates the in-process
// FilePolicyManager and optionally invokes onUpdate to propagate the change
// elsewhere (e.g., to running sandboxes via RPC). onUpdate may be nil.
func StartFilePolicyServer(mgr *FilePolicyManager, addr string, onUpdate func(path, perm string)) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update_permission", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req PolicyUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		perm := FilePermission(req.Permission)
		mgr.SetPermission(req.Path, perm)
		if onUpdate != nil {
			onUpdate(req.Path, req.Permission)
		}
		w.WriteHeader(http.StatusOK)
	})
	return http.ListenAndServe(addr, mux)
}