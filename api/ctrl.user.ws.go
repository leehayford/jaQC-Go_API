package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2" // go get github.com/gofiber/websocket/v2
	"github.com/gofiber/fiber/v2/log"

	"github.com/google/uuid"     // go get github.com/google/uuid
	"golang.org/x/crypto/bcrypt" // go get golang.org/x/crypto/bcrypt

	"jaQC-Go-API/utils"
)

const WS_PING_DUR = time.Second * 30
const WS_MAX_ERR = int64(10)
const WS_MIN_ERR_SEC = 3

type UserSession struct {
	SID    uuid.UUID    `json:"sid"`
	REFTok string       `json:"ref_token"`
	ACCTok string       `json:"acc_token"`
	USR    UserResponse `json:"user"`

	Connected bool `json:"-"`
	RWMChan   *sync.RWMutex  `json:"-"`
	DataOut chan string `json:"-"`
	WSClosedByClient chan struct{} `json:"-"`
	WSSendErrorLimit chan struct{} `json:"-"`
	WSReceiveErrorLimit chan struct{} `json:"-"`
	CloseWSListen chan struct{} `json:"-"`
	CloseWSSend chan struct{} `json:"-"`
}

type UserSessionMap map[string]UserSession
var UserSessionsMap = make(UserSessionMap)
var UserSessionsMapRWMutex = sync.RWMutex{}

func UserSessionsMapWrite(u UserSession) (err error) {
	// log.Info("UserSessionsMapWrite() ", u)
	sid := u.SID.String()
	if !utils.ValidateUUIDString(sid) {
		err = utils.LogErr(fmt.Errorf("invalid user session ID %s", u.SID))
		return
	}

	UserSessionsMapRWMutex.Lock()
	UserSessionsMap[sid] = u
	UserSessionsMapRWMutex.Unlock()
	return
}
func UserSessionsMapRead(sid string) (u UserSession, err error) {
	// log.Info("UserSessionsMapRead() ", sid)
	UserSessionsMapRWMutex.Lock()
	u = UserSessionsMap[sid]
	UserSessionsMapRWMutex.Unlock()

	if !utils.ValidateUUIDString(u.SID.String()) {
		// err = utils.LogErr(fmt.Errorf("user session not found; please log in"))
		err = fmt.Errorf("user session not found; please log in")
	}
	return
}
func UserSessionsMapCopy() (usm UserSessionMap) {
	UserSessionsMapRWMutex.Lock()
	usm = UserSessionsMap
	UserSessionsMapRWMutex.Unlock()
	return
}
func UserSessionsMapRemove(usid string) {
	// log.Info("UserSessionsMapRemove() ", usid)
	UserSessionsMapRWMutex.Lock()
	delete(UserSessionsMap, usid)
	UserSessionsMapRWMutex.Unlock()
}

/* AUTHENTICATE USER INPUT AND RETURN JWTs */
func LoginUser(ulinp UserLoginInput) (ussn UserSession, err error) {
	// log.Info("LoginUser( )")

	user := User{}
	/* CHECK EMAIL */
	res := MDB.First(&user, "email = ?", strings.ToLower(ulinp.Email))
	if res.Error != nil {
		/* log to file only */ log.Info(fmt.Sprintf("LOGIN FAILED; BAD EMAIL : %s", strings.ToLower(ulinp.Email)))
		err = fmt.Errorf("invalid email or password")
		return
	}
	// utils.Json("LoginUser() -> user:", user)

	/* CHECK PASSWORD */
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(ulinp.Password)); err != nil {
		/* log to file only */ log.Info(fmt.Sprintf("LOGIN FAILED; BAD PASSWORD : %s", strings.ToLower(ulinp.Email)))
		err = fmt.Errorf("invalid email or password")
		return
	}
	// log.Info("LoginUser() -> hashed pw:", user.Password)

	/* CREATE A USER SESSION ID */
	ussn.SID = uuid.New()
	// log.Info("LoginUser() -> ussn.SID:", ussn.SID)

	/*  FILTER USER DATA */
	ussn.USR = user.FilterUserRecord() // Json("LoginUser() -> user session:", us)
	// utils.Json("LoginUser() -> ussn.USR:", ussn.USR)

	/* CREATE REFRESH TOKEN*/
	if err = ussn.CreateRefreshToken(); err != nil {
		utils.LogErr(err)
		return
	}
	// log.Info("LoginUser() -> ussn.REFTok:", ussn.REFTok)

	/* CREATE ACCESS TOKEN */
	if err = ussn.CreateAccessToken(); err != nil {
		utils.LogErr(err)
		return
	}
	// log.Info("LoginUser() -> ussn.ACCTok:", ussn.ACCTok)

	/* UPDATE USER SESSION MAP */
	
	ussn.RWMChan = &sync.RWMutex{}
	ussn.CloseWSListen = make(chan struct{})
	ussn.CloseWSSend = make(chan struct{})
	ussn.WSClosedByClient = make(chan struct{})
	ussn.WSReceiveErrorLimit = make(chan struct{})
	ussn.WSSendErrorLimit = make(chan struct{})
	ussn.DataOut = make(chan string)
	err = UserSessionsMapWrite(ussn)

	/* log to file only */ log.Info(fmt.Sprintf("LOGIN SUCCESS : %s", strings.ToLower(ulinp.Email)))
	return
}

