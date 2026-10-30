package education

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EducationRepository interface {
	CreateEducation(ctx context.Context, edu *Education) error
	DeleteEducation(ctx context.Context, id, applicantID uuid.UUID) error
	UpdateEducation(ctx context.Context, edu *Education) error
	GetEducation(ctx context.Context, appID string) ([]Education, error)
}

type educationRepository struct {
	db *sqlx.DB
}

func NeweducationRepository(db *sqlx.DB) EducationRepository {
	return &educationRepository{db: db}
}

func (r *educationRepository) CreateEducation(ctx context.Context, edu *Education) error {
	query := `INSERT INTO educations (applicant_id, institution, degree, field_of_study, periode) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, edu.ApplicantID, edu.Institution, edu.Degree, edu.FieldOfStudy, edu.Periode)
	return err
}

func (r *educationRepository) DeleteEducation(ctx context.Context, id, applicantID uuid.UUID) error {
	query := `DELETE FROM educations WHERE id = $1 AND applicant_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, applicantID)
	return err
}
func (r *educationRepository) UpdateEducation(ctx context.Context, edu *Education) error {
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

	if edu.Institution != "" {
		appendClause("institution", edu.Institution)
	}
	if edu.Degree != "" {
		appendClause("degree", edu.Degree)
	}
	if edu.FieldOfStudy != nil {
		appendClause("field_of_study", edu.FieldOfStudy)
	}

	if edu.Periode != "" {
		appendClause("periode", edu.Periode)
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE educations SET %s WHERE id = $%d AND applicant_id = $%d",
		strings.Join(setClauses, ", "), argCounter, argCounter+1)

	args = append(args, edu.ID, edu.ApplicantID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}
func (r *educationRepository) GetEducation(ctx context.Context, appID string) ([]Education, error) {
	query := `SELECT * FROM educations WHERE applicant_id = $1 ORDER BY created_at DESC`

	var edus []Education
	if err := r.db.SelectContext(ctx, &edus, query, appID); err != nil {
		return nil, err
	}

	return edus, nil
}
