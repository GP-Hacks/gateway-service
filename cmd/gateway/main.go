package main

import (
	"context"
	"net/http"
	"os"

	"github.com/GP-Hacks/kdt2024-commons/api/proto"
	"github.com/GP-Hacks/kdt2024-gateway/config"
	authclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/auth"
	charityclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/charity"
	chatclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/chat"
	placesclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/places"
	usersclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/users"
	votesclient "github.com/GP-Hacks/kdt2024-gateway/internal/grpc-clients/votes"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/auth"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/charity"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/chat"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/places"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/tokens"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/users"
	"github.com/GP-Hacks/kdt2024-gateway/internal/http-server/handlers/votes"
	"github.com/GP-Hacks/kdt2024-gateway/internal/kafka"
	"github.com/GP-Hacks/kdt2024-gateway/internal/storage"
	"github.com/GP-Hacks/kdt2024-gateway/internal/utils/logger"
	websocket "github.com/GP-Hacks/kdt2024-gateway/internal/web_socket"
	proto_auth "github.com/GP-Hacks/proto/pkg/api/auth"
	proto_charity "github.com/GP-Hacks/proto/pkg/api/charity"
	proto_chat "github.com/GP-Hacks/proto/pkg/api/chat"
	proto_users "github.com/GP-Hacks/proto/pkg/api/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for handler",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	cpuUsage = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "system_cpu_usage",
			Help: "Current CPU usage as a percentage",
		},
		getCPUUsage,
	)
	memoryUsage = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "system_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
		getMemoryUsage,
	)
)

func main() {
	cfg := config.MustLoad()
	logger.SetupLogger(true, "http://infrastructure_vector_1:9880")

	log.Info().Msg("=== Gateway starter ===")

	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(memoryUsage)

	log.Info().Msg("Prometheus metrics registred")

	if err := connectToMongoDB(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed connect to mongo db")
		os.Exit(1)
	}

	chatClient, err := setupChatClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup chat client")
		os.Exit(1)
	}

	placesClient, err := setupPlacesClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup places client")
		os.Exit(1)
	}

	charityClient, err := setupCharityClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup charity client")
		os.Exit(1)
	}

	votesClient, err := setupVotesClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup votes client")
		os.Exit(1)
	}

	authClient, err := setupAuthClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup auth client")
		os.Exit(1)
	}

	usersClient, err := setupUsersClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setup users client")
		os.Exit(1)
	}

	ks := setupKafka(cfg)
	log.Info().Msg("Kafka setuped")
	defer ks.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ks.StartResponseConsumer(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed start kafka response consumer")
	}

	hub := setupWebSocket()
	log.Info().Msg("Setup web socket hub")

	router := setupRouter(cfg, charityClient, chatClient, placesClient, votesClient, authClient, usersClient, ks, hub)
	startServer(cfg, router)
}

func connectToMongoDB(cfg *config.Config) error {
	err := storage.Connect(cfg.MongoDBPath, cfg.MongoDBName, cfg.MongoDBCollection)
	if err != nil {
		return err
	}
	log.Info().Msg("Connected to mongo db")
	return nil
}