/* AUTHENTICATION MIDDLEWARE */
func GetAuthUserSession(sid string) (ussn UserSession, err error) {
	// log.Info("GetAuthUserSession( )")

	ussn, err = UserSessionsMapRead(sid)
	if err != nil {
		return
	} 
	// utils.Json("GetAuthUserSession( ) -> ussn: ", ussn)

	return
}

/* REMOVES THE SESSION FOR GIVEN USER FROM UserSessionsMap */
func LogoutUser(ussn UserSession) {
	log.Info("LogoutUser( )")

	UserSessionsMapRemove(ussn.SID.String())
}
/* REMOVES ALL SESSIONS FOR GIVEN USER FROM UserSessionsMap */
func TerminateUserSessions(usr User) (count int) {

	sess := UserSessionsMapCopy()

	count = 0
	for sid, us := range sess {
		if us.USR.ID == usr.ID {
			UserSessionsMapRemove(sid)
			count++
		}
	}

	return
}


func (ussn *UserSession) ValidatePostRequestBody(c *fiber.Ctx) (err error) {

	if err = utils.ParseRequestBody(c, ussn); err != nil {
		return
	}

	if !utils.ValidateUUIDString(ussn.SID.String()) {
		err = fmt.Errorf("invalid session ID: %s", ussn.SID.String())
		return
	}

	*ussn, err = UserSessionsMapRead(ussn.SID.String())

	return
}

/* CREATE REFRESH TOKEN*/
func (ussn *UserSession) CreateRefreshToken() (err error) {
	// log.Info("(*UserSession) CreateRefreshToken( )")
	if ussn.REFTok, err = JWT.CreateRefreshToken(ussn.USR.ID); err != nil {
		return utils.LogErr(fmt.Errorf("refresh token generation failed: %s", err.Error()))
	}
	// log.Info("(*UserSession) CreateRefreshToken( ) -> ussn.REFTok : ", ussn.REFTok)
	return
}

/* CREATE ACCESS TOKEN*/
func (ussn *UserSession) CreateAccessToken() (err error) {
	// log.Info("(*UserSession) CreateAccessToken( )")
	if ussn.ACCTok, err = JWT.CreateAccessToken(ussn.USR.ID, ussn.USR.Role); err != nil {
		return utils.LogErr(fmt.Errorf("access token generation failed: %s", err.Error()))
	}
	// log.Info("(*UserSession) CreateAccessToken( ) -> ussn.ACCTok : ", ussn.ACCTok)
	return
}

