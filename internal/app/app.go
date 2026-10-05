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
	ExperimentSvc := service.NewExperimentService(ExperimentRepo, flagRepo)

	approvalRepo := pgrepo.NewExperimentApprovalRepository(db)
	groupRepo := pgrepo.NewApproverGroupRepository(db)
	lifecycleSvc := service.NewLifecycleService(ExperimentRepo, approvalRepo, groupRepo)
	groupSvc := service.NewApproverGroupService(groupRepo, userRepo)

	engine := router.New(router.Deps{
		Health: handler.NewHealthHandler(sqlDB),
		Auth:   middlewares.Authenticate(userRepo),
		Handlers: []router.Registrar{
			handler.NewFeatureFlagHandler(flagSvc),
			handler.NewUsersHandler(userSvc),
			handler.NewExperimentHandler(ExperimentSvc, lifecycleSvc),
			handler.NewApproverGroupHandler(groupSvc)},
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
