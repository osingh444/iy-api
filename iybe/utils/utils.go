package utils

import (
	"time"
	"math/rand"
	"encoding/json"
	"encoding/base64"
	"net/http"

	"github.com/google/uuid"
)

func Message(message string) map[string]interface{} {
	return map[string]interface{}{"message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}, status uint) {
	w.Header().Add("Content-Type", "application/json")

	switch status {
		case 200: // break 200 Accept Request
			w.WriteHeader(http.StatusOK)
			break
		case 201: // break 201 created POST
			w.WriteHeader(http.StatusCreated)
			break
		case 204: // break 204 No Content (Just Delete Http)
			w.WriteHeader(http.StatusNoContent)
			break
		case 301: // break 301 Moved Permanently
			w.WriteHeader(http.StatusMovedPermanently)
			break
		case 400: // break 400 Bad Request
			w.WriteHeader(http.StatusBadRequest)
			break
		case 401: // break 401 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			break
		case 403: // break 403 Forbidden
			w.WriteHeader(http.StatusForbidden)
			break
		case 404: // break 404 Not Found
			w.WriteHeader(http.StatusNotFound)
			break
		case 500: // break 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			break
	}

	data["status"] = status
	json.NewEncoder(w).Encode(data)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString() string {
	b, _ := generateRandomBytes(32)
	return base64.URLEncoding.EncodeToString(b)
}

func GenerateUUID() string {
	id, _ := uuid.NewRandom()
	return id.String()
}
