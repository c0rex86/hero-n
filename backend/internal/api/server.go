package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/c0re/dft/backend/internal/core"
	"github.com/c0re/dft/backend/internal/crypto"
	"github.com/c0re/dft/backend/internal/storage"
)

// Server - HTTP API сервер
type Server struct {
	port          string
	router        *mux.Router
	core          *core.Core
	crypto        *crypto.Engine
	storage       storage.Storage
	messageService *MessageService
	userService   *UserService
	authService   *AuthService
	httpSrv       *http.Server
	logger        *log.Logger
}

// NewServer - создаем новый API сервер
func NewServer(port string, core *core.Core, p2pNode interface{}, storage storage.Storage, logger *log.Logger) *Server {
	router := mux.NewRouter()

	server := &Server{
		port:    port,
		router:  router,
		core:    core,
		crypto:  core.GetCrypto(),
		storage: storage,
		logger:  logger,
	}

	// TODO: Инициализировать сервисы
	// server.messageService = NewMessageService(core, server.crypto, storage, logger)
	// server.userService = NewUserService(storage, logger)
	// server.authService = NewAuthService(storage, logger)

	// Настраиваем маршруты
	server.setupRoutes()

	return server
}

// setupRoutes - настраиваем HTTP маршруты
func (s *Server) setupRoutes() {
	// TODO: Настроить middleware цепочку
	// - Сначала logging
	// - Потом CORS
	// - Потом аутентификация (для защищенных маршрутов)

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)

	// API v1
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check (без аутентификации)
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Публичные маршруты пользователей
	api.HandleFunc("/users", s.handleRegisterUser).Methods("POST")

	// Защищенные маршруты (нужна аутентификация)
	protected := api.NewRoute().Subrouter()
	protected.Use(s.authMiddleware)

	// Сообщения
	protected.HandleFunc("/messages", s.handleSendMessage).Methods("POST")
	protected.HandleFunc("/messages/{id}", s.handleGetMessage).Methods("GET")
	protected.HandleFunc("/messages", s.handleGetMessages).Methods("GET")

	// Пользователи (защищенные)
	protected.HandleFunc("/users/{id}", s.handleGetUser).Methods("GET")
}

// Start - запускаем сервер
func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	s.logger.Printf("API сервер запущен на порту %s", s.port)
	return s.httpSrv.ListenAndServe()
}

// Stop - останавливаем сервер
func (s *Server) Stop() error {
	if s.httpSrv != nil {
		s.logger.Println("Останавливаем API сервер...")
		return s.httpSrv.Shutdown(context.Background())
	}
	return nil
}

// handleHealth - обработчик health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать проверку здоровья системы
	// - Проверить подключение к БД
	// - Проверить P2P соединения
	// - Проверить статус криптографии
	// - Проверить использование памяти/CPU

	healthData := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "0.1.0",
	}

	s.SendSuccess(w, healthData)
}

// handleSendMessage - обработчик отправки сообщения
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать отправку сообщения
	// 1. Распарсить JSON тело запроса
	// 2. Валидировать данные
	// 3. Получить ID отправителя из контекста/заголовков
	// 4. Отправить через MessageService
	// 5. Вернуть результат

	s.SendNotImplemented(w, "отправка сообщений")
}

// handleGetMessage - обработчик получения сообщения по ID
func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать получение сообщения
	// 1. Получить ID из URL параметров
	// 2. Проверить права доступа
	// 3. Получить через MessageService
	// 4. Расшифровать и вернуть

	vars := mux.Vars(r)
	messageID := vars["id"]

	s.SendErrorWithDetails(w, http.StatusNotImplemented,
		ErrCodeNotImplemented,
		ErrMsgNotImplemented,
		"MessageID: "+messageID)
}

// handleGetMessages - обработчик получения списка сообщений
func (s *Server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать получение списка сообщений
	// 1. Распарсить query параметры (limit, offset, with_user)
	// 2. Получить ID пользователя из контекста
	// 3. Получить через MessageService
	// 4. Расшифровать и вернуть

	s.SendNotImplemented(w, "получение списка сообщений")
}

// handleRegisterUser - обработчик регистрации пользователя
func (s *Server) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать регистрацию пользователя
	// 1. Распарсить JSON тело
	// 2. Валидировать данные
	// 3. Проверить уникальность username
	// 4. Сохранить в хранилище
	// 5. Вернуть user_id

	s.SendNotImplemented(w, "регистрация пользователей")
}

// handleGetUser - обработчик получения пользователя
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать получение пользователя
	// 1. Получить ID из URL
	// 2. Получить из хранилища
	// 3. Проверить права доступа
	// 4. Вернуть данные

	vars := mux.Vars(r)
	userID := vars["id"]

	s.SendErrorWithDetails(w, http.StatusNotImplemented,
		ErrCodeNotImplemented,
		ErrMsgNotImplemented,
		"UserID: "+userID)
}

// corsMiddleware - middleware для CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	// TODO: Реализовать CORS middleware
	// - Добавить заголовки Access-Control-Allow-Origin
	// - Добавить Access-Control-Allow-Methods
	// - Добавить Access-Control-Allow-Headers
	// - Обработать preflight запросы

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Заголовки CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware - middleware для аутентификации
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	// TODO: Реализовать аутентификацию
	// - Проверять токен в заголовке Authorization
	// - Валидировать токен
	// - Добавлять user_id в контекст
	// - Обрабатывать ошибки

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пока пропускаем все запросы без аутентификации
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware - middleware для логирования
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	// TODO: Реализовать логирование запросов
	// - Логировать метод, URL, IP
	// - Логировать время выполнения
	// - Логировать ошибки

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Логируем входящий запрос
		s.logger.Printf("REQUEST: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		// Логируем время выполнения
		duration := time.Since(start)
		s.logger.Printf("RESPONSE: %s %s took %v", r.Method, r.URL.Path, duration)
	})
}

// writeJSON - вспомогательный метод для записи JSON ответа
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	// TODO: Реализовать JSON сериализацию с обработкой ошибок
	// - Сериализовать data в JSON
	// - Обработать ошибки сериализации
	// - Установить Content-Type
	// - Записать статус и тело

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Пока возвращаем простую строку
	w.Write([]byte(`{"status": "not implemented"}`))
}
