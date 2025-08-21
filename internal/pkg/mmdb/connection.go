package mmdb

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"moul.io/zapgorm2"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
NewDatabaseConnection creates a new connection to PostgreSQL database based on
the authentication configuration provided as input.
*/
func NewDatabaseConnection(dbHost string, dbUsername string, dbPassword string, dbName string, dbPort int, dbSslMode string, DbLogSlowQueryThresholdSeconds int, appMode string) *gorm.DB {
	zap.L().Info("Start connecting to DB...", zap.String("service", "db-connection"))
	// Set the string connection for the database.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbHost,
		dbUsername,
		dbPassword,
		dbName,
		dbPort,
		dbSslMode,
	)
	// Set logger for the database ORM.
	// It is possible to track slow queries and all the queries performed.
	dbLogger := zapgorm2.New(zap.L())
	dbLogger.SetAsDefault()
	dbLogger.SlowThreshold = time.Duration(DbLogSlowQueryThresholdSeconds) * time.Second
	if appMode == "release" {
		dbLogger.LogLevel = logger.Warn
	} else {
		dbLogger.LogLevel = logger.Info
	}
	// Connect to the database.
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 dbLogger,
	})

	if err != nil {
		zap.L().Error("Connection to DB failed!", zap.String("service", "db-connection"), zap.Error(err))
		panic(err)
	}
	zap.L().Info("Connection to DB done!", zap.String("service", "db-connection"))
	return database
}

/*
CloseDatabaseConnection closes the connection to the database. If the databse is already closed,
logs the error and silences it.
*/
func CloseDatabaseConnection(database *gorm.DB) {
	zap.L().Info("Closing DB connection...", zap.String("service", "db-connection"))
	sqlDB, _ := database.DB()
	err := sqlDB.Close()
	if err != nil {
		zap.L().Error("Closing DB Connection failed!", zap.String("service", "db-connection"), zap.Error(err))
	}
	zap.L().Info("DB connection closed!", zap.String("service", "db-connection"))
}
