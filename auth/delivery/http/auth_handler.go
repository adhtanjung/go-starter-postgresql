package http

import (
	"net/http"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

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
	e.POST("/forgot-password", handler.ForgotPassword)
	// apiGroup := e.Group("auth")
	// apiGroup.POST("/login", handler.Login)
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
	token, errLogin := a.AUsecase.Login(ctx, auth)

	if errLogin != nil {
		return c.JSON(getStatusCode(errLogin), ResponseError{Message: errLogin.Error()})
	}
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"token": token})

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
