package database

import (
	"testing"
	"time"

	databasemock "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database/mock"
	databasewrapper "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database/wrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConnectionDB(t *testing.T) {
	config := Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "test",
		Password:     "test",
		DatabaseName: "testdb",
		SslMode:      "disable",
		TimeZone:     "UTC",
		MaxRetry:     3,
		Timeout:      10 * time.Second,
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	assert.NotNil(t, dbSql)
	assert.Equal(t, 10*time.Second, dbSql.GetTimeout())
}

func TestConnect_Success(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	err := dbSql.Connect()
	require.NoError(t, err)
	assert.NotNil(t, dbSql.GetPmConnection())
}

func TestConnect_EmptyDriverName(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("", config, &databasemock.DatabaseMock{})
	err := dbSql.Connect()
	assert.Error(t, err)
}

func TestClosePmConnection_NoConnection(t *testing.T) {
	config := Config{}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	err := dbSql.ClosePmConnection()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already closed")
}

func TestClosePmConnection_WithConnection(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	// MockConn exists, Close should not error
	dbSql.Conn = &databasemock.DatabaseMock{} // set a mock conn
	err := dbSql.ClosePmConnection()
	assert.NoError(t, err)
}

func TestCheckConnection_PingSuccess(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
		MaxRetry: 1, Timeout: 1 * time.Second,
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	mockConn := &databasemock.DatabaseMock{DbPq: dbSql.SqlDb}
	dbSql.Conn = mockConn
	err := dbSql.CheckConnection()
	assert.NoError(t, err)
}

func TestAddCounter(t *testing.T) {
	config := Config{}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	assert.Equal(t, 0, dbSql.count)
	dbSql.AddCounter()
	assert.Equal(t, 1, dbSql.count)
}

func TestSetMaxIdleConns(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	dbSql.Conn = &databasemock.DatabaseMock{}
	// Should not panic
	dbSql.SetMaxIdleConns(5)
	dbSql.SetMaxOpenConns(10)
}

func TestStartTransaction(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	dbSql.Conn = &databasemock.DatabaseMock{}
	err := dbSql.StartTransaction()
	assert.NoError(t, err)
}

func TestDatabaseWrapper_Open(t *testing.T) {
	wrapper := &databasewrapper.DatabaseWrapper{}
	// This will try to open a real connection which fails, but shouldn't panic
	db, err := wrapper.Open("postgres", "invalid-connection-string")
	// sql.Open doesn't actually connect, just validates driver
	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Close()
}

func TestConnectionDB(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	err := dbSql.ConnectionDB()
	assert.NoError(t, err)
}

func TestCheckConnection_PingFail_TryConnect(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
		MaxRetry: 1, Timeout: 1 * time.Second,
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	// Simulate ping failure by setting a nil DbPq mock
	dbSql.Conn = &databasemock.DatabaseMock{DbPq: nil}
	// CheckConnection should try to reconnect; TryConnect will succeed on mock
	err := dbSql.CheckConnection()
	// With mock, reconnect succeeds
	assert.NoError(t, err)
}

func TestTryConnect_MaxRetryExceeded(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
		MaxRetry: 1, Timeout: 1 * time.Second,
	}
	// Use empty driver name to force Connect failure
	dbSql := InitConnectionDB("", config, &databasemock.DatabaseMock{})
	err := dbSql.TryConnect()
	assert.Error(t, err)
}

func TestCheckConnection_AlreadyReconnecting(t *testing.T) {
	config := Config{
		Host: "localhost", Port: "5432", User: "test", Password: "test",
		DatabaseName: "testdb", SslMode: "disable", TimeZone: "UTC",
		MaxRetry: 3, Timeout: 1 * time.Second,
	}
	dbSql := InitConnectionDB("postgres", config, &databasemock.DatabaseMock{})
	dbSql.Connect()
	dbSql.count = 1 // simulate already reconnecting
	mockConn := &databasemock.DatabaseMock{DbPq: dbSql.SqlDb}
	dbSql.Conn = mockConn
	err := dbSql.CheckConnection()
	assert.NoError(t, err)
}
