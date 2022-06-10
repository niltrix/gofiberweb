package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Subscription struct {
	Name    string `json:"name"`
	Product string `json:"product"`
}

func getSubscription(c *fiber.Ctx) error {
	subscription := Subscription{
		Name:    "Elon",
		Product: "Tesla",
	}
	return c.Status(fiber.StatusOK).JSON(subscription)
}

func createSubscription(c *fiber.Ctx) error {
	subs := new(Subscription)
	err := c.BodyParser(subs)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	return c.Status(fiber.StatusOK).JSON(subs)
}

func main() {
	// Print current process
	if fiber.IsChild() {
		fmt.Printf("[%d] Child\n", os.Getppid())
	} else {
		fmt.Printf("[%d] Master\n", os.Getppid())
	}

	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	app.Use(logger.New())
	app.Use(requestid.New(requestid.Config{
		Header: "x-request-id",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello")
	})

	app.Get("/subscription", getSubscription)
	app.Post("/subscription", createSubscription)

	app.Listen(":8080")
}
