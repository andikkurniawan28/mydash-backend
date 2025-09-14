package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Ambil stage
	stage := getEnv("APP_STAGE", "dev")

	// Tentukan koneksi database sesuai stage
	var dsn string
	switch stage {
	case "dev":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
			getEnv("LOCAL_DB_USER", "root"),
			getEnv("LOCAL_DB_PASSWORD", ""),
			getEnv("LOCAL_DB_HOST", "127.0.0.1"),
			getEnv("LOCAL_DB_NAME", "mysaas"),
		)
	case "vpn":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
			getEnv("VPN_DB_USER", "root"),
			getEnv("VPN_DB_PASSWORD", ""),
			getEnv("VPN_DB_HOST", "127.0.0.1"),
			getEnv("VPN_DB_NAME", "mysaas"),
		)
	case "vps":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
			getEnv("VPS_DB_USER", "andik"),
			getEnv("VPS_DB_PASSWORD", ""),
			getEnv("VPS_DB_HOST", "127.0.0.1"),
			getEnv("VPS_DB_NAME", "mysaas"),
		)
	default:
		log.Fatal("Unknown APP_STAGE: ", stage)
	}

	log.Println("APP_STAGE:", stage)
	log.Println("Using DSN:", dsn)

	// Connect DB
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Fiber setup
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

	// Jalankan server di port dari .env
	appPort := getEnv("APP_PORT", "3001")
	log.Fatal(app.Listen(":" + appPort))
}

// helper ambil env dengan default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
