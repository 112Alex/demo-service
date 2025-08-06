package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/112Alex/demo-service.git/internal/cache"
	"github.com/112Alex/demo-service.git/internal/db"
)

// Server представляет собой HTTP-сервер, который имеет доступ к кэшу и БД.
type Server struct {
	httpServer *http.Server
	cache      *cache.Cache
	db         *db.DBClient
}

// NewServer создает и возвращает новый HTTP-сервер.
func NewServer(port string, cache *cache.Cache, db *db.DBClient) *Server {
	router := http.NewServeMux()
	s := &Server{
		cache: cache,
		db:    db,
	}

	// 1. Обработчик для API-запросов
	router.HandleFunc("/order/", s.orderHandler)

	// 2. Правильная настройка для обслуживания статических файлов
	// Обработчик http.StripPrefix убирает префикс "/static/" из URL,
	// а http.FileServer ищет файлы в директории "./web/static".
	fs := http.FileServer(http.Dir("./web/static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	// 3. Обработчик для главной страницы
	// Он явно отдает index.html для корневого пути.
	router.HandleFunc("/", s.homeHandler)

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	return s
}

// Start запускает HTTP-сервер.
func (s *Server) Start() error {
	log.Printf("HTTP-сервер запущен на порту %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully останавливает HTTP-сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// homeHandler обслуживает файл index.html для корневого пути.
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./web/static/index.html")
}

// orderHandler обрабатывает запросы на получение заказа по order_uid.
func (s *Server) orderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 || pathParts[2] == "" {
		http.Error(w, "Некорректный URL. Ожидается /order/{order_uid}", http.StatusBadRequest)
		return
	}
	orderUID := pathParts[2]

	order, found := s.cache.Get(orderUID)
	if found {
		log.Printf("Заказ %s найден в кэше", orderUID)
		sendJSONResponse(w, order)
		return
	}

	log.Printf("Заказ %s не найден в кэше, ищем в БД...", orderUID)
	order, err := s.db.GetOrderFromDB(r.Context(), orderUID)
	if err != nil {
		log.Printf("Ошибка при получении заказа %s из БД: %v", orderUID, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	s.cache.Set(order.OrderUID, order)

	sendJSONResponse(w, order)
}

// sendJSONResponse отправляет ответ в формате JSON.
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Не удалось сериализовать JSON", http.StatusInternalServerError)
	}
}
