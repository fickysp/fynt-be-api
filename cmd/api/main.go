package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/infrastructure/database"
	"github.com/fickysp/fynt-be-api/infrastructure/router"
	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/auth"
	"github.com/fickysp/fynt-be-api/internal/company"
	"github.com/fickysp/fynt-be-api/internal/fileuploads"
	"github.com/fickysp/fynt-be-api/internal/jobs"
	"github.com/fickysp/fynt-be-api/internal/middleware"
	"github.com/fickysp/fynt-be-api/internal/questions"
	"github.com/fickysp/fynt-be-api/internal/sectors"
	"github.com/fickysp/fynt-be-api/internal/submissions"
	"github.com/fickysp/fynt-be-api/internal/transaction"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/fickysp/fynt-be-api/pkg/payment"
	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		time.Local = loc
	}
	cfg := config.Load()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected!")

	database.RunDBMigration(cfg.MigrationPath, cfg.DatabaseURL)
	val := validation.NewStructValidator()

	pgClient := payment.NewMidtransClient(cfg.MidtransServerKey, cfg.MidtransEnv)

	supabaseURL := cfg.SupabaseURL
	supabaseBucket := cfg.SupabaseBucket
	supabaseKey := cfg.SupabaseKey
	storageClient := storage.NewSupabaseStorage(supabaseURL, supabaseKey, supabaseBucket)

	authRepo := auth.NewAuthRepository(db)
	companyRepo := company.NewRepository(db)
	applicantRepo := applicant.NewRepository(db)
	profileRepo := auth.NewProfileRepositoryAdapter(companyRepo, applicantRepo)
	walletRepo := wallet.NewWalletRepository(db)
	transactionRepo := transaction.NewTransactionRepository(db)
	sectorRepo := sectors.NewSectorReporitory(db)
	jobsRepo := jobs.NewJobRepository(db)
	questionsRepo := questions.NewQuestionRepository(db)
	submissionRepo := submissions.NewSubmissionRepository(db)

	authUsecase := auth.NewAuthUsecase(authRepo, profileRepo, walletRepo, sectorRepo)
	walletUsecase := wallet.NewWalletUsecase(walletRepo, transactionRepo, pgClient, profileRepo)
	jobsUsecase := jobs.NewJobUsecase(jobsRepo, walletRepo, questionsRepo)
	submissionUsecase := submissions.NewSubmissionUsecase(submissionRepo, jobsRepo, walletRepo, applicantRepo)
	transactionUsecase := transaction.NewTransactionUsecase(transactionRepo)

	authHandler := auth.NewAuthHandler(authUsecase, val, cfg)
	walletHandler := wallet.NewWalletHandler(walletUsecase, val, cfg)
	jobsHandler := jobs.NewJobsHandler(jobsUsecase, val, cfg)
	submissionHandler := submissions.NewSubmissionsHandler(submissionUsecase, val)
	transactionHandler := transaction.NewTransactionHandler(transactionUsecase)
	fileUploadHandler := fileuploads.NewFileUploadsHandler(storageClient)

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())

	app.Use(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000, https://fynt.vercel.app"},
		AllowHeaders:     []string{"Origin, Content-Type, Accept, Authorization"},
		AllowMethods:     []string{"GET, POST, HEAD, PUT, DELETE, PATCH"},
		AllowCredentials: true,
	})
	router.SetupRoutes(app, cfg, middleware.NewAuthMiddleware(cfg), authHandler, jobsHandler, walletHandler, submissionHandler, transactionHandler, fileUploadHandler)

	slog.Info("HTTP server started", "port", cfg.AppPort)
	app.Listen(":" + cfg.AppPort)
}
