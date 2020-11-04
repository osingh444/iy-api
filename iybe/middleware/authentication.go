package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"iybe/models"
	"iybe/utils"
	"fmt"
	"os"
	"net/http"
)

var tk = &models.Claims{}

func AuthenticateToken(w http.ResponseWriter, r *http.Request, level string) (*models.Claims, bool) {
	cookie, err := r.Cookie("token")

	if err != nil || cookie.Value == ""{
		utils.Respond(w, utils.Message("Missing token"), 403)
		return nil, false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("token_password")), nil
	})

	if err != nil { //Malformed token
		fmt.Println(err)
		utils.Respond(w, utils.Message("Malformed or Expired authentication token"), 403)
		return nil, false
	}

	if !token.Valid {
		utils.Respond(w, utils.Message("Token is not valid."), 403)
		return nil, false
	}

	if ok := HasAuthorization(w, level, tk); !ok {
		return nil, false
	}

	return tk, true
}
