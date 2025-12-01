package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"

	"github.com/gofiber/fiber/v2" 
	"github.com/gofiber/fiber/v2/log" 
	"github.com/gofiber/fiber/v2/middleware/logger"

)

const ERR_DB_EXISTS string = "database already exists"
const ERR_DB_CONNSTR_EMPTY string = "database connection string is empty"
const ERR_DB_CONNSTR_MALFORMED string = "database connection string is malformed"
const ERR_DB_NAME_EMPTY string = "database name is empty"
const ERR_FILE_NAME_EMPTY string = "files name is empty"



func ConfigureLogging(app *fiber.App, fileName string) (file *os.File) {

	/* CONFIGURE LOG FILE */
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	logOutput := io.MultiWriter(os.Stdout, file)
	log.SetOutput(logOutput)
	log.SetLevel(log.LevelInfo)
	
	app.Use(logger.New(logger.Config{Output: logOutput}))
	
	log.Info("LOGGING CONFIGURED") 
	return
}


func FormatErr(errMsg, file, fnName string, line int) error {
	return fmt.Errorf("\n\tfile :\t%s\n\tfunc :\t%s\n\tline :\t%d\n\n\t%s\n\n***********", file, fnName, line, errMsg)
}

func LogErr(err error) error {
	pc, file, line, _ := runtime.Caller(1)
	fnName := runtime.FuncForPC(pc).Name()
	log.Error(FormatErr(err.Error(), file, fnName, line))
	return err
}

func LogFatal(err error) {
	pc, file, line, _ := runtime.Caller(1)
	fnName := runtime.FuncForPC(pc).Name()
	log.Fatal(FormatErr(err.Error(), file, fnName, line))
}

func LogChk(msg string) {
	pc, file, line, _ := runtime.Caller(1)
	fnName := runtime.FuncForPC(pc).Name()
	log.Info(FormatErr(msg, file, fnName, line))
}


func Json(name string, v any) {
	js, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		LogErr(err)
	}
	fmt.Printf("\nJSON: %s : %s\n", name, string(js))
}

func Object(obj interface{}) {
	objType := reflect.TypeOf(obj)
	fmt.Println(objType)
	for _, f := range reflect.VisibleFields(objType) {
		fmt.Printf("\n\t%s : %v\n", f.Name, f.Type)
	}
}