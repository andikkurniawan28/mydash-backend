package main

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// GET all tickets for user
func getAllTicket(c *fiber.Ctx) error {
	req := new(struct {
		UserID int `json:"userID"`
	})
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	rows, err := db.Query(`
		SELECT id, user_id, product_id, description, status, created_at, updated_at 
		FROM tickets 
		WHERE user_id = ? 
		ORDER BY created_at DESC
	`, req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProductID, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		tickets = append(tickets, t)
	}
	return c.JSON(tickets)
}

// GET ticket by ID
func getTicketByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var t Ticket
	err := db.QueryRow("SELECT id, user_id, product_id, description, status, created_at, updated_at FROM tickets WHERE id = ?", id).
		Scan(&t.ID, &t.UserID, &t.ProductID, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

// CREATE ticket (product_id = 1)
func createTicket(c *fiber.Ctx) error {
	t := new(Ticket)
	if err := c.BodyParser(t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if t.UserID == 0 || t.AppKey == "" || t.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user_id, app_key, and description required"})
	}

	// cek validasi app_key
	var valid int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id=? AND app_key=?", t.UserID, t.AppKey).Scan(&valid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if valid == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "invalid credentials"})
	}

	t.ProductID = 1
	t.Status = "open"

	res, err := db.Exec(`
		INSERT INTO tickets (user_id, product_id, description, status) 
		VALUES (?, ?, ?, ?)
	`, t.UserID, t.ProductID, t.Description, t.Status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	t.ID = int(id)

	return c.JSON(t)
}

// UPDATE ticket (description + status)
func updateTicket(c *fiber.Ctx) error {
	id := c.Params("id")
	t := new(Ticket)
	if err := c.BodyParser(t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	_, err := db.Exec("UPDATE tickets SET description=? WHERE id=?",
		t.Description, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	t.ID = atoi(id)
	return c.JSON(t)
}

// DELETE ticket
func deleteTicket(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := db.Exec("DELETE FROM tickets WHERE id=?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
