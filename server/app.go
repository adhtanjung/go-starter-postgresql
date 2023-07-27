package server

import (
	"log"
	"os"
	"path/filepath"

	"github.com/adhtanjung/go-starter/pkg/middlewares"
	"github.com/adhtanjung/go-starter/pkg/renderer"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

type App struct {
	e *echo.Echo
}

func NewApp() *App {
	return &App{
		e: echo.New(),
	}
}

func (a *App) Initialize() {
	// Load configuration and setup database connections
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("Error reading configuration file:", err)
	}

	// Initialize middlewares
	a.e.Use(middleware.Logger())
	a.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))
	a.e.Use(middleware.Recover())
	a.e.Use(middlewares.MiddlewareLogging)

	// Initialize renderer
	dir, err := os.Getwd()
	if err != nil {
		log.Panic("Error getting working directory:", err)
	}
	a.e.Renderer = renderer.NewRenderer(filepath.Join(dir, "/web/*.html"), true)

	// Register routes and handlers
	a.registerApiRoutes()
}

func (a *App) Run() {
	// Start the server
	err := a.e.Start(viper.GetString("server.address"))
	if err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}
