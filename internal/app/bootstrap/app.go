package bootstrap

import (
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "project/docs"
	"project/internal/app/config"
	"project/internal/app/database"
	"project/internal/app/logger"
	"project/internal/app/proxy"
	"project/internal/app/server"
	"project/internal/modules/banks/catalog"
	catalogpostgres "project/internal/modules/banks/catalog/infrastructure/postgres"
	"project/internal/modules/banks/beeline"
	beelinepostgres "project/internal/modules/banks/beeline/infrastructure/postgres"
	"project/internal/modules/banks/rocketbank"
	rocketbankcheque "project/internal/modules/banks/rocketbank/infrastructure/cheque"
	rocketbankpostgres "project/internal/modules/banks/rocketbank/infrastructure/postgres"
)

var bootstrapLog = logger.New("bootstrap")

type App struct {
	Router *gin.Engine
	DB     *sql.DB
	Proxy  *proxy.Service
}

func NewApp() *App {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		bootstrapLog.Fatalf("connect postgres: %v", err)
	}

	if err := database.RunMigrations(ctx, db); err != nil {
		bootstrapLog.Fatalf("run migrations: %v", err)
	}

	router := server.NewHTTPServer()

	bankRepo := catalogpostgres.NewRepository(db)
	beelineRepo := beelinepostgres.NewRepository(db)
	rocketbankRepo := rocketbankpostgres.NewRepository(db)
	rocketbankChequeGenerator := rocketbankcheque.NewService()

	bankCatalogModule := catalog.NewModule(bankRepo)
	bankCatalogModule.RegisterRoutes(router)

	beelineModule := beeline.NewModule(beelineRepo)
	beelineModule.RegisterRoutes(router)

	rocketbankModule := rocketbank.NewModule(rocketbankRepo, rocketbankChequeGenerator)
	rocketbankModule.RegisterRoutes(router)

	if config, err := rocketbankRepo.GetConfig(ctx); err == nil {
		if err := rocketbankChequeGenerator.GenerateMissingSBPTransferCheques(config); err != nil {
			bootstrapLog.Warnf("generate rocketbank cheques: %v", err)
		}
	} else {
		bootstrapLog.Warnf("read rocketbank config for cheque generation: %v", err)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	proxyService, err := proxy.NewService(cfg.Proxy, cfg.Rocketbank, rocketbankRepo)
	if err != nil {
		bootstrapLog.Fatalf("create mitm proxy: %v", err)
	}
	if err := proxyService.Start(); err != nil {
		bootstrapLog.Fatalf("start mitm proxy: %v", err)
	}

	return &App{
		Router: router,
		DB:     db,
		Proxy:  proxyService,
	}
}

func (a *App) Close() {
	if a.Proxy != nil {
		a.Proxy.Close()
	}
	if a.DB != nil {
		a.DB.Close()
	}
}
