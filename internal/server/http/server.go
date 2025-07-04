package server

import (
	"fmt"
	"log/slog"

	"GophKeeper.ru/internal/server/http/middlewares"
	"GophKeeper.ru/internal/server/services"
	"GophKeeper.ru/internal/server/storage"
	"github.com/gin-gonic/gin"
)

// ServiceHTTP — общий интерфейс для HTTP-сервера.
// Определяет базовые методы, которые должен реализовать сервер.
type ServiceHTTP interface {
	Run(addr string) error // Запуск сервера по указанному адресу
	Stop() error           // Остановка сервера
	AddServices() error    // Регистрация сервисов (маршрутов)
}

// Server — структура, представляющая HTTP-сервер приложения.
// Содержит ссылку на Gin-движок, адрес запуска и подключение к БД.
type Server struct {
	addr   string            // Адрес, на котором будет запущен сервер
	engine *gin.Engine       // Gin-движок для обработки HTTP-запросов
	db     *storage.Database // Подключение к базе данных
}

// NewServer создаёт новый экземпляр сервера с указанным адресом.
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// Run запускает HTTP-сервер с TLS (HTTPS), используя указанный адрес,
// а также сертификат и приватный ключ из файлов.
func (s *Server) Run(addr string) error {
	return s.engine.RunTLS(addr,
		"./localhost/cert.pem", // Путь к SSL-сертификату
		"./localhost/key.pem")  // Путь к приватному ключу
}

// Stop останавливает сервер.
// В текущей реализации просто выводит сообщение в лог и возвращает nil.
func (s *Server) Stop() error {
	slog.Info("Сервер завершил работу")
	return nil
}

// New инициализирует и настраивает сервер:
// - устанавливает режим Gin
// - создаёт движок с middleware'ами
// - регистрирует маршруты сервисов
// - возвращает готовый сервер или ошибку
func New(db *storage.Database) (*Server, error) {
	if db == nil {
		return nil, fmt.Errorf("storage is nil")
	}

	server := &Server{
		db: db,
	}
	gin.SetMode(gin.ReleaseMode)
	server.engine = gin.New()

	server.engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE")

		if c.Request.Method != "POST" && c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	server.engine.Use(func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/json")
		ctx.Next()
	})

	server.engine.Use(middlewares.WarpAuth(server.db))

	apiGroup := server.engine.Group("api")
	{
		services.Auth(apiGroup, server.db)       // Маршруты аутентификации
		services.AccessData(apiGroup, server.db) // Маршруты доступа к данным
	}

	routes := server.engine.Routes()
	slog.Info("Зарегистрированные маршруты:")
	for _, route := range routes {
		slog.Info("Route",
			"method", route.Method,
			"path", route.Path,
		)
	}

	return server, nil
}
