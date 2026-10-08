package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"AB_system/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExperimentRepository struct {
	db *gorm.DB
}

func NewExperimentRepository(db *gorm.DB) *ExperimentRepository {
	return &ExperimentRepository{
		db: db,
	}
}
func (r *ExperimentRepository) GetAllExperiments(
	ctx context.Context,
) ([]models.Experiment, error) {
	const op = "GetAllExperiments"
	var result []models.Experiment
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Order("created_at DESC").
		Find(&result).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ExperimentRepository) GetExperimentByID(
	ctx context.Context, experimentID uuid.UUID,
) (models.Experiment, error) {
	const op = "GetExperimentByID"
	var result models.Experiment
	err := r.db.WithContext(ctx).
		Preload("Variants").
		First(&result, "id = ?", experimentID).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.Experiment{}, err
	}
	return result, nil
}

func (r *ExperimentRepository) ExperimentExists(
	ctx context.Context, experimentID uuid.UUID,
) (bool, error) {
	const op = "ExperimentExists"
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Experiment{}).
		Where("id = ?", experimentID). // этой строки не хватало
		Count(&count).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return false, err
	}
	return count > 0, nil
}

// have to know how created UpdateExperiment, what is mean a snapshot.
func (r *ExperimentRepository) UpdateExperiment(
	ctx context.Context, e *models.Experiment, expectedVersion int, snapshot []byte,
) error {
	const op = "UpdateExperiment"

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Experiment{}).
			Omit(clause.Associations).
			Where("id = ? AND version = ? AND status = ?",
				e.ID, expectedVersion, models.ExperimentStatusDraft).
			Select("Name", "AudienceBP", "Version").
			Updates(e)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errs.ErrConflict
		}

		// 2. варианты: удалить старые, вставить новые
		if err := tx.Where("experiment_id = ?", e.ID).
			Delete(&models.ExperimentVariant{}).Error; err != nil {
			return err
		}
		for i := range e.Variants {
			e.Variants[i].ExperimentID = e.ID
		}
		if err := tx.Create(&e.Variants).Error; err != nil {
			return err
		}

		// 3. снимок версии
		return tx.Create(&models.ExperimentVersion{
			ExperimentID: e.ID,
			Version:      e.Version,
			Snapshot:     snapshot,
		}).Error
	})

	if errors.Is(err, errs.ErrConflict) {
		return err // ожидаемая ситуация, в error-лог не пишем
	}
	return repository.CheckError(ctx, op, err)
}
func (r *ExperimentRepository) CreateExperiment(
	ctx context.Context,
	experiment *models.Experiment,
) (*models.Experiment, error) {
	const op = "CreateExperiment"
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(experiment).Error; err != nil {
			return err
		}
		snapshot, err := json.Marshal(experiment)
		if err != nil {
			return err
		}
		return tx.Create(&models.ExperimentVersion{
			ExperimentID: experiment.ID,
			Version:      experiment.Version,
			Snapshot:     snapshot,
		}).Error
	})
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return experiment, nil
}

func (r *ExperimentRepository) TransitionStatus(
	ctx context.Context,
	experimentID uuid.UUID,
	from, to models.ExperimentStatus,
) error {
	const op = "TransitionStatus"

	result := r.db.WithContext(ctx).
		Model(&models.Experiment{}).
		Where("id = ? AND status = ?", experimentID, from).
		Update("status", to)

	if err := repository.CheckError(ctx, op, result.Error); err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return errs.ErrInvalidTransition
	}
	return nil
}
func (r *ExperimentRepository) ListVersions(
	ctx context.Context,
	experimentID uuid.UUID,
) ([]models.ExperimentVersion, error) {
	const op = "ListVersions"
	var result []models.ExperimentVersion
	err := r.db.WithContext(ctx).
		Model(&models.ExperimentVersion{}).
		Where("experiment_id = ?", experimentID).
		Order("version ASC").
		Find(&result).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ExperimentRepository) GetVersion(ctx context.Context,
	experimentID uuid.UUID,
	version int,
) (models.ExperimentVersion, error) {
	const op = "GetVersion"
	var result models.ExperimentVersion
	err := r.db.WithContext(ctx).
		Model(&models.ExperimentVersion{}).
		Where("experiment_id = ? AND version = ?", experimentID, version).
		First(&result).Error
	if err := repository.CheckError(ctx, op, err); err != nil {
		return models.ExperimentVersion{}, err

	}
	return result, nil
}