func setupChatClient(cfg *config.Config) (proto_chat.ChatServiceClient, error) {
	client, err := chatclient.SetupChatClient(cfg.ChatAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("ChatClient setup successfully")
	return client, nil
}

func setupPlacesClient(cfg *config.Config) (proto.PlacesServiceClient, error) {
	client, err := placesclient.SetupPlacesClient(cfg.PlacesAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("PlacesClient setup successfully")
	return client, nil
}

func setupCharityClient(cfg *config.Config) (proto_charity.CharityServiceClient, error) {
	client, err := charityclient.SetupCharityClient(cfg.CharityAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("CharityClient setup successfully")
	return client, nil
}

func setupVotesClient(cfg *config.Config) (proto.VotesServiceClient, error) {
	client, err := votesclient.SetupVotesClient(cfg.VotesAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("VotesClient setup successfully")
	return client, nil
}

func setupAuthClient(cfg *config.Config) (proto_auth.AuthServiceClient, error) {
	client, err := authclient.SetupAuthClient(cfg.AuthAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("AuthClient setup successfully")
	return client, nil
}

func setupUsersClient(cfg *config.Config) (proto_users.UserServiceClient, error) {
	client, err := usersclient.SetupUsersClient(cfg.UsersAddress)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("UsersClient setup successfully")
	return client, nil
}

func setupRouter(cfg *config.Config, charityClient proto_charity.CharityServiceClient, chatClient proto_chat.ChatServiceClient, placesClient proto.PlacesServiceClient, votesClient proto.VotesServiceClient, authClient proto_auth.AuthServiceClient, usersClient proto_users.UserServiceClient, ks *kafka.KafkaService, hub *websocket.Hub) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(prometheusMiddleware)

	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		yamlFile, err := os.ReadFile("/root/open-api.yaml")
		if err != nil {
			http.Error(w, "Unable to read swagger.yaml", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		_, _ = w.Write(yamlFile)
	})
	router.Get("/api/docs/redoc", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		htmlFile, err := os.ReadFile("/root/redoc-static.html")
		if err != nil {
			http.Error(w, "Unable to read redoc", http.StatusInternalServerError)
			return
		}
		w.Write(htmlFile)
	})
	router.Get("/api/docs/swagger", httpSwagger.Handler(httpSwagger.URL("0.0.0.0:8080/swagger")))

	router.Get("/api/chat/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(hub, ks, cfg.ResponseTimeout, w, r)
	})
	router.Get("/api/chat/history", chat.NewGetHistoryHandler(chatClient))

	router.Post("/api/users/token", tokens.NewAddTokenHandler())

	router.Post("/api/places", places.NewGetPlacesHandler(placesClient))
	router.Get("/api/places/categories", places.NewGetCategoriesHandler(placesClient))
	router.Get("/api/places/tickets", places.NewGetTicketsHandler(placesClient))
	router.Post("/api/places/buy", places.NewBuyTicketHandler(placesClient))

	router.Get("/api/charity", charity.NewGetCollectionsHandler(charityClient))
	router.Get("/api/charity/categories", charity.NewGetCategoriesHandler(charityClient))
	router.Post("/api/charity/donate", charity.NewDonateHandler(charityClient))

	router.Get("/api/votes", votes.NewGetVotesHandler(votesClient))
	router.Get("/api/votes/categories", votes.NewGetCategoriesHandler(votesClient))
	router.Get("/api/votes/info", votes.NewGetVoteInfoHandler(votesClient))
	router.Post("/api/votes/rate", votes.NewVoteRateHandler(votesClient))
	router.Post("/api/votes/petition", votes.NewVotePetitionHandler(votesClient))
	router.Post("/api/votes/choice", votes.NewVoteChoiceHandler(votesClient))

	router.Post("/api/auth/sign_up", auth.NewSignUpHandler(authClient))
	router.Post("/api/auth/sign_in", auth.NewSignInHandler(authClient))
	router.Post("/api/auth/refresh_tokens", auth.NewRefreshTokensHandler(authClient))
	router.Post("/api/auth/logout", auth.NewLogoutHandler(authClient))
	router.Get("/api/auth/confirm/{token}", auth.NewConfirmEmailPageHandler(authClient))
	router.Post("/api/auth/resend_confirmation_mail", auth.NewResendConfiramtionMailHandler(authClient))

	router.Get("/api/users/me", users.NewGetMeHandler(usersClient))
	router.Post("/api/users/update", users.NewUpdateHandler(usersClient))
	router.Post("/api/users/upload_avatar", users.NewUploadAvatarHandler(usersClient))

	router.Handle("/metrics", promhttp.Handler())

	log.Info().Msg("Router successfully created with defined routes")
	return router
}

func startServer(cfg *config.Config, router *chi.Mux) {
	srv := http.Server{
		Addr:         cfg.LocalAddress,
		Handler:      router,
		WriteTimeout: cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info().Msg("Starting HTTP server")
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Msg("Server encountered an error")
		return
	}

	log.Info().Msg("Server shutdown gracefully")
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		excludedPaths := map[string]bool{
			"/api/docs/*": true,
			"/metrics":    true,
		}

		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.Method, r.URL.Path))
		defer timer.ObserveDuration()

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		next.ServeHTTP(w, r)
	})
}

func getCPUUsage() float64 {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0.0
	}
	return percentages[0] * 100
}

func getMemoryUsage() float64 {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0.0
	}
	return vmStat.UsedPercent
}

func setupWebSocket() *websocket.Hub {
	hub := websocket.NewHub()
	go hub.Run()

	return hub
}

func setupKafka(config *config.Config) *kafka.KafkaService {
	kafkaService, err := kafka.NewKafkaService(config)
	if err != nil {
		log.Fatal().Msg("Ошибка создания Kafka сервиса: " + err.Error())
	}

	return kafkaService
}
