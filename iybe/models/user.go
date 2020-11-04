package models

import (
	"strings"
	"time"
	"os"
	"fmt"

	"database/sql"

	"iybe/dbwrapper"
	"iybe/utils"
	"iybe/services"

	"golang.org/x/crypto/bcrypt"
	"github.com/joho/godotenv"
  "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type User struct {
	DisplayName string
  Email       string
  Password    string
}

type UserItem struct {
	Email        string
	DisplayName  string
	PasswordHash string
	Confirmed    bool
	ConfirmToken string
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
	ResetToken   string
	TokenExp     time.Time
	Level        string
}

type Claims struct {
	Email       string
	Level       string
	Confirmed   bool
	ID          string
	DisplayName string
	jwt.StandardClaims
}

type Cookies struct {
	Token      string
	ID         string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
}

func (user *User) Validate() (map[string]interface{}, bool) {
  user.Email = strings.ToLower(user.Email)

	if strings.TrimSpace(user.DisplayName) == ""  {
		return utils.Message("Invalid display name"), false
	}

	if len(user.Password) < 6 {
		return utils.Message("Password must be at least 6 characters long"), false
	}

	if !strings.Contains(user.Email, "@") {
		return utils.Message("Invalid email address"), false
	}

	if utils.ContainsBadWords(user.DisplayName) {
		return utils.Message("Inappropriate display name"), false
	}

	//check if user already exists
	rows, err := dbwrapper.GetDB().Query("SELECT 1 FROM users WHERE email=?", user.Email)

	if err != nil {
		return utils.Message("Server error"), false
	}

	defer rows.Close()
	var exists int

	err = DBHandleInt(rows, &exists)

	if err != nil {
		return utils.Message("Server error"), false
	}

	if exists == 1 {
		return utils.Message("User with this email already exists"), false
	}

	return nil, true
}

