package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	databasewrapper "github.com/bankease/payment-service/server/lib/database/wrapper"
	_ "github.com/lib/pq"
)

type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
	SslMode      string
	TimeZone     string
	MaxRetry     int
	Timeout      time.Duration
}

type DbSql struct {
	SqlDb      *sql.DB
	SqlTx      *sql.Tx
	driverName string
	count      int
	Config     Config
	Dbw        databasewrapper.DatabaseInterface
	Conn       databasewrapper.DatabaseConnectionInterface
}

func InitConnectionDB(driverName string, config Config, dbw databasewrapper.DatabaseInterface) *DbSql {
	return &DbSql{
		driverName: driverName,
		count:      0,
		Config:     config,
		Dbw:        dbw,
	}
}

func (ds *DbSql) GetTimeout() time.Duration {
	return ds.Config.Timeout
}

func (ds *DbSql) AddCounter() {
	ds.count++
}

func (ds *DbSql) Connect() error {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		ds.Config.Host,
		ds.Config.Port,
		ds.Config.User,
		ds.Config.Password,
		ds.Config.DatabaseName,
		ds.Config.SslMode,
		ds.Config.TimeZone)

	var errDb error
	ds.SqlDb, errDb = ds.Dbw.Open(ds.driverName, connString)
	if errDb != nil {
		return errDb
	}
	ds.Conn = ds.SqlDb
	return nil
}

func (ds *DbSql) GetPmConnection() *sql.DB {
	return ds.SqlDb
}

func (ds *DbSql) CheckConnection() error {
	if ds.count > 0 {
		log.Println("server is still trying to connect to DB")
	}
	if err := ds.Conn.Ping(); err != nil {
		ds.ClosePmConnection()
		return ds.TryConnect()
	}
	return nil
}

func (ds *DbSql) TryConnect() error {
	for {
		ds.AddCounter()
		log.Printf("trying to connect %v times....", ds.count)

		err := ds.ConnectionDB()
		if err == nil {
			ds.count = 0
			return nil
		}

		if ds.count >= ds.Config.MaxRetry {
			log.Println("stop reconnecting max retries exceeded")
			ds.count = 0
			return err
		}
	}
}

func (ds *DbSql) ConnectionDB() error {
	return ds.Connect()
}

func (ds *DbSql) ClosePmConnection() error {
	if ds.Conn != nil {
		return ds.Conn.Close()
	}
	return errors.New("database connection already closed")
}

func (ds *DbSql) SetMaxIdleConns(n int) {
	ds.Conn.SetMaxIdleConns(n)
}

func (ds *DbSql) SetMaxOpenConns(n int) {
	ds.Conn.SetMaxOpenConns(n)
}

func (ds *DbSql) StartTransaction() error {
	var trxErr error
	ds.SqlTx, trxErr = ds.Conn.Begin()
	return trxErr
}
