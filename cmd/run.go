package cmd

import (
	"fmt"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {

}

var runCMD = &cobra.Command{
	Use:   "run",
	Short: "Run ",
	Long:  `Run `,
	RunE:  runCmdE,
}

func runCmdE(cmd *cobra.Command, args []string) error {
	zlog.Info("setting up run command")
	storeDSN := fmt.Sprintf("user=%s host=%s port=%d database=%s password=%s sslmode=disable TimeZone=Etc/UTC",
		config.Confs.Database.User,
		config.Confs.Database.Host,
		config.Confs.Database.Port,
		config.Confs.Database.Database,
		config.Confs.Database.Password)

	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	// print configs to std
	zlog.Info("configs", zap.Any("configs", config.Confs))

	// Log to file in production mode
	if !config.Confs.Debug {
		if err := initProdLog(gormConfig, &logConfig{
			gorm: true,
			gin:  true,
			zap:  true,
		}); err != nil {
			return err
		}
	}
	db, err := store.NewGormDB(postgres.Open(storeDSN), gormConfig)
	if err != nil {
		zlog.Error("unable to setup store")
		return errors.Newf("unable to setup store: %w", err)
	}


	zlog.Info("setup rbac middleware")
	rbacMiddleware, err := auth.NewRBACMiddleware("configs/auth_model.conf", "configs/routes_policy.csv", config.Confs.Auth0.Namespace) //FIXME
	if err != nil {
		zlog.Error("unable to setup rbac middleware", zap.String("error", err.Error()))
		return errors.Newf("unable to setup rbac middleware: %v", err)
	}

	zlog.Sugar().Infof("storage conf %v\n", config.Confs.Storage)
	if err != nil {
		return errors.Newf("unable to setup storage: %w", err)
	}
	if config.Confs.Debug {
		zlog.Sugar().Infof("setting gin debug mode")
		gin.SetMode(gin.DebugMode)
	} else {
		zlog.Sugar().Infof("setting gin release mode")
		gin.SetMode(gin.ReleaseMode)
	}

	server := server.NewServer(authenticator, rbacMiddleware,  db,
		&config.Confs.OAuthConfig, &server.ServerConfig{})

	listenAddr := config.Confs.Port
	return server.Launch(":"+strconv.Itoa(listenAddr), config.Confs.SSL)

}

func init() {
	RootCmd.AddCommand(runCMD)

	// port
	runCMD.Flags().Uint("port", 5050, "HTTP server listen address")
	viper.BindPFlag("port", runCMD.Flags().Lookup("port"))

	//auth keys
	// TODO: Remove if not needed
	runCMD.Flags().String("auth_pkey", "", "jwt authentication private key")
	runCMD.Flags().String("auth_pubkey", "", "jwt authentication public key")

	//postgres flags
	runCMD.Flags().String("postgres_user", "amir", "Define postgres user")
	viper.BindPFlag("database.user", runCMD.Flags().Lookup("postgres_user"))

	// TODO: Remove if not needed
	runCMD.Flags().String("postgres_connect_name", "", "Define postgres connect name from gcp")

	runCMD.Flags().String("postgres_pwd", "amir123", "Define postgres db password")
	viper.BindPFlag("database.password", runCMD.Flags().Lookup("postgres_pwd"))

	runCMD.Flags().String("postgres_db", "datalead", "Define postgres db name")
	viper.BindPFlag("database.database", runCMD.Flags().Lookup("postgres_db"))

	runCMD.Flags().String("postgres_host", "localhost", "Define postgres host address .e.g localhost")
	viper.BindPFlag("database.host", runCMD.Flags().Lookup("postgres_host"))

	runCMD.Flags().Int("postgres_port", 5432, "Define postgres host address .e.g localhost")
	viper.BindPFlag("database.port", runCMD.Flags().Lookup("postgres_port"))

	// log flags
	runCMD.Flags().String("log_path", "", "Define log path")
	viper.BindPFlag("log_path", runCMD.Flags().Lookup("log_path"))

	// ssl flags
	runCMD.Flags().Bool("ssl", false, "Define ssl")
	viper.BindPFlag("ssl", runCMD.Flags().Lookup("ssl"))

	// debug flags
	runCMD.Flags().Bool("debug", false, "Define debug")

}