func (user *User) UserCreation(level string) (map[string]interface{}, uint) {
	if(level == "user") {
		if resp, ok := user.Validate(); !ok {
			return resp, 400
		}
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	item := UserItem{
		Email: user.Email,
		ConfirmToken: utils.GenerateRandomString(),
		ResetToken: " ",
		DisplayName: user.DisplayName,
		PasswordHash: string(hashedPassword),
		Confirmed: false,
		ID: utils.GenerateUUID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: time.Time{},
		TokenExp: time.Time{},
		Level: level,
	}

	stmt, err := dbwrapper.GetDB().Prepare(`INSERT INTO users (email, password, confirmToken, displayName, level,
	                                     	userID, confirmed, createdAt, updatedAt, deletedAt, tokenExp)
	                                     	VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer stmt.Close()

	if _, err = stmt.Exec(item.Email, item.PasswordHash, item.ConfirmToken, item.DisplayName, item.Level, item.ID, item.Confirmed,
	                      item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.TokenExp); err != nil {
		return utils.Message("Server error"), 500
	}

	_ = services.SendEmailConfirmationEmail()

	response := utils.Message("User created")
	//need to delete this line, going to send with email once email setup
	response["confirm"] = item.ConfirmToken

	return response, 201
}

func Login(email string, password string) (*Cookies, map[string]interface{}, uint) {
	row, err := dbwrapper.GetDB().Query("SELECT * FROM users WHERE email = ?", strings.ToLower(email))

	if err != nil {
		return nil, utils.Message("Server error"), 500
	}

	defer row.Close()

	users := make([]UserItem, 0)
	err = DBHandleUser(row, &users)
	if err != nil {
		fmt.Println(err)
		return nil, utils.Message("Server error"), 500
	}

	if len(users) == 0 {
		return nil, utils.Message("User does not exist"), 404
	}
	item := users[0]

	if err = bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(password)); err != nil {
		return nil, utils.Message("Email or password incorrect"), 401
	}

	//make the token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: email,
		Level: item.Level,
		Confirmed: item.Confirmed,
		ID: item.ID,
		DisplayName: item.DisplayName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("token_password")))
	if err != nil {
		return nil, utils.Message("Server error"), 500
	}

	response :=
		map[string]interface{}{
			"message":   "Logged In",
			"Expires":   expirationTime,
		}

	cookies := &Cookies{
		Token: tokenString,
		ID:    item.ID,
	}

	return cookies, response, 200
}

func ConfirmEmail(token string) (map[string]interface{}, uint) {
	stmt, err := dbwrapper.GetDB().Prepare("UPDATE users SET confirmed = ?, updatedAt = ? WHERE confirmToken = ?")
	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer stmt.Close()

	if _, err = stmt.Exec(true, time.Now(), token); err != nil {
		return utils.Message("Server error"), 500
	}

	return utils.Message("Email confirmed"), 200
}

func UserExists(email string) bool {
	rows, err := dbwrapper.GetDB().Query("SELECT 1 FROM users WHERE email=?", email)

	if err != nil {
		return false
	}

	defer rows.Close()
	var exists int

	err = DBHandleInt(rows, &exists)
	if err != nil {
		return false
	}

	return exists == 1
}

func SetupPasswordReset(email string) (map[string]interface{}, uint) {
	email = strings.ToLower(email)
	if !UserExists(email) {
		return utils.Message("No user found"), 400
	}

	resetToken := utils.GenerateRandomString()

	stmt, err := dbwrapper.GetDB().Prepare("UPDATE users SET resetToken = ?, updatedAt = ?, tokenExp = ? WHERE email = ?")
	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer stmt.Close()

	expirationTime := time.Now().Add(24 * time.Hour)
	if _, err = stmt.Exec(resetToken, time.Now(), expirationTime, email); err != nil {
		return utils.Message("Server error"), 500
	}

	if err = services.SendPasswordResetEmail(email, resetToken); err != nil {
		return utils.Message("Server error"), 500
	}

	return utils.Message("Password reset email sent"), 200
}

func IsExpired(timeToCheck time.Time, expiration time.Time) bool {
	return timeToCheck.Before(expiration)
}

func ResetPassword(token string, newPassword string) (map[string]interface{}, uint) {
	row, err := dbwrapper.GetDB().Query("SELECT * FROM users WHERE resetToken = ?", token)

	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer row.Close()

	users := make([]UserItem, 0)
	err = DBHandleUser(row, &users)
	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	if len(users) == 0 {
		return utils.Message("User does not exist"), 404
	}
	user := users[0]

	if IsExpired(time.Now(), user.TokenExp) {
		return utils.Message("This link is expired"), 400
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if string(hashedPassword) == user.PasswordHash {
		return utils.Message("New password cannot be the same"), 400
	}

	newPasswordHash := string(hashedPassword)

	stmt, err := dbwrapper.GetDB().Prepare("UPDATE users SET password = ?, resetToken = ?, updatedAt = ?, tokenExp = ? WHERE resetToken = ?")
	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer stmt.Close()

	if _, err = stmt.Exec(newPasswordHash, " ", time.Now(), time.Time{}, token); err != nil {
		return utils.Message("Server error"), 500
	}

	return utils.Message("Password changed"), 200
}

func Seed() bool {
	rows, err := dbwrapper.GetDB().Query("SELECT 1 FROM users WHERE email=?", "admin")

	if err != nil {
		panic(err.Error())
	}

	defer rows.Close()
	var exists int

	err = DBHandleInt(rows, &exists)
	if err != nil {
		panic(err.Error())
	}

	if exists == 1 {
		return true
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("abc123"), bcrypt.DefaultCost)
	id, _ := uuid.NewRandom()

	user := UserItem{
		Email: "admin",
		ConfirmToken: "",
		ResetToken: "admin",
		DisplayName: "admin",
		PasswordHash: string(hashedPassword),
		ID: id.String(),
		Confirmed: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TokenExp: time.Time{},
		Level: "admin",
	}

	stmt, err := dbwrapper.GetDB().Prepare(`INSERT INTO users (email, password, confirmToken, displayName, level,
	                                     	userID, confirmed, createdAt, updatedAt, deletedAt, tokenExp)
	                                     	VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		panic(err.Error())
	}

	defer stmt.Close()

	if _, err = stmt.Exec(user.Email, user.PasswordHash, user.ConfirmToken, user.DisplayName, user.Level, user.ID, user.Confirmed,
	                      user.CreatedAt, user.UpdatedAt, user.DeletedAt, user.TokenExp); err != nil {
		panic(err.Error())
	}

	return true
}

func DBHandleInt(rows *sql.Rows, num *int) (error) {
	for rows.Next() {
		err := rows.Scan(num)
		if err != nil {
			return err
		}
	}
	return nil
}

func DBHandleUser(rows *sql.Rows, arr *[]UserItem) (error) {
	for rows.Next() {
		var item UserItem
		err := rows.Scan(&item.Email, &item.PasswordHash, &item.DisplayName, &item.Level,
		                 &item.Confirmed, &item.ID, &item.CreatedAt, &item.UpdatedAt,
										 &item.DeletedAt, &item.ConfirmToken, &item.ResetToken, &item.TokenExp)
		if err != nil {
			return err
		}
		*arr = append(*arr, item)
	}
	return nil
}
