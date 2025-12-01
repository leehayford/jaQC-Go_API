package api

import (
	"fmt"
)

const USER_WRITE_ERR = "error creating user record in main database"

func GetUserList() (usrs []User, err error) {
	qry := MDB.Raw(`
		SELECT *
		FROM ` + TBL_USERS + `
	`)
	err = MDB.Scanner(qry, &usrs)
	return
}
func GetUserByID( id int64 ) (usr User, err error) {
	// log.Info("GetUserByID( )...")

	qry := MDB.Raw(`
		SELECT * 
		FROM `+TBL_USERS+`
		WHERE id = ?
		`,
		id,
	)

	if err = MDB.Scanner(qry, &usr); err != nil {
		return
	}

	if usr.ID == 0 {
		err = fmt.Errorf("user with id %d does not exist", id)
		return
	}

	return
}
func GetUserByEMail(email string) (usr User, err error) {
	// log.Info("GetUserByEMail( )...")

	qry := MDB.Raw(`
		SELECT * 
		FROM `+TBL_USERS+`
		WHERE email = ?
		`,
		email,
	)

	if err = MDB.Scanner(qry, &usr); err != nil {
		return
	}

	if usr.ID == 0 {
		err = fmt.Errorf("user with email %s does not exist", email)
		return
	}

	return
}

func (usr *User) UpdateUser(ussn UserSession) (err error) {
	// log.Info("*User) UpdateUser( )...")

	orgUser, err := GetUserByID( usr.ID )
	if err != nil {
		return 
	}

	if usr.Role == ROLE_SUPER || orgUser.Role == ROLE_SUPER {
		return fmt.Errorf("you can't mess with SUPER")
	}

	orgUser.Role = usr.Role
	orgUser.Email = usr.Email
	orgUser.Name = usr.Name
	orgUser.UpdatedBy = ussn.USR.ID
	if res := MDB.Save(&orgUser); res.Error != nil {
		return fmt.Errorf("%s: %s", USER_WRITE_ERR, res.Error.Error())
	}
	// utils.Json("(*User) UpdateUser( ) -> usr : ", orgUser)
	
	TerminateUserSessions(orgUser)
	return
}