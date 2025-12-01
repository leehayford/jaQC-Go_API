package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt" 	// go get github.com/golang-jwt/jwt
)

/* CREATE A JWTConfiguration  */
type JWTConfiguration struct {
	Secret    string        // ~$ node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
	AccDur    time.Duration // time.Duration(time.Minute * 15)
	RefDur    time.Duration // time.Duration(time.Hour * 24)
	AuthType  string        // "Bearer "
	CookieKey string        // "token"
	QueryKey  string        // "access_token"
}

/* CREATES A JWT REFRESH TOKEN; USED ON LOGIN ONLY */
func (cfg *JWTConfiguration) CreateRefreshToken(uid int64) (tok string, err error) {
	// log.Info("(*JWTConfiguration) CreateRefreshToken( )")

	now := time.Now().Unix()
	exp := now + int64(cfg.RefDur.Seconds())

	/* CREATE JWT CLAIMS FOR A GIVEN USER */
	claims := jwt.MapClaims{
		"sub": uid, // SUBJECT
		"exp": exp,
		"iat": now, // ISSUED AT
		"nbf": now, // NOT VALID BEFORE
	}
	// log.Info("(*JWTConfiguration) CreateRefreshToken( ) -> claims : ", claims)

	tokBytes := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// log.Info("(*JWTConfiguration) CreateRefreshToken( ) -> tokBytes : ", tokBytes)

	if tok, err = tokBytes.SignedString([]byte(cfg.Secret)); err != nil {
		err = fmt.Errorf("failed to sign refresh token: %s", err.Error())
	}
	// log.Info("(*JWTConfiguration) CreateRefreshToken( ) -> tok : ", tok)
	return
}

/* CREATES A JWT ACCESS TOKEN; USED FOR LOGIN AND REFRESH */
func (cfg *JWTConfiguration) CreateAccessToken(uid int64, role string) (tok string, err error) {
	// log.Info("(*JWTConfiguration) CreateAccessToken( )")

	now := time.Now().Unix()
	exp := now + int64(cfg.AccDur.Seconds())

	// log.Info("CreateAccessToken -> now : ", now)
	// log.Info("CreateAccessToken -> exp : ", exp)
	// log.Info("CreateAccessToken -> diff : ", (exp - now))

	/* CREATE JWT CLAIMS FOR A GIVEN USER */
	claims := jwt.MapClaims{
		"sub": uid,  // SUBJECT
		"rol": role, // ROLE
		"exp": exp,
		"iat": now, // ISSUED AT
		"nbf": now, // NOT VALID BEFORE
	}
	tokBytes := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if tok, err = tokBytes.SignedString([]byte(cfg.Secret)); err != nil {
		err = fmt.Errorf("failed to sign access token: %s", err.Error())
	}
	return
}

/* RETURNS ALL TOKEN CLAIMS */
func (cfg *JWTConfiguration) ClaimsFromTokenString(token string) (claims jwt.MapClaims, err error) {

	/* PARSE TOKEN STRING */
	tokBytes, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, jwt_err := jwtToken.Method.(*jwt.SigningMethodHMAC); !jwt_err {
			return nil, fmt.Errorf("unexpected signing method: %s", jwtToken.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return
	}

	claims, ok := tokBytes.Claims.(jwt.MapClaims)
	if !ok || !tokBytes.Valid {
		err = fmt.Errorf("invalid token claim")
		return
	}
	return
}

func (cfg *JWTConfiguration) Authenticate(c *fiber.Ctx) (err error) {
	// log.Info("(*JWTConfiguration) Authenticate")

	tok := ""

	/* CHECK REQUEST HEADER */
	req_auth := c.Get("Authorization")
	if strings.HasPrefix(req_auth, cfg.AuthType) {
		tok = strings.TrimPrefix(req_auth, cfg.AuthType)

		/* CHECK REQUEST COOKIES */
	} else if c.Cookies(cfg.CookieKey) != "" {
		tok = c.Cookies(cfg.CookieKey)

		/* CHECK URL */
	} else if c.Query(cfg.QueryKey) != "" {
		tok = c.Query(cfg.QueryKey)
	}

	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication failed; please log in")
	}

	/* GET TOKEN CLAIMS */
	claims, err := cfg.ClaimsFromTokenString(tok)
	if err != nil {
		txt := fmt.Sprintf("authentication failed: %s", err.Error())
		return c.Status(fiber.StatusUnauthorized).SendString(txt)
	}

	/* CHECK IF TOKEN HAS EXPIRED */
	exp := int64(claims["exp"].(float64))
	// log.Info("JWT.exp : ", exp)
	now := time.Now().UTC().Unix()
	// log.Info("JWT.now : ", now)
	if exp < now {
		return c.Status(fiber.StatusUnauthorized).SendString("token is expired")
	}

	/* PASS USER AND ROLE DATA ALONG TO THE NEXT HANDLER */
	c.Locals("sub", claims["sub"])
	c.Locals("role", claims["rol"])

	return c.Next()
}
