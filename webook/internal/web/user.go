package web

import (
	"log"
	"net/http"

	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
}

func NewUserHandler() *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#\$%\^&\*]).{8,}$`
	)
	emailExp := regexp2.MustCompile(emailRegexPattern, regexp2.None)
	passwordExp := regexp2.MustCompile(passwordRegexPattern, regexp2.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	server.Group("/users").
		POST("/signup", u.Signup).
		POST("/login", u.Login).
		POST("/edit", u.Edit).
		GET("/profile", u.Profile)

}

func (u *UserHandler) Signup(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}

	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq"})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "password_mismatch",
			"message": "password and confirm password do not match",
		})
		return
	}

	if ok, err := u.passwordExp.MatchString(req.Password); err != nil {
		log.Printf("password regex2 match failed: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	} else if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid password format",
			"message": "password must be at least 8 characters with lowercase, uppercase, digit, and special character",
		})
		return
	}

	if ok, err := u.emailExp.MatchString(req.Email); err != nil {
		log.Printf("email regex2 match failed: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	} else if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid email format",
			"message": "please provide a valid email address",
		})
		return
	}

	if ok, err := HashPassword(req.Password); err != nil {
		log.Printf("password hashing failed for user %s: ", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "password hashing failed",
			"message": "unable to process password",
		})
		return
	} else {
		log.Printf("User signup successful: email=%s, hashedPassword=%s", req.Email, ok)
		// TODO: Save user to database
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "siginup successfully"})
}

func (u *UserHandler) Login(ctx *gin.Context) {
}

func (u *UserHandler) Edit(ctx *gin.Context) {
}

func (u *UserHandler) Profile(ctx *gin.Context) {
}

func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashBytes), nil
}
