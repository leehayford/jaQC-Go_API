package main

import (
	"flag"
	// "fmt"

	"github.com/gofiber/fiber/v2" // go get github.com/gofiber/fiber/v2
	"github.com/gofiber/fiber/v2/log"

	"jaQC-Go-API/utils"
	"jaQC-Go-API/api"
)

func main() {
	
	/* CREATE GOFIBER APPLICATION */
	app := fiber.New(fiber.Config{
		DisablePreParseMultipartForm: true,
		StreamRequestBody:            false,
		BodyLimit:                    30 * 1024 * 1024, // 30MB upload limit
	})

	/* CONFIGURE LOGGING */
	defer (utils.ConfigureLogging(app, api.LOG_FILE)).Close()
	log.Info("\n\n\nJaQC SERVICE STARTING...\n\n\n")

	
	/* CHECK COMMAND LINE ARGUMENTS  ~$ go run . --clean */
	clean := flag.Bool("clean", false, "archive, drop and recreate database")
	flag.Parse() // log.Info("FLAG -> clean : ", *clean)


	/* FILE SYSTEM */
	if err := api.ConfigureFileSystem(*clean); err != nil {
		utils.LogFatal(err)
	}

	/* AUTH / SECURITY */
	api.ConfigureCORS(
		app,
		api.CORS_ORIGINS,
		api.CORS_HEADERS,
		api.CORS_METHODS,
		api.CORS_CREDETIALS,
	)

	api.ConfigureJWT(
		api.JWT_SECRET,
		api.JWT_REQ_AUTH_TYPE,
		api.JWT_REQ_KEY_COOKIE,
		api.JWT_REQ_KEY_QUERY,
		api.JWT_ACCESS_DURATION,
		api.JWT_REFRESH_DURATION,
	)

	/* EMAIL */
	api.ConfigureEmail(
		api.EMAIL_HOST,
		api.EMAIL_PORT,
		api.EMAIL_ADDRESS,
		api.EMAIL_PW,
	)
	
	/* API END POINTS */




	/* DEBUG ONLY ******************************************************************************/
	log.Info("**************************** main -> START DEBUG CODE\n\n\n")

	if *clean {
		log.Info("\n\n\nJaQC CLEANING...\n\n\n")
	}

	
	app.Get("/ping", func(c *fiber.Ctx) error {
        return c.SendString("pong")
    })

	
	// update device html, css, js
	app.Static("/api/jaqc/update_web", "./jaqcweb")


	log.Info("**************************** main -> END DEBUG CODE\n\n\n")
	/* END DEBUG ONLY *************************************************************************/

	// log.Fatal(app.Listen(api.API_HOST))
	log.Fatal(app.Listen("0.0.0.0:8013"))
}