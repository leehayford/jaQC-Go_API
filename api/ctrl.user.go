package api

import (
	"fmt"
	"strings"
	"sync"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"     // go get github.com/google/uuid
	"golang.org/x/crypto/bcrypt" // go get golang.org/x/crypto/bcrypt

	"jaQC-Go-API/utils"
)

const ROLE_SUPER string = "super"
const ROLE_ADMIN string = "admin"
const ROLE_OPERATOR string = "operator"
const ROLE_VIEWER string = "viewer"

const AUTH_MSG_SUPER = "you must be a super admin to perform this action"
const AUTH_MSG_ADMIN = "you must be an admin to perform this action"
const AUTH_MSG_OPERATOR = "you must be an oporator to perform this action"
const AUTH_MSG_VIEWER = "you must be a viewer to perform this action"

func RoleCheckSuper(c *fiber.Ctx) (err error) {
	role := c.Locals("role").(string)
	if role == ROLE_SUPER {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).SendString(AUTH_MSG_SUPER)
}
func RoleCheckAdmin(c *fiber.Ctx) (err error) {
	role := c.Locals("role").(string)
	if role == ROLE_SUPER || role == ROLE_ADMIN {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).SendString(AUTH_MSG_ADMIN)
}
func RoleCheckOperator(c *fiber.Ctx) (err error) {
	role := c.Locals("role").(string)
	if role == ROLE_SUPER || role == ROLE_ADMIN || role == ROLE_OPERATOR {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).SendString(AUTH_MSG_OPERATOR)
}
func RoleCheckViewer(c *fiber.Ctx) (err error) {
	role := c.Locals("role").(string)
	if role == ROLE_SUPER || role == ROLE_ADMIN || role == ROLE_OPERATOR || role == ROLE_VIEWER {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).SendString(AUTH_MSG_VIEWER)
}

/* SAFE RESPONSE DATA */
func (user *User) FilterUserRecord() UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}


/* USER REGISTRATION */
func (urinp *UserRegistrationInput) HashPassword() (err error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(urinp.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %s", err.Error())
	}

	urinp.Password = string(hash)
	return
}
func (urinp *UserRegistrationInput) RegisterUser(c *fiber.Ctx) (err error) {
	fmt.Printf("RegisterUser( )\n")

	if urinp.Password != urinp.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).SendString("passwords do not match")
	}

	if err = urinp.HashPassword(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	user := User{}
	user.Name = urinp.Name
	user.Email = strings.ToLower(urinp.Email)
	user.Password = urinp.Password
	user.Role = ROLE_VIEWER

	if res := MDB.Create(&user); res.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(
			fmt.Sprintf("failed to create user in database: %s", res.Error.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
}
func (urinp *UserRegistrationInput) UpdatePassword(user User) (err error) {
	// fmt.Printf("UpdatePassword( )\n")

	if urinp.Password != urinp.PasswordConfirm {
		err = fmt.Errorf("passwords do not match")
		return
	}

	if err = urinp.HashPassword(); err != nil {
		return 
	}

	user.Password = urinp.Password
	// fmt.Printf("urinp.Password : %s\n", urinp.Password)

	user.UpdatedBy = user.ID
	user.UpdatedAt = time.Now().UTC().UnixMilli()
	if res := MDB.Save(&user); res.Error != nil {
		err = fmt.Errorf("failed to update user in database: %s", res.Error.Error())
		return
	}

	TerminateUserSessions(user)
	return 
}

type PWResetSession struct {
	Code string `json:"code"`
	Email string `json:"email"`
	Expire int64 `json:"exp"`
}
type PWResetSessionMap map[string]PWResetSession
var PWResetCodes = make(PWResetSessionMap)
var PWResetCodesRWMutex = sync.RWMutex{}

func PWResetMapWrite(code string, rs PWResetSession) (err error) {
	PWResetCodesRWMutex.Lock()
	PWResetCodes[code] = rs
	PWResetCodesRWMutex.Unlock()
	return
}
func PWResetMapRead(code string) (rs PWResetSession, err error) {
	/* REMOVE ALL EXPIRED */
	PWResetMapClearExpired()
	PWResetCodesRWMutex.Lock()
	rs = PWResetCodes[code]
	PWResetCodesRWMutex.Unlock()
	// utils.Json("PWResetMapRead( ) -> rs : ", rs)

	if rs.Expire == 0 {
		err = fmt.Errorf("invalid reset code")
	}
	return
}
func PWResetRemove(code string) {
	// log.Info("PWResetRemove() ", code)
	PWResetCodesRWMutex.Lock()
	delete(PWResetCodes, code)
	PWResetCodesRWMutex.Unlock()
}
func PWResetMapClearExpired() {
	now := time.Now().UTC().UnixMilli()
	exp := []string{}
	PWResetCodesRWMutex.Lock()
	for c, r := range PWResetCodes {
		// utils.Json(`PWResetMapClearExpired( ` + c + ` ) -> r : `, r)
		if r.Expire < now {
			exp = append(exp, c)
			// fmt.Println("expired code : ", c)
		}
	}
	for _, c := range exp {
		delete(PWResetCodes, c)
		// fmt.Println("removed code : ", c)
	}
	PWResetCodesRWMutex.Unlock()
}

func (urinp *UserRegistrationInput) FrogotPassword(c *fiber.Ctx) (err error) {

	/* CHECK FOR USER WITH GIVEN EMAIL */
	user := User{}
	MDB.Where("email = ?", urinp.Email).Last(&user)
	if user.Email != urinp.Email {
		c.Status(fiber.StatusNotFound).SendString("user does not exist")
	} 
	// utils.Json("found user: ", user)

	/* GENERATE TEMP SESSION */
	code := strings.Split(uuid.New().String(), "-")[4]
	pwrs := PWResetSession{
		Code: code,
		Email: user.Email,
		Expire: time.Now().Add(time.Minute * 2).UnixMilli(),
	}
	// fmt.Println("FrogotPassword( ) -> code : ", code)
	PWResetMapWrite(code, pwrs)

	if pwrs, err = PWResetMapRead(pwrs.Code); err != nil {
		return
	}
	// fmt.Println("FrogotPassword( ) -> MapRead - > code : ", pwrs.Code)
	
	/* SEND EMAIL */
	tplt_vars := struct {
		Expire, Code string
	}{
		Expire: time.UnixMilli(pwrs.Expire).Format("2006-01-02 15:04:05"),
		Code:  pwrs.Code,
	}
	if err = EML.SendHTML(
		[]string{urinp.Email},
		"templates/confirm_pw_reset.html",
		"Confirm Password Reset",
		tplt_vars,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	TerminateUserSessions(user)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Check your email."})
}
func (urinp *UserRegistrationInput) ResetPassword(code string) (err error) {

	code = strings.Trim(code, "\"")

	// fmt.Println("ResetPassword( ) -> code : ", code)
	pwrs, err := PWResetMapRead(code)
	if err != nil {
		return
	}
	// utils.Json("pw reset session : ", pwrs)

	/* GET USER FROM AUTH DATA */
	user, err := GetUserByEMail(pwrs.Email)
	if err != nil {
		return
	}
	utils.Json("found user: ", user)

	if err = urinp.UpdatePassword(user); err != nil {
		return
	}

	PWResetRemove(code)
	return
}

