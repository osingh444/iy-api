package controllers

import (
	"iybe/models"
	"iybe/utils"
	"iybe/middleware"

  "encoding/json"
	"net/http"
	"strings"
	"time"
	"fmt"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Invalid request"), 400)
		return
	}

	resp, status := user.UserCreation("user")
	utils.Respond(w, resp, status)
}

func CreateMod(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "admin")
	if !ok {
		return
	}

	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Invalid request"), 400)
		return
	}

	resp, status := user.UserCreation("mod")
	utils.Respond(w, resp, status)
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err :=json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	cookies, resp, status := models.Login(strings.ToLower(user.Email), user.Password)

	if cookies == nil {
		utils.Respond(w, resp, status)
		return
	}
	//add secure to token later
	http.SetCookie(w, &http.Cookie{Name: "id", Value: cookies.ID, HttpOnly: false, Path: "/", Expires: time.Now().Add(24 * time.Hour),})
	http.SetCookie(w, &http.Cookie{Name: "token", Value: cookies.Token, Expires: time.Now().Add(24 * time.Hour),})
	http.SetCookie(w, &http.Cookie{Name: "token_set", Value: "true", HttpOnly: false, Path: "/", Expires: time.Now().Add(24 * time.Hour),})
	utils.Respond(w, resp, status)
}

func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	confirmToken := r.URL.Query()["token"]
	if len(confirmToken) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	token := confirmToken[0]

	if token == "" {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	resp, status := models.ConfirmEmail(token)
	utils.Respond(w, resp, status)
}

func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	eemail := r.URL.Query()["email"]
	if len(eemail) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	email := eemail[0]

	if email == "" {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	resp, status := models.SetupPasswordReset(email)
	utils.Respond(w, resp, status)
}

func UpdateUserPassword(w http.ResponseWriter, r *http.Request) {



}
