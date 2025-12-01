package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2/log"

	"gorm.io/gorm"          // go get gorm.io/gorm
	"gorm.io/gorm/logger"
	"gorm.io/driver/sqlite" // go get gorm.io/driver/sqlite
)

type SQLiteClient struct {
	Conn string
	Segs    int
	*gorm.DB
	*sync.RWMutex
	Connected bool
}

func ConfigureDBSQLiteClient(dir, name string) (dbc SQLiteClient, err error) {

	if dir == "" {
		err = fmt.Errorf("directory name is blank")
		return
	}

	if name == "" {
		err = fmt.Errorf("database name is blank")
		return
	}

	dbc.Conn =  fmt.Sprintf("%s/%s", dir, name)
	dbc.Segs = len(strings.Split(dbc.Conn, "/"))
	dbc.RWMutex = &sync.RWMutex{}
	return
}

/* VALIDATE CLIENT SETTINGS */
// CALLED BY:
// - InitializeDatabaseClient()
// - Connect()
// - Disconnect()
// - ConnectionOK()
func (client *SQLiteClient) ValidateClientSettings() (dbName string, err error) {
	log.Info("\n(*SQLiteClient) ValidateClientSettings()...\n")

	/* CHECK FOR EMPTY CONNECTION STRING */
	if client.Conn == "" {
		err = fmt.Errorf(ERR_DB_CONNSTR_EMPTY)
		return
	}

	/* CHECK FOR CONNECTION STRING SEGMENT COUNT */
	if client.Segs == 0 {
		err = fmt.Errorf("database client was not initialized")
		return
	}

	/* GET CONNECTION STRING SEGMENT COUNT */
	segs := len(strings.Split(client.Conn, "/"))
	if segs < 2 || client.Segs != segs {
		err = fmt.Errorf("%s: %s", ERR_DB_CONNSTR_MALFORMED, client.Conn)
		return
	}

	dbName = strings.Split(client.Conn, "/")[client.Segs-1]
	return
}

func (client *SQLiteClient) Connect() (err error) {

	db_name, err := client.ValidateClientSettings()
	if err != nil {
		return
	}
	// fmt.Printf("\n(*SQLiteClient) Connect() -> db_name: %s \n", db_name)

	if client.DB, err = gorm.Open(sqlite.Open(client.Conn), &gorm.Config{}); err != nil {
		log.Info("(*SQLiteClient) Connect() -> FAILED :", db_name)
		return err
	}

	// client.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"") //
	client.DB.Logger = logger.Default.LogMode(logger.Error)

	client.Connected = true
	// fmt.Printf("\n(*SQLiteClient) Connect() -> %s -> connected... \n", db_name)
	return
}

func (client *SQLiteClient) Disconnect() (err error) {

	db_name, err := client.ValidateClientSettings()
	if err != nil {
		return
	} // fmt.Printf("\n(*SQLiteClient) Disconnect() -> db_name: %s \n", db_name)

	db, err := client.DB.DB()
	if err != nil {
		return err
	}

	if err = db.Close(); err != nil {
		log.Info("(*SQLiteClient) Disconnect() -> FAILED :", db_name)
		return err
	}

	client.Connected = false
	// fmt.Printf("\n(*SQLiteClient) Disconnect() -> %s -> connection closed. \n", db_name)
	return
}

func (client *SQLiteClient) ConnectionOK() bool {

	if client.Connected {
		
		if _, err := client.ValidateClientSettings(); err != nil {
			client.Connected = false
			return client.Connected
		}
	
		/* CHECK IF SQLITE DATABASE FILE EXISTS */
		_, err := os.Stat(client.Conn)
		client.Connected = !os.IsNotExist(err)
	} 

	return client.Connected
}

func (client *SQLiteClient) Scanner(qry *gorm.DB, out interface{}) (err error) {

	rows, err := qry.Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		client.ScanRows(rows, out)
	}

	return
}

func (client *SQLiteClient) InitializeDatabaseClient(conn_str string, segs int) (err error) {

	err_msg := "failed to initialize database : "

	/* SET CONNECTION STRING */
	client.Conn = conn_str

	/* SET CONNECTION STRING SEGMENT COUNT */
	client.Segs = len(strings.Split(conn_str, "/"))

	/* VALIDATE CLIENT AND GET DB NAME */
	db_name, err := client.ValidateClientSettings()
	if err != nil {
		err = fmt.Errorf("%s name -> %s: %s", err_msg, db_name, err.Error())
		return
	}
	fmt.Printf("\n(*SQLiteClient) InitializeDatabaseClient() -> ValidateClientSettings() -> %s -> OK \n", db_name)

	/* CONFIRM CONNECTION */
	if err = client.Connect(); err != nil {
		err = fmt.Errorf("%s: %s", err_msg, err.Error())
		return
	}
	fmt.Printf("\n(*SQLiteClient) InitializeDatabaseClient() -> Connect() -> %s -> OK \n", db_name)

	/* CONFIRM DISCONNECT */
	if err = client.Disconnect(); err != nil {
		err = fmt.Errorf("%s: %s", err_msg, err.Error())
		return
	}
	fmt.Printf("\n(*SQLiteClient) InitializeDatabaseClient() -> Disconnect() -> %s -> OK \n", db_name)

	log.Info("(*SQLiteClient) -> OK ", db_name)
	return
}