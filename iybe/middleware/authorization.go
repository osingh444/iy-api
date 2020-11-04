package middleware

import (
	"net/http"
	"iybe/models"
	"iybe/utils"
)

func HasAuthorization(w http.ResponseWriter, permissionType string, claims *models.Claims) bool {
	if permissionType == claims.Level || permissionType == "all" {
		return true
	}

	utils.Respond(w, utils.Message("You are not authorized to access."), 403)
	return false
}
