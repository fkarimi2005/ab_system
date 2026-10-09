package app

import (
	"AB_system/internal/http/middlewares"
	"context"
	"errors"
	"log/slog"

	"AB_system/internal/configs"
	"AB_system/internal/domain/service"
	handler "AB_system/internal/http/handlers"
	"AB_system/internal/http/router"
	pgrepo "AB_system/internal/repository/postgres"
	"AB_system/internal/server"
	database "AB_system/pkg/db/postgres"

	"gorm.io/gorm"
)

type App struct {
	db     *gorm.DB
	server *server.Server
}

func New(ctx context.Context, cfg configs.Config) (*App, error) {
	db, err := database.ConnectDB(ctx, cfg.DB)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db); err != nil {
		_ = database.CloseDB(db)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		_ = database.CloseDB(db)
		return nil, err
	}

	flagRepo := pgrepo.NewFeatureFlagRepository(db)
	flagSvc := service.NewFeatureFlagService(flagRepo)
	roleRepo := pgrepo.NewRoleRepository(db)
	//roleSvc := service.NewRoleService(roleRepo)
	userRepo := pgrepo.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, roleRepo)
	ExperimentRepo := pgrepo.NewExperimentRepository(db)
	eventTypeRepo := pgrepo.NewEventTypeRepository(db)
	metricRepo := pgrepo.NewMetricRepository(db)
	ExperimentSvc := service.NewExperimentService(ExperimentRepo, flagRepo, metricRepo)
	eventTypeSvc := service.NewEventTypeService(eventTypeRepo)
	metricSvc := service.NewMetricService(metricRepo, eventTypeRepo)
	groupRepo := pgrepo.NewApproverGroupRepository(db)
	groupSvc := service.NewApproverGroupService(groupRepo, userRepo)
	approvalRepo := pgrepo.NewExperimentApprovalRepository(db)
	approvalSvc := service.NewApprovalService(ExperimentRepo, approvalRepo, groupRepo)
	decideSvc := service.NewDecideService(flagRepo, ExperimentRepo)
	engine := router.New(router.Deps{
		Health: handler.NewHealthHandler(sqlDB),
		Auth:   middlewares.Authenticate(userRepo),
		Handlers: []router.Registrar{
			handler.NewFeatureFlagHandler(flagSvc),
			handler.NewUsersHandler(userSvc),
			handler.NewExperimentHandler(ExperimentSvc),
			handler.NewApprovalHandler(approvalSvc, ExperimentSvc),
			handler.NewApproverGroupHandler(groupSvc),
			handler.NewDecideHandler(decideSvc),
			handler.NewEventTypeHandler(eventTypeSvc),
			handler.NewMetricHandler(metricSvc),
		},
	},
	)

	return &App{
		db:     db,
		server: server.New(cfg.Port, engine),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() { errCh <- a.server.Run() }()

	select {
	case err := <-errCh:
		return errors.Join(err, database.CloseDB(a.db))
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownErr := a.server.Shutdown(context.Background())
	closeErr := database.CloseDB(a.db)
	return errors.Join(shutdownErr, closeErr)
}
