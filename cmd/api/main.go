package main

import (
	"log/slog"
	"os"
	"time"

	_ "github.com/fickysp/fynt-be-api/cmd/api/docs"
	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/infrastructure/database"
	"github.com/fickysp/fynt-be-api/infrastructure/router"
	"github.com/fickysp/fynt-be-api/internal/applicant"
	"github.com/fickysp/fynt-be-api/internal/auth"
	"github.com/fickysp/fynt-be-api/internal/company"
	"github.com/fickysp/fynt-be-api/internal/documents"
	"github.com/fickysp/fynt-be-api/internal/education"
	"github.com/fickysp/fynt-be-api/internal/experiences"
	"github.com/fickysp/fynt-be-api/internal/fileuploads"
	"github.com/fickysp/fynt-be-api/internal/jobs"
	"github.com/fickysp/fynt-be-api/internal/middleware"
	"github.com/fickysp/fynt-be-api/internal/questions"
	"github.com/fickysp/fynt-be-api/internal/sectors"
	"github.com/fickysp/fynt-be-api/internal/submissions"
	"github.com/fickysp/fynt-be-api/internal/transaction"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/fickysp/fynt-be-api/pkg/payment"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/storage"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// @title Fynt API
// @version 1.0
// @description REST API untuk sistem rekrutmen Fynt (Seeker & Recruiter).
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		time.Local = loc
	}
	cfg := config.Load()

	database.RunDBMigration(cfg.MigrationPath, cfg.DatabaseURL)
	val := validation.NewStructValidator()

	pgClient := payment.NewMidtransClient(cfg.MidtransServerKey, cfg.MidtransEnv)

	supabaseURL := cfg.SupabaseURL
	supabaseBucket := cfg.SupabaseBucket
	supabaseKey := cfg.SupabaseKey
	storageClient := storage.NewSupabaseStorage(supabaseURL, supabaseKey, supabaseBucket)

	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected!")

	authRepo := auth.NewAuthRepository(db)
	companyRepo := company.NewRepository(db)
	applicantRepo := applicant.NewRepository(db)
	profileRepo := auth.NewProfileRepositoryAdapter(companyRepo, applicantRepo)
	walletRepo := wallet.NewWalletRepository(db)
	transactionRepo := transaction.NewTransactionRepository(db)
	sectorRepo := sectors.NewSectorRepository(db)
	jobsRepo := jobs.NewJobRepository(db)
	questionsRepo := questions.NewQuestionRepository(db)
	submissionRepo := submissions.NewSubmissionRepository(db)
	documentRepo := documents.NewDocumentRepository(db)
	educationRepo := education.NeweducationRepository(db)
	experienceRepo := experiences.NewExperienceRepository(db)

	authUsecase := auth.NewAuthUsecase(authRepo, profileRepo, walletRepo, sectorRepo)
	applicantUsecase := applicant.NewApplicantUsecase(applicantRepo, storageClient)
	companyUsecase := company.NewCompanyUsecase(companyRepo, walletRepo, sectorRepo, storageClient)
	walletUsecase := wallet.NewWalletUsecase(walletRepo, transactionRepo, pgClient, profileRepo)
	jobsUsecase := jobs.NewJobUsecase(jobsRepo, walletRepo, questionsRepo)
	submissionUsecase := submissions.NewSubmissionUsecase(submissionRepo, jobsRepo, walletRepo, applicantRepo)
	transactionUsecase := transaction.NewTransactionUsecase(transactionRepo)
	documentUsecase := documents.NewDocumentUsecase(documentRepo, storageClient)
	educationUsecase := education.NewEducationUsecase(educationRepo)
	experienceUsecase := experiences.NewExperienceUsecase(experienceRepo)
	questionsUsecase := questions.NewQuestionUsecase(questionsRepo)
	sectorUsecase := sectors.NewSectorsUsecase(sectorRepo)

	authHandler := auth.NewAuthHandler(authUsecase, val, cfg)
	walletHandler := wallet.NewWalletHandler(walletUsecase, val, cfg)
	jobsHandler := jobs.NewJobsHandler(jobsUsecase, val, cfg)
	submissionHandler := submissions.NewSubmissionsHandler(submissionUsecase, val)
	transactionHandler := transaction.NewTransactionHandler(transactionUsecase, cfg)
	fileUploadHandler := fileuploads.NewFileUploadsHandler(storageClient)
	applicationHandler := applicant.NewApplicantHandler(applicantUsecase, val)
	educationHandler := education.NewEducationHandler(educationUsecase)
	companyHandler := company.NewCompanyHandler(companyUsecase, val, cfg)
	docmentHandler := documents.NewDocumentHandler(documentUsecase)
	experienceHandler := experiences.NewExperienceHandler(experienceUsecase)
	questionHandler := questions.NewQuestionHandler(questionsUsecase)
	sectorHandler := sectors.NewSectorsHandler(sectorUsecase)

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(response.BaseResponse{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://fynt.vercel.app"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH"},
		AllowCredentials: true,
	}))
	router.SetupRoutes(app, cfg, middleware.NewAuthMiddleware(cfg), authHandler, applicationHandler, companyHandler, jobsHandler, walletHandler, submissionHandler, transactionHandler, fileUploadHandler, sectorHandler, educationHandler, experienceHandler, docmentHandler, questionHandler)

	port := os.Getenv("PORT")

	if port == "" {
		port = cfg.AppPort
	}

	if port == "" {
		port = "8080"
	}

	slog.Info("HTTP server started", "port", port)

	if err := app.Listen("0.0.0.0:" + port); err != nil {
		slog.Error("failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
