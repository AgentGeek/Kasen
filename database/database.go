package database

import (
	_ "embed"

	"database/sql"
	"fmt"

	"kasen/config"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs1703/logger"
)

type Database struct {
	*sql.DB
}

var ReadDB *Database
var WriteDB *Database

//go:embed schema.sql
var schema []byte

func init() {
	cfg := config.GetDatabase()
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Passwd, cfg.SSLMode)

	readConn, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Err.Fatalln(err)
	}

	if err := readConn.Ping(); err != nil {
		logger.Err.Fatalln(err)
	}

	writeConn, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Err.Fatalln(err)
	}

	if err := writeConn.Ping(); err != nil {
		logger.Err.Fatalln(err)
	}

	if _, err = writeConn.Exec(string(schema)); err != nil && err != sql.ErrNoRows {
		logger.Err.Fatalln(err)
	}

	ReadDB = &Database{readConn}
	WriteDB = &Database{writeConn}
}
