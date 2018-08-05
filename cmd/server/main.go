package main

import (
	"flag"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"fmt"
	"net/http"
	"time"

	"database/sql"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/cmd/setting"
	migrate "github.com/infobloxopen/atlas-contacts-app/db"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"os"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
)

func main() {
	doneC := make(chan error)
	logger := NewLogger()

	go func() { doneC <- ServeInternal(logger) }()
	go func() { doneC <- ServeExternal(logger) }()

	if err := <-doneC; err != nil {
		logger.Fatal(err)
	}
}

func init() {
	configFile := flag.String("config", "", "path to config file")
	homePath := flag.String("homepath", "", "path to app install/home path, defaults to working directory")
	flag.Parse()
	err := setting.NewConfigContext(&setting.CommandLineArgs{
		Config:   *configFile,
		HomePath: *homePath,
		Args:     flag.Args(),
	})
	
	if err != nil {
		os.Exit(1)
	}
	
	cmd.LoadConfig()
	
	resource.RegisterApplication(cmd.ApplicationID)
}

func NewLogger() *logrus.Logger {
	logger := logrus.StandardLogger()

	// Set the log level on the default logger based on command line flag
	logLevels := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}
	if level, ok := logLevels[cmd.LogLevel]; !ok {
		logger.Errorf("Invalid value %q provided for log level", cmd.LogLevel)
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}

// ServeInternal builds and runs the server that listens on InternalAddress
func ServeInternal(logger *logrus.Logger) error {
	healthChecker := health.NewChecksHandler("/healthz", "/ready")
	healthChecker.AddReadiness("DB ready check", dbReady)
	healthChecker.AddLiveness("ping", health.HTTPGetCheck(fmt.Sprint("http://", cmd.InternalAddress, "/ping"), time.Minute))

	s, err := server.NewServer(
		// register our health checks
		server.WithHealthChecks(healthChecker),
		// this endpoint will be used for our health checks
		server.WithHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("pong"))
		})),
	)
	if err != nil {
		logger.Errorf("Failed to Start Internal Server %v", err)
		return err
	}
	l, err := net.Listen("tcp", cmd.InternalAddress)
	if err != nil {
		logger.Errorf("Failed to Listen at healthChecker %q %v", cmd.InternalAddress, err)
		return err
	}

	logger.Debugf("serving internal http at %q", cmd.InternalAddress)

	return s.Serve(nil, l)
}

// ServeExternal builds and runs the server that listens on ServerAddress and GatewayAddress
func ServeExternal(logger *logrus.Logger) error {
	dbSQL, err := sql.Open("postgres", cmd.DBConnectionString)
	if err != nil {
		logger.Errorf("Failed to Open DB %s, %v", cmd.DBConnectionString, err)
		return err
	}
	defer dbSQL.Close()
	if err := migrate.MigrateDB(*dbSQL); err != nil {
		logger.Errorf("Failed to Migrate DB %s, %v", cmd.DBConnectionString, err)
		return err
	}
	db, err := gorm.Open("postgres", dbSQL)
	if err != nil {
		logger.Errorf("Failed to Open DB %s, %v", cmd.DBConnectionString, err)
		return err
	}
	defer db.Close()

	grpcServer, err := NewGRPCServer(logger, db)
	if err != nil {
		logger.Errorf("Failed to create GRPC Server %v", err)
		return err
	}

	s, err := server.NewServer(
		// register our grpc server
		server.WithGrpcServer(grpcServer),
		// register the gateway to proxy to the given server address with the service registration endpoints
		server.WithGateway(
			gateway.WithServerAddress(cmd.ServerAddress),
			gateway.WithEndpointRegistration("/v1/", pb.RegisterProfilesHandlerFromEndpoint, pb.RegisterGroupsHandlerFromEndpoint, pb.RegisterContactsHandlerFromEndpoint),
		),
		// serve swagger at the root
		server.WithHandler("/swagger", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.ServeFile(writer, request, cmd.SwaggerFile)
		})),
	)
	if err != nil {
		logger.Errorf("Failed to Start External Server %v", err)
		return err
	}

	// open some listeners for our server and gateway
	grpcL, err := net.Listen("tcp", cmd.ServerAddress)
	if err != nil {
		logger.Errorf("Failed to Listen at http %q %v", cmd.ServerAddress, err)
		return err
	}
	gatewayL, err := net.Listen("tcp", cmd.GatewayAddress)
	if err != nil {
		logger.Errorf("Failed to Listen at gRPC %q %v", cmd.GatewayAddress, err)
		return err
	}

	logger.Debugf("serving gRPC at %q", cmd.ServerAddress)
	logger.Debugf("serving http at %q", cmd.GatewayAddress)

	return s.Serve(grpcL, gatewayL)
}

func dbReady() error {
	db, err := gorm.Open("postgres", cmd.DBConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.DB().Ping()
}
