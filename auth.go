package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/smtp"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// =======================================
// LOGIN PROCESS
// =======================================
func loginProcess(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	var (
		user             User
		hashedPassword   string
		isActive         bool
		accessToProduct1 bool
	)

	// Ambil user berdasarkan email (tambahkan app_key)
	err := db.QueryRow(`
		SELECT id, role_id, name, email, password, is_active, access_to_product_1, organization, whatsapp, app_key
		FROM users 
		WHERE email = ?
		LIMIT 1
	`, req.Email).Scan(
		&user.ID, &user.RoleID, &user.Name, &user.Email,
		&hashedPassword, &isActive, &accessToProduct1,
		&user.Organization, &user.Whatsapp, &user.AppKey,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Cek apakah user aktif
	if !isActive {
		return c.Status(403).JSON(fiber.Map{"error": "user not active"})
	}

	// Cek apakah user punya akses ke produk 1
	if !accessToProduct1 {
		return c.Status(403).JSON(fiber.Map{"error": "user does not have access to this system"})
	}

	// Cek password bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// Login sukses
	return c.JSON(fiber.Map{
		"message": "login success",
		"user":    user,
	})
}

// =======================================
// REGISTER PROCESS
// =======================================
func registerProcess(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		Whatsapp     string `json:"whatsapp"`
		Password     string `json:"password"`
	}

	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// cek email sudah ada atau belum
	var existing int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&existing)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if existing > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "email already registered"})
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}

	// generate app_key
	appKey := generateRandomString(8)

	// insert user baru (email_verified_at = NULL)
	result, err := db.Exec(`
		INSERT INTO users (role_id, name, email, password, is_active, organization, whatsapp, app_key, email_verified_at, access_to_product_1)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, 1)
	`,
		3, req.Name, req.Email, string(hashedPassword),
		0, req.Organization, req.Whatsapp, appKey,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	lastID, _ := result.LastInsertId()

	// bikin token verifikasi
	token := generateVerificationToken(req.Email, appKey)

	// bikin link verifikasi
	verificationLink := fmt.Sprintf("http://147.139.177.186:3378/api/verify?email=%s&token=%s", req.Email, token)

	// kirim email (async)
	go sendVerificationEmail(req.Email, verificationLink)

	return c.JSON(fiber.Map{
		"message": "register success, please check your email to verify account",
		"user": fiber.Map{
			"id":           lastID,
			"role_id":      3,
			"name":         req.Name,
			"email":        req.Email,
			"organization": req.Organization,
			"whatsapp":     req.Whatsapp,
			"app_key":      appKey,
			"is_active":    1,
		},
	})
}

// =======================================
// CHANGE PASSWORD PROCESS
// =======================================
func changePasswordProcess(c *fiber.Ctx) error {
	type ChangePasswordRequest struct {
		UserID          int    `json:"user_id"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	req := new(ChangePasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// Ambil password lama dari DB
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE id = ?", req.UserID).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Cek current password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.CurrentPassword)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "current password is incorrect"})
	}

	// Hash password baru
	newHashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash new password"})
	}

	// Update ke DB
	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHashed), req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "password updated successfully"})
}

// =======================================
// Helper: Generate Random String
// =======================================
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	for i := range bytes {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		bytes[i] = letters[num.Int64()]
	}
	return string(bytes)
}

// =======================================
// HELPER: Generate Verification Token
// =======================================
func generateVerificationToken(email, appKey string) string {
	data := fmt.Sprintf("%s:%s:%d", email, appKey, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// =======================================
// HELPER: Send Verification Email
// =======================================
func sendVerificationEmail(to string, link string) error {
	from := "optimateknologiindustri@gmail.com"
	password := "gzew ksdw kdef bcex" // ⚠️ ini App Password, bukan password Gmail biasa

	// Gmail SMTP server
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Pesan email
	subject := "Subject: Verify your email\n"
	body := fmt.Sprintf("Halo,\n\nSilakan klik link berikut untuk verifikasi email Anda:\n%s\n\nTerima kasih.", link)
	msg := []byte(subject + "\n" + body)

	// Kirim
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}

// =======================================
// VERIFY EMAIL HANDLER
// =======================================
func verifyEmailHandler(c *fiber.Ctx) error {
	email := c.Query("email")
	token := c.Query("token")

	if email == "" || token == "" {
		return c.Status(400).SendString("Invalid verification link")
	}

	// (opsional) validasi token

	// update email_verified_at + aktifkan user
	_, err := db.Exec(
		"UPDATE users SET email_verified_at = ?, is_active = 1 WHERE email = ?",
		time.Now(), email,
	)
	if err != nil {
		return c.Status(500).SendString("Failed to verify email")
	}

	// redirect ke halaman tujuan
	return c.Redirect("http://147.139.177.186:4498/", fiber.StatusSeeOther)
}
