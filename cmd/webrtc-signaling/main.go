package main

import (
	"context"
	"errors"
	"github.com/caarlos0/env/v11"
	"github.com/ownerofglory/webrtc-signaling-go/config"
	"github.com/ownerofglory/webrtc-signaling-go/internal/handler"
	"github.com/ownerofglory/webrtc-signaling-go/internal/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slog.Info("Starting app")

	var cfg config.WebRTCSignalingAppConfig
	err := env.Parse(&cfg)
	if err != nil {
		slog.Error("Failed to parse config", "error", err)
	}

	logLevel := slog.LevelInfo
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(logLevel)

	h := http.NewServeMux()
	h.HandleFunc(handler.GetVersionPath, handler.HandleGetVersion)

	rtcConfigHanlder := handler.NewRTCConfigHandler(&cfg)
	h.HandleFunc(handler.GetRTCConfigPath, rtcConfigHanlder.HandleGetRTCConfig)

	wsHandler := handler.NewWSHandler(&cfg)
	h.HandleFunc(handler.WSPath, wsHandler.HandleWS)

	fs := http.FileServer(http.Dir("web"))
	h.Handle("/webrtc-signaling/ws/app/", http.StripPrefix("/webrtc-signaling/ws/app/", fs))

	httpServer := http.Server{
		Addr:    cfg.ServerAddr,
		Handler: middleware.CORS(cfg.AllowedOrigins)(h),
	}

	go func() {
		slog.Info("Starting HTTP Server")

		err := httpServer.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server shutdown unexpected", "err", err)
		}

		slog.Info("HTTP Server finished")
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL)
	<-sigCh

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error:", "err", err)
	}

	slog.Info("App finished")
}