/* CREATES A NEW ACCESS TOKEN IF REFRESH TOKEN HAS NOT EXPIRED */
func (ussn *UserSession) RefreshAccessToken() (err error) {
	// log.Info("RefreshAccessToken( )")

	/* GET USER FROM SESSION MAP */
	mus, err := UserSessionsMapRead(string(ussn.SID.String()))
	if err != nil {
		return
	} 
	// utils.Json("RefreshAccessToken( ) -> UserSessionsMapRead( ) -> mus: ", mus)

	/* CHECK REFRESH TOKEN EXPIRE DATE IN MAPPED USER SESSION. IF TIMEOUT, DENY */
	ref_claims, err := JWT.ClaimsFromTokenString(mus.REFTok)
	if err != nil {
		return 
	}
	exp := 0
	now := int(time.Now().Unix())
	if fExp, ok := ref_claims["exp"].(float64); ok {
		exp = int(fExp)
	} 
	// log.Info("RefreshAccessToken( ) -> exp: %d", exp) 
	// log.Info("RefreshAccessToken( ) -> now: %d", now)

	if exp < now {
		return fmt.Errorf("authorization failed; your refresh token has expired; please log in")
	}

	if err = ussn.CreateAccessToken(); err != nil {
		return
	}

	return UserSessionsMapWrite(*ussn)
}

func (ussn *UserSession) WSConnect(ws *websocket.Conn) {

	ussn.RWMChan = &sync.RWMutex{}
	ussn.CloseWSListen = make(chan struct{})
	ussn.CloseWSSend = make(chan struct{})
	ussn.WSClosedByClient = make(chan struct{})
	ussn.WSReceiveErrorLimit = make(chan struct{})
	ussn.WSSendErrorLimit = make(chan struct{})
	ussn.DataOut = make(chan string)
	ussn.Connected = true

	if err := UserSessionsMapWrite(*ussn); err != nil {
		utils.LogErr(err)
	}

	go ussn.WSListenForMessages(ws)

	go ussn.WSRunMessageSender(ws)

	nextPing := time.Now().UTC().Add(WS_PING_DUR).UnixMilli()
	// conn := true
	for ussn.Connected  {
		select {

		case <- ussn.WSClosedByClient:
			log.Info("WSConnect() -> ussn.WSClosedByClient -> CLOSING...")
			ussn.Connected  = false

		case <- ussn.WSReceiveErrorLimit:
			log.Info("WSConnect() -> ussn.WSReceiveErrorLimit -> CLOSING...")
			ussn.Connected  = false
		
		case <- ussn.WSSendErrorLimit:
			log.Info("WSConnect() -> ussn.WSSendErrorLimit -> CLOSING...")
			ussn.Connected  = false

		default:
			if time.Now().UTC().UnixMilli() > nextPing {

				nextPing = time.Now().UTC().Add(WS_PING_DUR).UnixMilli()
				
				if err := ussn.WSSendMessage("live", time.Now().UTC()); err != nil {
					log.Info("WSConnect() -> ERROR SENDING PING : ", err.Error())
				}

			}
		}
	}

	if ussn.CloseWSListen != nil {
		ussn.CloseWSListen <- struct{}{}
		close(ussn.CloseWSListen)
		ussn.CloseWSListen = nil
	}

	if ussn.CloseWSSend != nil {
		ussn.CloseWSSend <- struct{}{}
		close(ussn.CloseWSSend)
		ussn.CloseWSSend = nil
	}

	if ussn.DataOut != nil {
		close(ussn.DataOut)
		ussn.DataOut = nil
	}

	if err := UserSessionsMapWrite(*ussn); err != nil {
		utils.LogErr(err)
	}

	log.Info("WSConnect() -> CLOSED.")
}


func MaxWSError(start, count int64) (outStart, outCount int64, limit bool) {
	outCount = count+1
	now := time.Now().UTC().Unix()
	
	/* ENOUGH TIME HAS PASSED SINCE THE LAST ERROR */
	clearCount := ( start - now ) > WS_MIN_ERR_SEC
	if clearCount {
		limit = false
		outStart = now
		outCount = 0
		return
	} 
	
	/* WE MIGHT BE GETTING TOO MANY ERRORS */
	limit = outCount > WS_MAX_ERR

	return
}

