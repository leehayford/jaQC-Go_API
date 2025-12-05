package main

import (
	"flag"
	// "fmt"

	/* /api/jaqc/update_web imports **********/
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
    "path/filepath"
	/* END /api/jaqc/update_web imports ******/


	"github.com/gofiber/fiber/v2" // go get github.com/gofiber/fiber/v2
	"github.com/gofiber/fiber/v2/log"

	"jaQC-Go-API/utils"
	"jaQC-Go-API/api"
)


/* /api/jaqc/update_web stuff **************/
type fileDesc struct {
    Name string // e.g., "index.html"
    Path string // e.g., "/storage/index.html"
}

func hashFileHex(fullPath string) (hexStr string, size int64, err error) {
    f, err := os.Open(fullPath)
    if err != nil {
        return "", 0, err
    }
    defer f.Close()

    h := sha256.New()
    n, err := io.Copy(h, f)
    if err != nil {
        return "", 0, err
    }
    sum := h.Sum(nil)
    return hex.EncodeToString(sum), n, nil
}

func RegisterManifestRoute(app *fiber.App) {
    // Configure your base URL and server-side storage root
    const baseURL = "http://192.168.1.165:8013/api/jaqc/update_web"
    const webroot = "./jaqcweb" // where the source files are on your server

    // List the assets you want to publish
    assets := []fileDesc{
        {Name: "index.html", Path: "/storage/index.html"},
        {Name: "app.css",    Path: "/storage/app.css"},
        {Name: "app.js",     Path: "/storage/app.js"},
        {Name: "favicon.svg",Path: "/storage/favicon.svg"},
    }

    app.Get("/api/jaqc/manifest", func(c *fiber.Ctx) error {
        type item struct {
            URL    string `json:"url"`
            Path   string `json:"path"`
            SHA256 string `json:"sha256"`
            Size   int64  `json:"size,omitempty"`
        }
        out := struct {
            Files []item `json:"files"`
        }{
            Files: make([]item, 0, len(assets)),
        }

        for _, a := range assets {
            fullPath := filepath.Join(webroot, a.Name)
            hexStr, size, err := hashFileHex(fullPath)
            if err != nil {
                // Fail fast so the device knows something is wrong
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "error":   "hashing failed",
                    "file":    a.Name,
                    "details": err.Error(),
                })
            }

            out.Files = append(out.Files, item{
                URL:    baseURL + "/" + a.Name,
                Path:   a.Path,
                SHA256: hexStr, // 64 hex chars
                Size:   size,   // optional
            })
        }

        // You may also add manifest-level metadata (version, timestamp)
        return c.JSON(out)
    })
}
/* END /api/jaqc/update_web stuff **********/

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


    // Static assets (place files in ./jaqcweb/)
	app.Static("/api/jaqc/update_web", "./jaqcweb") // serves index.html, app.css, app.js, favicon.svg
	
    // Manifest
	/* OLD /api/jaqc/manifest **********************/
    // app.Get("/api/jaqc/manifest", func(c *fiber.Ctx) error {
    //     return c.JSON(fiber.Map{
    //         "files": []fiber.Map{
    //             {"url": "http://192.168.1.165:8013/api/jaqc/update_web/index.html", "path": "/storage/index.html", "sha256": ""},
    //             {"url": "http://192.168.1.165:8013/api/jaqc/update_web/app.css",    "path": "/storage/app.css",    "sha256": ""},
    //             {"url": "http://192.168.1.165:8013/api/jaqc/update_web/app.js",     "path": "/storage/app.js",     "sha256": ""},
    //             {"url": "http://192.168.1.165:8013/api/jaqc/update_web/favicon.svg","path": "/storage/favicon.svg","sha256": ""},
    //         },
    //     })
    // })
	/* END OLD /api/jaqc/manifest ******************/

	// Configure our base URL and server-side storage root
    const baseURL = "http://192.168.1.165:8013/api/jaqc/update_web" // TODO: Construct this from other constants
    const webroot = "./jaqcweb" // where the source files are on your server

    // List the assets we want to publish
    assets := []fileDesc{
        {Name: "index.html", Path: "/storage/index.html"},
        {Name: "app.css",    Path: "/storage/app.css"},
        {Name: "app.js",     Path: "/storage/app.js"},
        {Name: "favicon.svg",Path: "/storage/favicon.svg"},
    }
	
	app.Get("/api/jaqc/manifest", func(c *fiber.Ctx) error {
        type item struct {
            URL    string `json:"url"`
            Path   string `json:"path"`
            SHA256 string `json:"sha256"`
            Size   int64  `json:"size,omitempty"`
        }

        out := struct { 
			Files []item `json:"files"`
        }{ 	// TODO: ASK about this syntax
			Files: make([]item, 0, len(assets)),
        }

        for _, a := range assets {
            fullPath := filepath.Join(webroot, a.Name)
            hexStr, size, err := hashFileHex(fullPath)
            if err != nil {
                // Fail fast so the device knows something is wrong
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "error":   "hashing failed",
                    "file":    a.Name,
                    "details": err.Error(),
                })
            }
			log.Info(hexStr)

            out.Files = append(out.Files, item{
                URL:    baseURL + "/" + a.Name,
                Path:   a.Path,
                SHA256: hexStr, // 64 hex chars
                Size:   size,   // optional
            })
        }

        // TODO: possibly add manifest-level metadata (version, timestamp)
        return c.JSON(out)
    })

	log.Info("**************************** main -> END DEBUG CODE\n\n\n")
	/* END DEBUG ONLY *************************************************************************/

	// log.Fatal(app.Listen(api.API_HOST))
	log.Fatal(app.Listen("0.0.0.0:8013"))
}