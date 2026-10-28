package router

import (
	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/internal/auth"
	"github.com/fickysp/fynt-be-api/internal/fileuploads"
	"github.com/fickysp/fynt-be-api/internal/jobs"
	"github.com/fickysp/fynt-be-api/internal/middleware"
	"github.com/fickysp/fynt-be-api/internal/submissions"
	"github.com/fickysp/fynt-be-api/internal/transaction"
	"github.com/fickysp/fynt-be-api/internal/wallet"
	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, authMid middleware.AuthMiddleware, ah *auth.AuthHandler, jh *jobs.JobsHandler, wh *wallet.WalletHandler, sh *submissions.SubmissionsHandler, th *transaction.TransactionHandler, fh *fileuploads.FileUploadsHandler) {
	//base API
	api := app.Group("/api/v1")
	api.Group("/dashboard", authMid.RequireToken(), authMid.RequireRole("seeker"))
	api.Group("/recruiter/dashboard", authMid.RequireToken(), authMid.RequireRole("recruiter"))

	//public endpoint
	authGroup := api.Group("/auth")
	authGroup.Post("/login", ah.Login)
	authGroup.Post("/register/seeker", ah.RegisterSeeker)
	authGroup.Post("/register/recruiter", ah.RegisterRecruiter)
	authGroup.Post("/logout", authMid.RequireRole(), ah.Login)

	jobsGroup := api.Group("/jobs")
	jobsGroup.Get("/categories", jh.GetAllCategories)
	jobsGroup.Get("/find", jh.GetAllJob)
	jobsGroup.Get("/:id", jh.GetJobByID)

	transactionGroup := api.Group("/transaction")
	transactionGroup.Post("/notification", th.MidtransNotification)

	//authorized endpoint
	walletGroup := api.Group("/wallet", authMid.RequireToken(), authMid.RequireRole("seeker", "recruiter"))
	walletGroup.Get("/get-balance", wh.GetBalance)
	walletGroup.Post("/top-up", wh.TopUp)

	companyJobsGroup := jobsGroup.Group("/", authMid.RequireToken(), authMid.RequireRole("recruiter"))
	companyJobsGroup.Post("/", jh.CreateJob)
	companyJobsGroup.Patch("/:id", jh.UpdateJob)

	seekerJobsGroup := jobsGroup.Group("/", authMid.RequireToken(), authMid.RequireRole("seeker"))
	seekerJobsGroup.Patch("/:id/apply", sh.ApplyJob)

	fileUploadGroup := api.Group("/upload", authMid.RequireToken())
	fileUploadGroup.Post("/document", authMid.RequireRole("seeker", "recruiter"), fh.UploadDocument)
	fileUploadGroup.Post("/picture", authMid.RequireRole("seeker", "recruiter"), fh.UploadPicture)

}
