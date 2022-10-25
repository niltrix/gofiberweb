package main

import (
	"context"
	// "context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

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
		err := c.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			fmt.Printf("Error : [%s]", err)
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(subs)
}

var tracer = otel.Tracer("fiber-server")

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// ODIC configuration
	// context = context.Background()
	// provider, err := oidc.NewProvider(context, "https://")

	// Print current process
	if fiber.IsChild() {
		fmt.Printf("[%d] Child\n", os.Getppid())
	} else {
		fmt.Printf("[%d] Master\n", os.Getppid())
	}

	app := fiber.New(fiber.Config{
		// Prefork: true,
	})
	app.Use(logger.New())
	app.Use(requestid.New(requestid.Config{
		Header: "x-request-id",
	}))

	app.Get("/", func(ctx *fiber.Ctx) error {
		_, span := tracer.Start(ctx.UserContext(), "getUser", oteltrace.WithAttributes(attribute.String("id", ctx.BaseURL())))
		defer span.End()
		return ctx.SendString("Hello")
	})

	app.Get("/subscription", getSubscription)
	app.Post("/subscription", createSubscription)

	err := app.Listen(":8080")
	if err != nil {
		fmt.Printf("Error : [%s]", err)
	}
}

func initTracer() *sdktrace.TracerProvider {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	//exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("my-service"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
