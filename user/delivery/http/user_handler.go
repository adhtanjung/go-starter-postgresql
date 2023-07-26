package http

import (
	"net/http"

	"github.com/adhtanjung/go-starter/domain"
	"github.com/google/uuid"
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

	e.POST("/users", handler.Store)
	e.POST("/users/resend-email-verification", handler.Store)
	e.PUT("/users/:id", handler.Update)
	e.GET("/users/:id", handler.GetByID)
	// e.GET("/refresh-token", handler.RefreshToken)

}

func (u *UserHandler) RefreshToken(c echo.Context) (err error) {
	userID := c.Get("user_id")
	ctx := c.Request().Context()
	toUUIDType, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	refreshToken, accessToken, err := u.UUsecase.GetUsingRefreshToken(ctx, toUUIDType)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": accessToken, "refresh_token": refreshToken})

}

func (u *UserHandler) Store(c echo.Context) (err error) {
	// log.Println(c.Get("user_id"))
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
	toUUIDType, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	user, err := u.UUsecase.GetByID(ctx, toUUIDType)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UserHandler) Update(c echo.Context) (err error) {
	var user domain.User
	file, _ := c.FormFile("profile_pic")
	if file != nil {
		user.File = file
	}
	err = c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	id := c.Param("id")
	if len(id) <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{"message": "id is required"})
	}
	toUUIDType, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "invalid uuid"})
	}
	user.ID = toUUIDType
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
