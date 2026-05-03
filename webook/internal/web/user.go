package web

import (
	"errors"
	"log"
	"net/http"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/TeamMeng/go-demo/webook/internal/service"
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#\$%\^&\*]).{8,}$`
	)
	emailExp := regexp2.MustCompile(emailRegexPattern, regexp2.None)
	passwordExp := regexp2.MustCompile(passwordRegexPattern, regexp2.None)
	return &UserHandler{
		svc:         svc,
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
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "jsonReq"})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "password and confirm password do not match",
		})
		return
	}

	if ok, err := u.passwordExp.MatchString(req.Password); err != nil {
		log.Printf("password regex2 match failed: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	} else if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "password must be at least 8 characters with lowercase, uppercase, digit, and special character",
		})
		return
	}

	if ok, err := u.emailExp.MatchString(req.Email); err != nil {
		log.Printf("email regex2 match failed: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	} else if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "please provide a valid email address",
		})
		return
	}

	err := u.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})

	if errors.Is(err, service.ErrUserDuplicalicate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "siginup successfully"})
}

func (u *UserHandler) Login(ctx *gin.Context) {
}

func (u *UserHandler) Edit(ctx *gin.Context) {
}

func (u *UserHandler) Profile(ctx *gin.Context) {
}
