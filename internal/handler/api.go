package handler

import (
	"encoding/json"
	"net/http"
	"webshark/internal/service"
)

type InterfaceResponse struct {
	Success bool                     `json:"success"`
	Data    []service.NetworkInterface `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

type CaptureRequest struct {
	Host            string   `json:"host"`
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	Interfaces      []string `json:"interfaces"`
	BPFFilter       string   `json:"bpf_filter"`
	WiresharkFilter string   `json:"wireshark_filter"`
}

type CaptureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func getInterfaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	host := r.URL.Query().Get("host")
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if host == "" || username == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, InterfaceResponse{
			Success: false,
			Error:   "Missing required parameters",
		})
		return
	}

	interfaces, err := service.GetRemoteInterfaces(host, username, password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, InterfaceResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, InterfaceResponse{
		Success: true,
		Data:    interfaces,
	})
}

func startCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CaptureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, CaptureResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	sessionID, err := service.StartCapture(req.Host, req.Username, req.Password, req.Interfaces, req.BPFFilter, req.WiresharkFilter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, CaptureResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, CaptureResponse{
		Success: true,
		Message: sessionID,
	})
}

func stopCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, CaptureResponse{
			Success: false,
			Error:   "Missing session_id",
		})
		return
	}

	err := service.StopCapture(sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, CaptureResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, CaptureResponse{
		Success: true,
		Message: "Capture stopped",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
