package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"jaQC-Go-API/utils"
)

type MainDatabase struct{ utils.SQLiteClient }

var MDB MainDatabase = MainDatabase{SQLiteClient: utils.SQLiteClient{}}

var JWT utils.JWTConfiguration

var EML utils.EmailConfiguration

/* FILE SYSTEM **************************************************************************/
func ConfigureFileSystem(clean bool) (err error) {
	// log.Info("CONFIGURING FILE SYSTEM...")

	if clean {
		/* ARCHIVE DATA_DIR */
		if err = ArchiveAPIDirectories(); err != nil {
			log.Info(err)
		}
	}

	/* CONFIRM REQUIRED DIRECTORIES EXIST */
	if err = ConfirmAPIDirectories(); err != nil {
		return 
	}

	log.Info("FILE SYSTEM CONFIGURED")
	return
}

/* IF ANY REQURIRED DIRECTORY FAILS TO EXIST, ACTIVELY DISAGREE */
func ConfirmAPIDirectories() (err error) {

	/* ENSURE THE DATA ROOT EXISTS */
	if err = utils.ConfirmDirectory(DATA_DIR); err != nil {
		return utils.LogErr(err)
	}
	log.Info("ConfirmAPIDirectories( ) : ", DATA_DIR)

	/* ENSURE THE ARCHIVE ROOT EXISTS */
	if err = utils.ConfirmDirectory(ARCHIVE_DIR); err != nil {
		return utils.LogErr(err)
	}
	log.Info("ConfirmAPIDirectories( ) : ", ARCHIVE_DIR)

	/* ENSURE THE DEVICE FILES ROOT EXISTS */
	if err = utils.ConfirmDirectory(DEVICE_FILE_DIR); err != nil {
		return utils.LogErr(err)
	}
	log.Info("ConfirmAPIDirectories( ) : ", DEVICE_FILE_DIR)

	/* ENSURE THE DATASET DATABASES ROOT EXISTS */
	if err = utils.ConfirmDirectory(DATASET_DB_DIR); err != nil {
		return utils.LogErr(err)
	}
	log.Info("ConfirmAPIDirectories( ) : ", DATASET_DB_DIR)

	/* ENSURE THE BIN FILES ROOT EXISTS */
	if err = utils.ConfirmDirectory(BIN_FILE_DIR); err != nil {
		return utils.LogErr(err)
	}
	log.Info("ConfirmAPIDirectories( ) : ", BIN_FILE_DIR)

	return
}

func ArchiveAPIDirectories() (err error) {
	// log.Info("ARCHIVING...")

	archive, err := utils.CheckDirectoryExists(DATA_DIR)
	if err != nil {
		return 
	} 

	if archive {
		arc, arc_err := utils.CreateArchiveDirectory(DATA_DIR, ARCHIVE_DIR)
		if arc_err != nil {
			err = arc_err
			return 
		}
	
		if arc_err = utils.ArchiveDirectoryContents(DATASET_DB_DIR, arc+"/"+DATASET_DB_SUB_DIR); arc_err != nil {
			err = arc_err
			return 
		}
	
		if arc_err = utils.ArchiveDirectoryContents(BIN_FILE_DIR, arc+"/"+BIN_FILE_SUB_DIR); arc_err != nil {
			err = arc_err
			return 
		}
	
		if arc_err = utils.ArchiveDirectoryContents(DATA_DIR+"/"+MAIN_DB_NAME, arc+"/"+MAIN_DB_NAME); arc_err != nil {
			err = arc_err
			return 
		}

		log.Info("ARCHIVE API DIRECTORIES : OK : ", arc)
		return
	}


	// log.Info("NOTHING TO ARCHIVE. ")
	return
}

/* MAIN DATABASE **********************************************************************/
func ConfigureMainDatabase(dir, name string, clean bool) (err error) {
	/* MIGRATE MAIN DB & CONNECT -> DISCONNECT MUST BE HANDLED EXPLICITLY */

	if err = utils.ConfirmDirectory(dir); err != nil {
		return
	}

	if MDB.SQLiteClient, err = utils.ConfigureDBSQLiteClient(dir, name); err != nil {
		return
	}

	if err = MDB.Connect(); err != nil {
		return
	}

	/* CREATE OR MIGRATE TABLE MODELS */
	if MDB.ConnectionOK() {
		if err = MDB.AutoMigrate(
	
			/* TABLES */
			User{},
			// Gizmo{},
			// Calibration{},
			// Dataset{},
			// Process{},
			// Variate{},
			Aggregate{},
			// Correlate{},
	
		); err != nil {
			log.Fatal(err)
		}
	} else {
		if err = MDB.Migrator().CreateTable(
	
			/* TABLES */
			User{},
			// Gizmo{},
			// Calibration{},
			// Dataset{},
			// Process{},
			// Variate{},
			Aggregate{},
			// Correlate{},
	
		); err != nil {
			log.Fatal(err)
		}
	}

	if ( clean ) {
		urinp := UserRegistrationInput{ Password: SPR_PW }
		urinp.HashPassword()
		user := User{
			Name: SPR_NAME,
			Email: strings.ToLower(SPR_EMAIL),
			Password: urinp.Password,
			Role: ROLE_SUPER,
		}
		if res := MDB.Create(&user); res.Error != nil {
			utils.LogErr(fmt.Errorf("failed to create user in database: %s", res.Error.Error()))
		}
	}

	log.Info("MAIN DATABASE CONFIGURED : ", MDB.Conn)
	return 
}

var TBL_USERS = (User{}).TableName()
// var TBL_GIZMOS = (Gizmo{}).TableName()
// var TBL_CALS = (Calibration{}).TableName()
// var TBL_DATS = (Dataset{}).TableName()
// var TBL_PROCS = (Process{}).TableName()
// var TBL_VARS = (Variate{}).TableName()
var TBL_AGGS = (Aggregate{}).TableName()
// var TBL_CRLTS = (Correlate{}).TableName()

func ConfigureCORS(app *fiber.App, origins, headers, methods string, cred bool) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     headers,
		AllowMethods:     methods,
		AllowCredentials: cred,
	}))
	log.Info("CORS CONFIGURED")
}

func ConfigureJWT(secret, authType, keyCookie, keyQuery string, accDur, refDur time.Duration) {

	JWT = utils.JWTConfiguration{}
	JWT.Secret = secret
	JWT.AccDur = accDur
	JWT.RefDur = refDur
	JWT.AuthType = authType
	JWT.CookieKey = keyCookie
	JWT.QueryKey = keyQuery

	log.Info("JWT CONFIGURED")
}

func ConfigureEmail(host, port, from, pw string) {

	EML = utils.EmailConfiguration{}
	EML.Host = host
	EML.Port = port
	EML.From = from
	EML.Password = pw

	log.Info("EMAIL CONFIGURED")
}