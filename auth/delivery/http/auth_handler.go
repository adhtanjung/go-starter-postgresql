package http

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adhtanjung/go-starter/domain"
	"github.com/adhtanjung/go-starter/pkg/responses"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

var DIR, _ = os.Getwd()

type ResponseError struct {
	Message string `json:"message"`
}

type AuthHandler struct {
	AUsecase domain.AuthUsecase
}
type UserResponse struct {
	Data       domain.User `json:"data"`
	StatusCode int         `json:"status_code"`
}

func NewAuthHandler(e *echo.Echo, au domain.AuthUsecase) {
	handler := &AuthHandler{
		AUsecase: au,
	}

	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	e.GET("/get-cookie", handler.GetCookie)
	e.POST("/forgot-password", handler.ForgotPassword)
	e.GET("/google-oauth-login", handler.GoogleOauthLogin)
	e.GET("/auth", handler.GoogleOauth)
	e.GET("/auth/callback", handler.GoogleOauthCallback)
	// apiGroup := e.Group("auth")
	// apiGroup.POST("/login", handler.Login)
}
func (a *AuthHandler) GetCookie(c echo.Context) (err error) {
	cookie1, err := c.Cookie("test_cookie")
	if err != nil {
		return err
	}
	log.Println(cookie1.Value)

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return err
	}
	cookie2, err := c.Cookie("access_token")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{"refresh_token": cookie.Value, "access_token": cookie2.Value})
}

func (a *AuthHandler) GoogleOauthLogin(c echo.Context) (err error) {
	return c.Render(http.StatusOK, "oauth_login.html", false)
}
func (a *AuthHandler) GoogleOauth(c echo.Context) (err error) {

	q := c.Request().URL.Query()
	q.Add("prompt", "select_account")

	c.Request().URL.RawQuery = q.Encode()
	// log.Println(q)
	// log.Println(c.Request().URL.RawQuery)
	// c.Request().RequestURI = "/auth?provider=google&prompt=select_account"
	log.Println(c.Request().RequestURI)
	gothic.BeginAuthHandler(c.Response(), c.Request())
	return
}
func (a *AuthHandler) GoogleOauthCallback(c echo.Context) (err error) {
	// q := c.Request().URL.Query()
	// q.Add("prompt", "select_account")

	// c.Request().URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	now := time.Now()

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	ctx := c.Request().Context()
	var userDetail domain.User
	userDetail.Email = user.Email
	userDetail.OauthProvider = "google"
	userDetail.OauthToken = user.AccessToken
	userDetail.VerifiedAt = &now

	registerData, err := a.AUsecase.Register(ctx, &userDetail, &domain.UserRole{}, true)
	if err != nil {
		if strings.Contains(err.Error(), "taken") {
			data, err := a.AUsecase.Login(ctx, domain.Auth{UsernameOrEmail: user.Email}, true)
			if err != nil {
				return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
			}

			c.SetCookie(&http.Cookie{
				Name:     "refresh_token",
				Value:    data.RefreshToken,
				Domain:   "localhost",
				MaxAge:   60 * 60 * 24 * 7,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(time.Hour),
			})
			c.SetCookie(&http.Cookie{
				Name:     "access_token",
				Value:    data.Token,
				Domain:   "localhost",
				MaxAge:   60 * 60 * 24,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(time.Hour),
			})
			c.SetCookie(&http.Cookie{
				Name:     "logged_in",
				Value:    "true",
				Domain:   "localhost",
				MaxAge:   60 * 60 * 24,
				Path:     "/",
				HttpOnly: false,
				Expires:  time.Now().Add(time.Hour),
			})
			// b, err := json.Marshal(map[string]string{"tokennn": token, "refresh_token": refreshToken})
			// return c.JSON(http.StatusOK, echo.Map{"token": token, "refresh_token": refreshToken})
			// c.Response().Write(b)
			return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/dashboard")
			// return c.Render(http.StatusOK, "oauth_success.html", user)
		} else {
			return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		}
	} // b, err := json.Marshal(map[string]string{"token": tokenRegist, "refresh_token": refreshTokenRegist})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    registerData.RefreshToken,
		Domain:   "localhost",
		MaxAge:   60 * 60 * 24 * 7,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour),
	})
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    registerData.Token,
		Domain:   "localhost",
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour),
	})
	c.SetCookie(&http.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Domain:   "localhost",
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		HttpOnly: false,
		Expires:  time.Now().Add(time.Hour),
	})
	// c.Response().Write(b)
	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/dashboard")
	// return c.JSON(http.StatusOK, echo.Map{"token": tokenRegist, "refresh_token": refreshTokenRegist})

}
func (a *AuthHandler) ForgotPassword(c echo.Context) (err error) {
	var email domain.ForgotPassword
	err = c.Bind(&email)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var ok bool
	if ok, err = isRequestValidForgotPass(&email); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()
	err = a.AUsecase.ForgotPassword(ctx, email.Email)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"Message": "email sent"})

}

// LoginAccount godoc
// @Summary      login
// @Description  login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param		 account	body		domain.Auth true	"login account"
// @Success      200  {object}  responses.Response
// @Router       /login [post]
func (a *AuthHandler) Login(c echo.Context) (err error) {
	var auth domain.Auth

	err = c.Bind(&auth)
	log.Println(auth.Password)
	log.Println(auth.UsernameOrEmail)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	ctx := c.Request().Context()
	data, errLogin := a.AUsecase.Login(ctx, auth, false)

	if errLogin != nil {
		return c.JSON(http.StatusNotFound, ResponseError{Message: errLogin.Error()})
	}
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	// c.SetCookie(&http.Cookie{
	// 	Name:       "refresh_token",
	// 	Value:      data.RefreshToken,
	// 	Path:       "/refresh_token",
	// 	Domain:     "localhost",
	// 	Expires:    time.Time{},
	// 	RawExpires: "",
	// 	MaxAge:     60 * 60 * 24 * 7,
	// 	Secure:     false,
	// 	HttpOnly:   false,
	// })
	// return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/dashboard")
	// tokenData := echo.Map{
	// 	"token":         data.Token,
	// 	"refresh_token": data.RefreshToken,
	// }
	response := responses.NewResponse(data, http.StatusOK, "success", "operation success")

	return c.JSON(http.StatusOK, response)

}

// RegisterAccount godoc
// @Summary      register new account
// @Description  register
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param		 account	body		domain.AuthRegister true	"Add account"
// @Success      200  {object}  domain.AuthResponse
// @Router       /register [post]
func (u *AuthHandler) Register(c echo.Context) (err error) {
	var user domain.User
	err = c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	var ok bool
	if ok, err = isRequestValidUser(&user); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()

	data, err := u.AUsecase.Register(ctx, &user, &domain.UserRole{}, false)
	if err != nil {
		if strings.Contains(err.Error(), "taken") {
			return c.JSON(http.StatusConflict, ResponseError{Message: err.Error()})
		}
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	response := responses.NewResponse(data, http.StatusOK, "success", "operation success")
	return c.JSON(http.StatusCreated, response)
	// return c.JSON(http.StatusCreated, echo.Map{"token": token, "refresh_token": refreshToken})
	// return c.JSON(http.StatusCreated, user)

}
func isRequestValidUser(m *domain.User) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}
func isRequestValid(m *domain.Auth) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}
func isRequestValidForgotPass(m *domain.ForgotPassword) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