func (ussn *UserSession) WSListenForMessages(ws *websocket.Conn) {

	start := time.Now().UTC().Unix()
	count := int64(0)
	limit := false

	listen := true
	for listen {

		select {

		case <- ussn.CloseWSListen:
			// log.Info("WSListenForMessages() -> CLOSING.")
			listen = false

		default:
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if strings.Contains(err.Error(), "close") {
					log.Info("error reading websocket message: ", err.Error())
					msg = []byte("close")
				} else {
					log.Info("error reading websocket message: ", err.Error())
					if start, count, limit = MaxWSError(start, count); limit {
						log.Error("CLOSING WS CONNECTION; MAX RECEIVE ERRORS")
						ussn.WSReceiveErrorLimit <- struct{}{}
					}
				}
			}

			if string(msg) == "close" {
				ussn.WSClosedByClient <- struct{}{}
			}
		}

	}

	if ussn.WSClosedByClient != nil {
		close(ussn.WSClosedByClient)
		ussn.WSClosedByClient = nil
	}

	if ussn.WSReceiveErrorLimit != nil {
		close(ussn.WSReceiveErrorLimit)
		ussn.WSReceiveErrorLimit = nil
	}// log.Info("WSListenForMessages() -> STOPPED.")
}

func (ussn *UserSession) WSRunMessageSender(ws *websocket.Conn) {

	start := time.Now().UTC().Unix()
	count := int64(0)
	limit := false

	send := true
	for send {

		select {

		case <- ussn.CloseWSSend:
			// log.Info("WSRunMessageSender() -> CLOSING.")
			send = false

		case data := <- ussn.DataOut:
			if err := ws.WriteJSON(data); err != nil {
				log.Error("error sending websocket message: ", err.Error())
				if start, count, limit = MaxWSError(start, count); limit {
					log.Error("CLOSING WS CONNECTION; MAX SEND ERRORS")
					ussn.WSSendErrorLimit <- struct{}{}
				}
			}
		}
	}

	if ussn.WSSendErrorLimit != nil {
		close(ussn.WSSendErrorLimit)
		ussn.WSSendErrorLimit = nil
	}// log.Info("WSRunMessageSender() -> STOPPED.")
}


type ProgressMessage struct {
	Source string `json:"source"`
	Label string `json:"label"`
	Percent int `json:"percent"`
}
func (ussn *UserSession) WSSendProgressStartMessage(source, label string) (err error) {
	return ussn.WSSendMessage("progress",  ProgressMessage{ source, label, 0 })
}
func (ussn *UserSession) WSSendProgressCompleteMessage(source, label string) (err error) {
	return ussn.WSSendMessage("progress",  ProgressMessage{ source, label, 100 })
}
func (ussn *UserSession) WSSendProgressMessage(source, label string, curr, end float32 ) (err error) {
	percent := int((float32(curr)/float32(end))*float32(100)) 
	// log.Info(fmt.Sprintf("WSSendProgressMessage( ) -> %s : %d : ", label, percent), source)
	return ussn.WSSendMessage("progress",  ProgressMessage{ source, label, percent })
}


type AggregateMessage struct {
	Source string `json:"source"`
	Aggregate `json:"agg"`
}
func (ussn *UserSession) WSSendClusterMessage(src string, cluster Aggregate) ( err error ) {
	// log.Info(fmt.Sprintf("WSSendClusterMessage( ) -> %s : %d : ", src, cluster.ID))
	return ussn.WSSendMessage("cluster",  AggregateMessage{ src, cluster })
}
func (ussn *UserSession) WSSendQSetMessage(src string, qset Aggregate) ( err error ) {
	// log.Info(fmt.Sprintf("WSSendQSetMessage( ) -> %s : %d : ", src, qset.ID))
	return ussn.WSSendMessage("qset",  AggregateMessage{ src, qset })
}


type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
func (ussn *UserSession) WSSendMessage(typ string, data interface{}) (err error) {
	// log.Info("WSSendMessage( ) -> typ : ", typ)
	if ( typ == "") {
		err = fmt.Errorf("error sending ws message: no message type")
		return
	}
	js, err := json.Marshal(&WSMessage{Type: typ, Data: data})
	if err != nil {
		err = fmt.Errorf("error marshaling websocket message: %s", err.Error())
		return
	}
	
	if *ussn, err = UserSessionsMapRead(ussn.SID.String()); err != nil {
		return
	}

	if ussn.Connected {
		if ussn.RWMChan == nil {
			ussn.RWMChan = &sync.RWMutex{}
		}
	
		ussn.RWMChan.Lock()
		if ussn.DataOut != nil {
			ussn.DataOut <- string(string(js))
		}
		ussn.RWMChan.Unlock()
	}

	// log.Info("WSSendMessage( ) -> DONE")
	return
}

