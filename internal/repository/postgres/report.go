package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepository struct {
	DB *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{
		DB: db,
	}
}
func (r *ReportRepository) Save(
	ctx context.Context,
	report *models.Report,
) error {
	const op = "Report.Save"

	db := r.DB.WithContext(ctx)

	var existingReport models.Report

	err := db.
		Where(
			"experiment_id = ? AND variant_id = ? AND metric_id = ?",
			report.ExperimentID,
			report.VariantID,
			report.MetricID,
		).
		First(&existingReport).Error

	if err == nil {
		err = db.Save(report).Error
		if err != nil {
			return repository.CheckError(ctx, op, err)
		}

		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.CheckError(ctx, op, err)
	}
	fmt.Printf(
		"Saving report: experiment_id=%s, metric_id=%s, variant_id=%s, count=%f\n",
		report.ExperimentID,
		report.MetricID,
		report.VariantID,
		report.Count,
	)
	err = db.Create(report).Error
	if err != nil {
		return repository.CheckError(ctx, op, err)
	}

	return nil
}
func (r *ReportRepository) Get(ctx context.Context, id uuid.UUID) ([]models.Report, error) {
	const op = "Report.Get"
	var reports []models.Report
	err := r.DB.WithContext(ctx).Where("experiment_id=?", id).Find(&reports).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return reports, nil

}
func (r *ReportRepository) GetAll(ctx context.Context) ([]models.Report, error) {
	const op = "Report.GetAll"
	var reports []models.Report
	err := r.DB.WithContext(ctx).Find(&reports).Error
	if err = repository.CheckError(ctx, op, err); err != nil {
		fmt.Println("11")
		return nil, err

	}
	return reports, nil
}
