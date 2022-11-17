package http

import (
	"net/http"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type RoleHandler struct {
	RUsecase domain.RoleUsecase
}
type ResponseError struct {
	Message string `json:"message"`
}

func NewRoleHandler(e *echo.Group, r domain.RoleUsecase) {
	handler := &RoleHandler{
		RUsecase: r,
	}

	e.POST("/roles", handler.Store)
	e.GET("/roles/:name", handler.GetByName)
}

func (r *RoleHandler) Store(c echo.Context) (err error) {
	var role domain.Role
	err = c.Bind(&role)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	ctx := c.Request().Context()
	err = r.RUsecase.Store(ctx, &role)
	if err != nil {

		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"data": role,
	})

}

func (r *RoleHandler) GetByName(c echo.Context) (err error) {
	name := c.QueryParam("name")
	ctx := c.Request().Context()
	role, err := r.RUsecase.GetByName(ctx, name)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"data": role,
	})
}
func isRequestValid(m *domain.Article) (bool, error) {
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
