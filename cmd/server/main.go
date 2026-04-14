package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/rooted-dating/rooted-server/internal/admin"
	"github.com/rooted-dating/rooted-server/internal/chat"
	"github.com/rooted-dating/rooted-server/internal/matching"
	"github.com/rooted-dating/rooted-server/internal/media"
	"github.com/rooted-dating/rooted-server/internal/moderation"
	"github.com/rooted-dating/rooted-server/internal/notification"
	"github.com/rooted-dating/rooted-server/internal/payment"
	"github.com/rooted-dating/rooted-server/internal/shared/config"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
	"github.com/rooted-dating/rooted-server/internal/shared/middleware"
	"github.com/rooted-dating/rooted-server/internal/shared/telegram"
	"github.com/rooted-dating/rooted-server/internal/user"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Infrastructure ---

	cfg := config.Load()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("PostgreSQL: %v", err)
	}
	defer db.Close()

	redisClient, err := database.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis not available: %v (caching disabled)", err)
	}
	rdb := database.NewSafeRedis(redisClient)
	defer rdb.Close()

	dynConfig := config.NewDynamicConfig(db, rdb)
	bot := telegram.NewBot(cfg.TelegramBotToken)

	// --- Repositories ---

	userRepo := user.NewPostgresRepo(db)
	matchingRepo := matching.NewPostgresRepo(db)
	chatRepo := chat.NewPostgresRepo(db)
	paymentRepo := payment.NewPostgresRepo(db)

	// --- Services ---

	userService := user.NewService(userRepo, rdb)
	matchingService := matching.NewService(matchingRepo, rdb, dynConfig)
	chatService := chat.NewService(chatRepo, rdb, dynConfig)
	paymentService := payment.NewService(paymentRepo, bot, dynConfig)
	notifService := notification.NewService(cfg.TelegramBotToken, rdb, dynConfig)

	// Initialize media service (R2/S3)
	var mediaService *media.Service
	if cfg.R2AccountID != "" {
		r2Resolver := s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID),
			}, nil
		})
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, "")),
			awsconfig.WithRegion("auto"),
		)
		if err != nil {
			log.Printf("Warning: R2 not configured: %v", err)
		} else {
			s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
				o.EndpointResolver = r2Resolver
			})
			mediaService = media.NewService(s3Client, cfg.R2BucketName, cfg.R2PublicURL)
			log.Println("R2 media storage initialized")
		}
	} else {
		log.Println("Warning: R2 not configured — photo uploads disabled")
	}

	// --- Handlers ---

	userHandler := user.NewHandler(userService, mediaService)
	matchingHandler := matching.NewHandler(matchingService, userService, chatService, notifService)
	chatHandler := chat.NewHandler(chatService, userService)
	paymentHandler := payment.NewHandler(paymentService, userService)
	moderationHandler := moderation.NewHandler(db, dynConfig, userService)
	adminHandler := admin.NewHandler(db, dynConfig)
	adminAuthHandler := admin.NewAuthHandler(db, cfg.JWTSecret)
	webhookHandler := telegram.NewWebhookHandler(bot, userService, chatService, cfg.TelegramWebAppURL)

	bot.SetBotCommands(ctx)

	// --- Fiber App ---

	app := fiber.New(fiber.Config{
		AppName:      "Rooted Dating API",
		ServerHeader: "Rooted",
		BodyLimit:    10 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} ${method} ${path} ${latency}\n",
	}))
	app.Use(middleware.RequestLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Telegram-Init-Data",
	}))

	// --- Routes ---

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/webhook/telegram", webhookHandler.Handle)

	api := app.Group("/api", middleware.TelegramAuth(cfg.TelegramBotToken, rdb))
	userHandler.RegisterRoutes(api)
	matchingHandler.RegisterRoutes(api)
	chatHandler.RegisterRoutes(api)
	paymentHandler.RegisterRoutes(api)
	moderationHandler.RegisterRoutes(api)

	// Admin auth (login — no JWT required)
	adminAuthHandler.RegisterRoutes(app)

	// Admin routes (JWT required)
	adminGroup := app.Group("/admin", adminAuthHandler.AdminJWTAuth())
	adminHandler.RegisterRoutes(adminGroup)

	// --- Graceful Shutdown ---

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		app.Shutdown()
	}()

	log.Printf("Rooted starting on :%s (%s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
