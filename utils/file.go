package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"
	"github.com/gofiber/fiber/v2/log" 
)

func ConfirmDirectory(name string) (err error) {
	
	if err = os.Mkdir(name, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			/* THERE'S SOME OTHER ISSUE */
			return LogErr(err)
		}
		err = nil
	} // log.Info("DIRECTORY CONFIRMED : ", name)
	return
}

func CheckDirectoryExists(dir string) (exists bool, err error) {

	exists = true
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			exists = false
			err = nil
			// log.Info("DIRECTORY DOES NOT EXIST : ", dir)
			return
		} else {
			/* THERE'S SOME OTHER ISSUE */
			log.Info(`DIRECTORY CHECK ERROR : `+dir+` : `, err)
			return 
		}
	} 
	// log.Info("DIRECTORY EXISTS : ", dir)
	return
}

/* TIME OF ARCHIVING - YYYYMMDD_HHmmss */
func CreateArchiveDirectoryName() (arc_dir string) {
	now := time.Now().UTC()
	return fmt.Sprintf(
		"%d%02d%02d_%02d%02d%02d", 
		now.Year(), 
		int(now.Month()), 
		now.Day(), 
		now.Hour(), 
		now.Minute(), 
		now.Second(),
	)
}

func CreateArchiveDirectory(dataDir, archiveDir string) (arc_dir string, err error) {
	
	/* TIME OF ARCHIVING */
	arc_dir = archiveDir + "/" + CreateArchiveDirectoryName()

	if err = os.Mkdir(arc_dir, os.ModePerm); err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		/* THERE'S SOME OTHER ISSUE */
		err = LogErr(err)
		return 
	}
	// log.Info(fmt.Sprintf("ARCHIVE CREATED : %s/..., %s ", dataDir, arc_dir))
	return
}

func ArchiveDirectoryContents(dir, arc string) (err error) {

	/* ONLY ARCHIVE dir IF IT EXISTS */
	archive, err := CheckDirectoryExists(dir)
	if err != nil {
		return 
	} 	

	if archive {
		/* dir EXISTS, ARCHIVE IT */
		if err = os.Rename(dir, arc); err != nil {
			return LogErr(err)
		}
		// log.Info(`ARCHIVED : `, arc)
	}
	return
}


/* BINARY FILES **********************************************************************/


/* RETURNS ALL BYTES FROM POSTED BIN FILE */
func BytesFromBin(file *multipart.FileHeader) (bytes []byte, err error) {
	
	content, err := file.Open()
	if err != nil {
		err = fmt.Errorf("failed to open file: %s", err.Error())
		return
	}

	bytes, err = io.ReadAll(content)
	if err != nil {
		err = fmt.Errorf("failed to read file content: %s", err.Error())
		return
	}

	return
}

// const ERR_DB_EXISTS string = "database already exists"
// const ERR_DB_CONNSTR_EMPTY string = "database connection string is empty"
// const ERR_DB_CONNSTR_MALFORMED string = "database connection string is malformed"
// const ERR_DB_NAME_EMPTY string = "database name is empty"
// const ERR_FILE_NAME_EMPTY string = "files name is empty"

/* APPENDS MODEL HEX VALUES TO ~/BIN_FILE_DIR/dirName/fileName.bin */
func WriteModelBytesToBinFile(defDataDir, subDir, fileName string, buf []byte) (err error) {
	if fileName == "" {
		return LogErr(fmt.Errorf(ERR_FILE_NAME_EMPTY))
	}
	dir := fmt.Sprintf("%s/%s", defDataDir, subDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		LogErr(err)
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s.bin", dir, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return LogErr(err)
	}
	defer file.Close()

	_, err = file.Write(buf)
	if err != nil {
		return LogErr(err)
	}

	return
}

/* RETURNS ALL BYTES FROM ~/BIN_FILE_DIR/dirName/fileName.bin */
func ReadModelBytesFromBinFile(defDataDir, subDir, fileName string) (buf []byte, err error) {

	if fileName == "" {
		return nil, LogErr(fmt.Errorf(ERR_FILE_NAME_EMPTY))
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s/%s.bin", defDataDir, subDir, fileName), os.O_RDONLY, 0600)
	if err != nil {
		return nil, LogErr(err)
	}
	defer file.Close()

	buf, err = io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("failed to read file content: %s", err.Error())
		return
	}
	return
}