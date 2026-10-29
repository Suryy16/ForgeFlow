package authorization

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSONBody(w http.ResponseWriter, data any) error {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode authorization response: %v", err)
		return err
	}
	return nil
}
