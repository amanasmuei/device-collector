package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var DbSql *sql.DB

func ConnectMariaDb() error {

	var DATABASE_URI string = "host=timescaledb user=postgres password=Otta2024! dbname=postgres port=5432 sslmode=disable"

	var err error

	fmt.Println("Connection established to DB")
	DB, err = gorm.Open(postgres.Open(DATABASE_URI), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		return err
	}

	// Set the default schema to 'Iot'
	DB.Session(&gorm.Session{FullSaveAssociations: true})
	DB.Debug().Table("public.").Session(&gorm.Session{})

	// Create db object
	DbSql, err = DB.DB()
	if err != nil {
		log.Fatalf("Error getting database instance: %v", err)
		return err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	DbSql.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	DbSql.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	DbSql.SetConnMaxLifetime(time.Hour)

	return nil
}
