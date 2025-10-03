package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type UserRequest struct {
	Name string `json:"name"`
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || idStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}
		json.NewEncoder(w).Encode(map[string]int{"user_id": id})

	case http.MethodPost:
		var req UserRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid name"}`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"created": req.Name})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}

}
