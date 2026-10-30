package experiences

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ExperienceRepository interface {
	CreateExperience(ctx context.Context, exp *Experience) error
	DeleteExperience(ctx context.Context, id, applicantID uuid.UUID) error
	UpdateExperience(ctx context.Context, exp *Experience) error
	GetExperience(ctx context.Context, appID string) ([]Experience, error)
}

type experienceRepository struct {
	db *sqlx.DB
}

func NewExperienceRepository(db *sqlx.DB) ExperienceRepository {
	return &experienceRepository{db: db}
}

func (r *experienceRepository) GetExperience(ctx context.Context, appID string) ([]Experience, error) {
	query := `SELECT * FROM experience WHERE applicant_id = $1 ORDER BY created_at DESC`
	var exp []Experience
	if err := r.db.SelectContext(ctx, &exp, query, appID); err != nil {
		return nil, err
	}
	return exp, nil
}

func (r *experienceRepository) CreateExperience(ctx context.Context, exp *Experience) error {
	query := `INSERT INTO experience (applicant_id, position, company_name, periode, jobdesc) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, exp.ApplicantID, exp.Position, exp.CompanyName, exp.Periode, exp.Jobdesc)
	return err
}

func (r *experienceRepository) DeleteExperience(ctx context.Context, id, applicantID uuid.UUID) error {
	query := `DELETE FROM experience WHERE id = $1 AND applicant_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, applicantID)
	return err
}
func (r *experienceRepository) UpdateExperience(ctx context.Context, exp *Experience) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var setClauses []string
	var args []interface{}
	argCounter := 1

	appendClause := func(column string, value interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argCounter))
		args = append(args, value)
		argCounter++
	}

	if exp.Position != "" {
		appendClause("position", exp.Position)
	}
	if exp.CompanyName != "" {
		appendClause("company_name", exp.CompanyName)
	}
	if exp.Jobdesc != nil {
		appendClause("jobdesc", exp.Jobdesc)
	}
	if exp.Periode != "" {
		appendClause("periode", exp.Periode)
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE experience SET %s WHERE id = $%d AND applicant_id = $%d",
		strings.Join(setClauses, ", "), argCounter, argCounter+1)

	args = append(args, exp.ID, exp.ApplicantID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}
