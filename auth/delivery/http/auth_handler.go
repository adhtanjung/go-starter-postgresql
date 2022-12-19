package http

import (
	"net/http"
	"os"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/pkg/helpers"
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
	e.POST("/forgot-password", handler.ForgotPassword)
	e.GET("/google-oauth-login", handler.GoogleOauthLogin)
	e.GET("/auth", handler.GoogleOauth)
	e.GET("/auth/callback", handler.GoogleOauthCallback)
	// apiGroup := e.Group("auth")
	// apiGroup.POST("/login", handler.Login)
}

func (a *AuthHandler) GoogleOauthLogin(c echo.Context) (err error) {
	// Get the current working directory.
	// dir, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// htmlPath := filepath.Join(DIR, "/web/oauth_login.html")
	// t, err := template.ParseFiles(htmlPath)
	// if err != nil {
	// 	return c.String(http.StatusNotFound, "failed to fetch HTML template: template not found")
	// }
	// t.Execute(c.Response(), false)
	return c.Render(http.StatusOK, "oauth_login.html", false)
}
func (a *AuthHandler) GoogleOauth(c echo.Context) (err error) {
	gothic.BeginAuthHandler(c.Response(), c.Request())
	return
}
func (a *AuthHandler) GoogleOauthCallback(c echo.Context) (err error) {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())

	if err != nil {
		// fmt.Fprintln(c.Response(), err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// htmlPath := filepath.Join(DIR, "/web/oauth_success.html")

	// t, _ := template.ParseFiles(htmlPath)
	// t.Execute(c.Response(), user)
	hashed, err := helpers.HashPassword(user.UserID)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	var userDetail domain.User
	userDetail.Email = user.Email
	userDetail.Name = user.Name
	userDetail.Username = user.FirstName + "googlegenerated"
	userDetail.Password = hashed
	err = a.AUsecase.Register(c.Request().Context(), &userDetail, &domain.UserRole{})
	if err == nil {
		// TODO: login if user already registered
		// a.AUsecase.Login()
		return c.Render(http.StatusOK, "oauth_success.html", user)
	}
	return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})

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

func (a *AuthHandler) Login(c echo.Context) (err error) {
	var auth domain.Auth
	err = c.Bind(&auth)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	ctx := c.Request().Context()
	token, refreshToken, errLogin := a.AUsecase.Login(ctx, auth)

	if errLogin != nil {
		return c.JSON(getStatusCode(errLogin), ResponseError{Message: errLogin.Error()})
	}
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"token": token, "refresh_token": refreshToken})

}
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
	err = u.AUsecase.Register(ctx, &user, &domain.UserRole{})
	if err != nil {

		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, user)

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
