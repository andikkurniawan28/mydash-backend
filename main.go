package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/mysaas"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()
	app.Use(cors.New())

	// ===== Auth =====
	app.Post("/api/login", loginProcess)
	app.Post("/api/register", registerProcess)
	app.Post("/api/change-password", changePasswordProcess)
	app.Get("/api/verify", verifyEmailHandler)

	// ===== ProfitLoss CRUD =====
	app.Post("/api/profitloss/stats", getProfitLossStats)
	app.Post("/api/profitloss/list", getAllProfitLoss)
	app.Get("/api/profitloss/:id", getProfitLossByID)
	app.Post("/api/profitloss", createProfitLoss)
	app.Put("/api/profitloss/:id", updateProfitLoss)
	app.Delete("/api/profitloss/:id", deleteProfitLoss)

	// ===== Ticket CRUD =====
	app.Post("/api/ticket/list", getAllTicket)
	app.Get("/api/ticket/:id", getTicketByID)
	app.Post("/api/ticket", createTicket)
	app.Put("/api/ticket/:id", updateTicket)
	app.Delete("/api/ticket/:id", deleteTicket)

	log.Fatal(app.Listen(":3001"))
}
