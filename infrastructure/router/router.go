package router

import (
	"github.com/fickysp/fynt-be-api/config"
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
	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, authMid middleware.AuthMiddleware, ah *auth.AuthHandler, applh *applicant.ApplicantHandler, ch *company.CompanyHandler, jh *jobs.JobsHandler, wh *wallet.WalletHandler, sh *submissions.SubmissionsHandler, th *transaction.TransactionHandler, fh *fileuploads.FileUploadsHandler, secH *sectors.SectorHandler, eduH *education.EducationHandler, exH *experiences.ExperienceHandler, dH *documents.DocumentHandler, qH *questions.QuestionHandler) {

	api := app.Group("/api/v1")
	api.Get("/categories", jh.GetAllCategories)
	api.Get("/jobs", jh.GetAllJob)
	api.Post("/transaction/notification", th.MidtransNotification)
	api.Post("/refresh-token", ah.RefreshToken)
	api.Post("/upload/document", authMid.RequireToken(), fh.UploadDocument)
	api.Post("/upload/picture", authMid.RequireToken(), fh.UploadPicture)

	// AUTH
	authGroup := api.Group("/auth")
	authGroup.Post("/login", ah.Login)
	authGroup.Post("/register/seeker", ah.RegisterSeeker)
	authGroup.Post("/register/recruiter", ah.RegisterRecruiter)
	authGroup.Get("/google/login", ah.LoginGoogle)
	authGroup.Get("/google/callback", ah.CallbackGoogle)
	authGroup.Post("/logout", authMid.RequireToken(), ah.Logout)

	// SEEKER
	seekerGroup := api.Group("/seeker", authMid.RequireToken(), authMid.RequireRole("seeker"))
	seekerGroup.Get("/profile", applh.GetApplicant)
	seekerGroup.Patch("/profile", applh.UpdateProfile)
	seekerGroup.Patch("/profile/avatar", applh.UpdateAvatar)
	seekerGroup.Get("/submissions", sh.GetMySubmissions)
	seekerGroup.Get("/submissions/:id", sh.GetSubmissionDetail)

	eduGroup := seekerGroup.Group("/educations")
	eduGroup.Get("", eduH.GetEducations)
	eduGroup.Post("", eduH.AddEducation)
	eduGroup.Delete("/:id", eduH.DeleteEducation)
	eduGroup.Patch("/:id", eduH.UpdateEducation)

	expGroup := seekerGroup.Group("/experiences")
	expGroup.Get("", exH.GetExperiences)
	expGroup.Post("", exH.AddExperience)
	expGroup.Delete("/:id", exH.DeleteExperience)
	expGroup.Patch("/:id", exH.UpdateExperience)

	documentGroup := seekerGroup.Group("/documents")
	documentGroup.Post("", dH.UploadDocument)
	documentGroup.Delete("/:id", dH.DeleteDocument)
	documentGroup.Patch("/:id/set-active", dH.SetActiveDocument)
	documentGroup.Get("", dH.GetAllDocuments)

	// RECRUITER
	recruiterGroup := api.Group("/recruiter", authMid.RequireToken(), authMid.RequireRole("recruiter"))
	recruiterGroup.Post("/on-boarding", ch.CreateCompanyProfile)
	recruiterGroup.Patch("/profile", ch.UpdateProfile)
	recruiterGroup.Patch("/profile/avatar", ch.UpdateAvatar)
	recruiterGroup.Get("/dashboard/stats", jh.GetJobSummary)

	// JOBS
	jobsGroup := api.Group("/job")
	jobsGroup.Get("/search", jh.SearchJobs)
	jobsGroup.Get("/:id", jh.GetJobByID)
	jobsGroup.Get("/:id/essay-questions", qH.GetQuestions)

	jobsGroup.Post("", authMid.RequireToken(), authMid.RequireRole("recruiter"), jh.CreateJob)
	jobsGroup.Patch("/:id", authMid.RequireToken(), authMid.RequireRole("recruiter"), jh.UpdateJob)
	jobsGroup.Patch("/:id/essay-questions", authMid.RequireToken(), authMid.RequireRole("recruiter"), qH.UpdateQuestion)
	jobsGroup.Post("/:id/apply", authMid.RequireToken(), authMid.RequireRole("seeker"), sh.ApplyJob)
	jobsGroup.Get("/:id/submissions-essays", authMid.RequireToken(), sh.GetApplicantEssaysByJob)

	// QUESTIONS
	questionGroup := api.Group("/questions")
	questionGroup.Get("", qH.GetQuestions)

	recruiterQuestionGroup := questionGroup.Group("", authMid.RequireToken(), authMid.RequireRole("recruiter"))
	recruiterQuestionGroup.Put("/:id", qH.UpdateQuestion)
	recruiterQuestionGroup.Delete("/:id", qH.DeleteQuestion)

	// TRANSACTION & WALLET
	walletGroup := api.Group("/wallet", authMid.RequireToken(), authMid.RequireRole("seeker", "recruiter"))
	walletGroup.Get("/balance", wh.GetBalance)
	walletGroup.Post("/top-up", wh.TopUp)

	// SUBMISSIONS (RECRUITER)
	submissionsGroup := api.Group("/submissions", authMid.RequireToken(), authMid.RequireRole("recruiter"))
	submissionsGroup.Get("/job/:job_id", sh.GetSubmissionsByJobID)
	submissionsGroup.Patch("/:id/status", sh.UpdateSubmissionStatus)
	submissionsGroup.Patch("/:id/notes", sh.UpdateInternalNotes)

}
