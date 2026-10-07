package postgres

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/repository"
	"errors"
	"gorm.io/gorm"
)

func Migrate(dbConn *gorm.DB) error {
	if dbConn == nil {
		return errors.New("database connection is not initialized")
	}

	tables := []interface{}{
		&models.Role{},
		&models.User{},
		&models.FeatureFlag{},
		&models.Experiment{},
		&models.ExperimentVariant{},
		&models.ExperimentApproval{},
		&models.ExperimentVersion{},
	}
	for _, table := range tables {
		if err := dbConn.AutoMigrate(table); err != nil {
			return err
		}
	}

	// не более одного running/paused эксперимента на флаг
	if err := dbConn.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uniq_active_experiment_per_flag
		ON experiments (feature_flag_id)
		WHERE status IN ('running', 'paused') AND deleted_at IS NULL`).Error; err != nil {
		return err
	}

	if err := SeedRoles(dbConn); err != nil {
		return err
	}
	if err := SeedAdmin(dbConn); err != nil {
		return err
	}
	return nil

}

func SeedRoles(db *gorm.DB) error {
	names := []string{
		models.RoleAdmin, models.RoleExperimenter,
		models.RoleApprover, models.RoleViewer,
	}
	for _, name := range names {
		role := models.Role{RoleName: name}
		if err := db.Where("role_name = ?", name).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}
func SeedAdmin(db *gorm.DB) error {
	var a models.Role
	err := db.Model(&models.Role{}).Where("role_name", models.RoleAdmin).First(&a).Error
	if err != nil {
		return repository.TranslateGormError(err)

	}

	admin := models.User{
		Email:    "firuzz.7@gmail.com",
		Name:     "Firuz",
		RoleID:   a.ID,
		IsActive: true,
	}
	if err := db.Where("email", admin.Email).FirstOrCreate(&admin).Error; err != nil {
		return repository.TranslateGormError(err)
	}
	return nil
}
