package utils

import (
	"bytes"
    "crypto/tls"
    "errors"
	"fmt"
	"net"
	"net/smtp"
	"text/template"
)

type EmailConfiguration struct {
	Host string
	Port string
	From string
	Password string
}

const MIME_TEXT_HTML = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

func (eml EmailConfiguration) Auth() smtp.Auth {
	return smtp.PlainAuth("", eml.From, eml.Password, eml.Host )
}

func (eml EmailConfiguration) AddrString() string {
	return fmt.Sprintf("%s:%s", eml.Host, eml.Port)
} 

func (eml EmailConfiguration) SendPlainText(to []string, msg string) error {
	return smtp.SendMail(
		eml.AddrString(),
		eml.Auth(),
		eml.From,
		to,
		[]byte(msg),
	)
}
type loginAuth struct {
    username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
    if more {
        switch string(fromServer) {
        case "Username:":
            return []byte(a.username), nil
        case "Password:":
            return []byte(a.password), nil
        default:
            return nil, errors.New("unknown from server")
        }
    }
    return nil, nil
}
func LoginAuth(username, password string) smtp.Auth {
    return &loginAuth{username, password}
}

func (eml EmailConfiguration) SendHTML(to []string, tplt, sub string, tpltVars interface{}) error {
	
	conn, err := net.Dial("tcp", eml.AddrString())
    if err != nil {
        return LogErr(err)
    }

    c, err := smtp.NewClient(conn, eml.Host)
    if err != nil {
        return LogErr(err)
    }

    tlsconfig := &tls.Config{
        ServerName: eml.Host,
    }

    if err = c.StartTLS(tlsconfig); err != nil {
        return LogErr(err)
    }

	auth := LoginAuth( eml.From, eml.Password)
	
	if err = c.Auth(auth); err != nil {
        return LogErr(err)
	} 

	t, _ := template.ParseFiles(tplt)

	body := bytes.Buffer{}

	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", sub, MIME_TEXT_HTML)))
	t.Execute(&body, tpltVars)

	return smtp.SendMail(
		eml.AddrString(),
		auth, // eml.Auth(),
		eml.From,
		to,
		body.Bytes(),
	)
}


/* 
EXAMPLE: SendHTML()

TEMPLATE -> "templates/confirm_pw_reset.html"

<!-- confirm_pw_reset.html -->
<!DOCTYPE html>
<html>
	<body>
		<h3>Password Reset:</h3>
		<span>Please click on the link below to reset your password.</span>
		<br/>
		<br/>
		<a href={{.Link}}>RESET MY PASSWORD</a>
		<br/>
	</body>
</html>

type HTMLResetPW struct {
	Link string
}

func HTMLEmailTest() (err error) {
	tplt := HTMLResetPW {
		Link: "\"https://leehayford.com\"",
	}
	return api.EML.SendHTML(
		[]string{ "leehayford@gmail.com" },
		"templates/confirm_pw_reset.html",
		"Reset Password",
		tplt,
	)
}
*/