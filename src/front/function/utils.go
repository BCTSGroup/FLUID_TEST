package function

import (
	"encoding/json"
	"net/http"
)

type ResponseInfo struct {
	Message string 		`json:"message"`
	Payload interface{}	`json:"payload"`
}

func RespondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	// utils.Log.Infof("Response: %s", string(response))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
