package http

import (
	"net/http"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"
)

type ResponseError struct {
	Message string `json:"message"`
}

type UserHandler struct {
	UUsecase domain.UserUsecase
}

func NewUserHandler(e *echo.Group, us domain.UserUsecase) {
	handler := &UserHandler{
		UUsecase: us,
	}

	// e.GET("/users", handler.FetchUser)
	e.POST("/users", handler.Store)
	e.PUT("/users/:id", handler.Update)
	e.GET("/users/:id", handler.GetByID)
	// e.GET("/users/:id")
	// e.DELETE("/users/:id")

}

// func (u *UserHandler) FetchUser(c echo.Context) error {
// 	numS := c.QueryParam("num")
// 	num, _ := strconv.Atoi(numS)
// 	cursor := c.QueryParam("cursor")
// 	ctx := c.Request().Context()

// 	listUs, nextCursor, err := u.UUsecase.Fetch(ctx, cursor, int64(num))
// 	if err != nil {
// 		return c.JSON(getStatusCode(err), ResponseError{
// 			Message: err.Error(),
// 		})
// 	}
// 	c.Response().Header().Set(`X-Cursor`, nextCursor)
// 	return c.JSON(http.StatusOK, listUs)

// }

func (u *UserHandler) Store(c echo.Context) (err error) {
	var user domain.User
	err = c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	var ok bool
	if ok, err = isRequestValid(&user); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()
	err = u.UUsecase.Store(ctx, &user, &domain.UserRole{})
	if err != nil {

		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, user)

}

func (u *UserHandler) GetByID(c echo.Context) (err error) {
	id := c.Param("id")
	if len(id) <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"message": "id is required"})
	}
	ctx := c.Request().Context()

	user, err := u.UUsecase.GetByID(ctx, id)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UserHandler) Update(c echo.Context) (err error) {
	var user domain.User
	err = c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	id := c.Param("id")
	if len(id) <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"message": "id is required"})
	}
	user.ID = id
	ctx := c.Request().Context()
	err = u.UUsecase.Update(ctx, &user)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "user updated"})
}

func isRequestValid(m *domain.User) (bool, error) {
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
