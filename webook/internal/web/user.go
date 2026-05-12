package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/TeamMeng/go-demo/webook/internal/service"
	"github.com/TeamMeng/go-demo/webook/internal/web/middleware"
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const biz = "login"

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
	jwtHandler  jwtHandler
	cmd         redis.Cmdable
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#\$%\^&\*]).{8,}$`
	)
	emailExp := regexp2.MustCompile(emailRegexPattern, regexp2.None)
	passwordExp := regexp2.MustCompile(passwordRegexPattern, regexp2.None)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		jwtHandler:  newJwtHandler(),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	server.Group("/users").
		POST("/signup", u.Signup).
		// POST("/login", u.Login).
		POST("/edit", u.Edit).
		GET("/profile", u.Profile).
		GET("/logout", u.Logout).
		POST("/loginJWT", u.LoginJWT).
		POST("/login_sms/code/send", u.SendLoginSMS).
		POST("/login_sms", u.LoginSMS).
		POST("/refresh_token", u.RefreshToken).
		POST("/logoutJWT", u.LogoutHandler)
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
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "jsonReq"})
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Save()

	ctx.JSON(http.StatusOK, gin.H{"message": "login successfully"})
}

func (u *UserHandler) Edit(ctx *gin.Context) {
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	claims, ok := c.(*middleware.UserClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	println(claims.Uid)

	ctx.JSON(http.StatusOK, gin.H{"message": "This is your profile"})
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	sess.Save()
	ctx.JSON(http.StatusOK, gin.H{"message": "logout successfully"})
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "jsonReq"})
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	if err := u.jwtHandler.setLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "login successfully"})
}

func (u *UserHandler) SendLoginSMS(ctx *gin.Context) {
	type SMSReq struct {
		Phone string `json:"phone"`
	}

	var req SMSReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "jsonReq"})
		return
	}

	if req.Phone == "" {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "phone cannot be empty",
		})
		return
	}

	if err := u.codeSvc.Send(ctx, biz, req.Phone); err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system err",
		})
		return
	}

	_, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 4,
			Msg:  "system err",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "send successfully",
	})

}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type SMSLoginReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req SMSLoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)

	if err != nil {
		if errors.Is(service.ErrVerifyCodeTooManyTimes, err) {
			ctx.JSON(http.StatusBadRequest, Result{
				Code: 4,
				Msg:  "verify code too many times",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "verify code error",
		})
		return
	}

	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	u.jwtHandler.setLoginToken(ctx, user.Id)

	ctx.JSON(http.StatusOK, Result{
		Msg: "code verify successfully",
	})
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) {

	refreshToken := middleware.ExtractToken(ctx)
	var rc RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(t *jwt.Token) (any, error) {
		return u.jwtHandler.rtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "Token Unauthorized",
		})
		return
	}

	logout, err := u.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", rc.Ssid)).Result()
	if err != nil || logout > 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = u.jwtHandler.setJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "refresh successfully",
	})
}

func (u *UserHandler) LogoutHandler(ctx *gin.Context) {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	claims := ctx.MustGet("claims").(*middleware.UserClaims)

	err := u.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "logout failed",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "logout successfully",
	})
}
