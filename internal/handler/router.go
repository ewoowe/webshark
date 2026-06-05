package handler

import (
	"net/http"
	"webshark/internal/middleware"
)

func SetupRouter() *http.ServeMux {
	router := http.NewServeMux()

	// 静态文件服务
	router.HandleFunc("/", serveStaticFile("static/index.html"))
	
	// API 路由
	router.Handle("/api/interfaces", middleware.Logging(http.HandlerFunc(getInterfaces)))
	router.Handle("/api/capture/start", middleware.Logging(http.HandlerFunc(startCapture)))
	router.Handle("/api/capture/stop", middleware.Logging(http.HandlerFunc(stopCapture)))
	
	// WebSocket 路由
	router.Handle("/ws/capture", middleware.Logging(http.HandlerFunc(captureWebSocket)))

	// 静态资源目录
	fs := http.FileServer(http.Dir("static/"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	return router
}

func serveStaticFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
