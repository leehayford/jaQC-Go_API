package utils

import (
	"encoding/json"
	"fmt"
	
	"github.com/gofiber/fiber/v2" // go get github.com/gofiber/fiber/v2
)

func ParseRequestBody(c *fiber.Ctx, obj interface{}) (err error) {
	/* USAGE */
	// thing := Thing{}
	// if err = pkg.ParseRequestBody(c, &thing ); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	// }
	if err = c.BodyParser(&obj); err != nil {
		err = fmt.Errorf("error parsing request body: %s", err.Error())
		return
	}

	return
}

func UnmarshalFormDataObject(c *fiber.Ctx, qry string, obj interface{}) (err error) {
	/* USAGE */
	// thing := Thing{}
	// if err = pkg.UnmarshalFormDataObject(c, "thing", &thing ); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	// }
	if err = json.Unmarshal([]byte(c.FormValue(qry)), &obj); err != nil {
		err = fmt.Errorf("error retrieving object from request form data: %s", err.Error())
	}
	return
}