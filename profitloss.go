package main

import (
	"database/sql"
	"time"

	"strconv"

	"github.com/gofiber/fiber/v2"
)

// helper atoi
func atoi(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// =======================================
// CRUD ProfitLoss
// =======================================

// GET all
func getAllProfitLoss(c *fiber.Ctx) error {
	req := new(struct {
		UserID int `json:"userID"`
	})
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	rows, err := db.Query(`
        SELECT id, date, revenue, expense, profitloss 
        FROM profit_losses 
        WHERE user_id = ? 
        ORDER BY date DESC
    `, req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var result []ProfitLoss
	for rows.Next() {
		var pl ProfitLoss
		if err := rows.Scan(&pl.ID, &pl.Date, &pl.Revenue, &pl.Expense, &pl.ProfitLoss); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		result = append(result, pl)
	}
	return c.JSON(result)
}

// GET by ID
func getProfitLossByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var pl ProfitLoss
	err := db.QueryRow("SELECT id, date, revenue, expense, profitloss FROM profit_losses WHERE id = ?", id).
		Scan(&pl.ID, &pl.Date, &pl.Revenue, &pl.Expense, &pl.ProfitLoss)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pl)
}

// CREATE with app_key validation
func createProfitLoss(c *fiber.Ctx) error {
	pl := new(ProfitLoss)

	// Parse body
	if err := c.BodyParser(pl); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":  "invalid input",
			"detail": err.Error(),
		})
	}

	// Validasi user_id & app_key
	if pl.UserID == 0 || pl.AppKey == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":  "user_id and app_key are required",
			"detail": "Pastikan field 'user_id' dan 'app_key' ada dan tidak kosong",
		})
	}

	// Cek apakah user_id & app_key valid
	var valid int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ? AND app_key = ?", pl.UserID, pl.AppKey).Scan(&valid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if valid == 0 {
		return c.Status(403).JSON(fiber.Map{
			"error":  "invalid credentials",
			"detail": "user_id dan app_key tidak valid",
		})
	}

	// Validasi tanggal unik per user
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM profit_losses WHERE date = ? AND user_id = ?", pl.Date, pl.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if exists > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error":  "date already exists",
			"detail": "Sudah ada record dengan tanggal yang sama untuk user ini",
		})
	}

	// Hitung profit/loss
	pl.ProfitLoss = pl.Revenue - pl.Expense

	// Insert ke database
	res, err := db.Exec(
		"INSERT INTO profit_losses (user_id, date, revenue, expense, profitloss) VALUES (?, ?, ?, ?, ?)",
		pl.UserID, pl.Date, pl.Revenue, pl.Expense, pl.ProfitLoss,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	pl.ID = int(id)

	return c.JSON(pl)
}

// UPDATE
func updateProfitLoss(c *fiber.Ctx) error {
	id := c.Params("id")
	pl := new(ProfitLoss)
	if err := c.BodyParser(pl); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	pl.ProfitLoss = pl.Revenue - pl.Expense
	_, err := db.Exec("UPDATE profit_losses SET date=?, revenue=?, expense=?, profitloss=? WHERE id=?",
		pl.Date, pl.Revenue, pl.Expense, pl.ProfitLoss, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	pl.ID = atoi(id)
	return c.JSON(pl)
}

// DELETE
func deleteProfitLoss(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := db.Exec("DELETE FROM profit_losses WHERE id=?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// GET stats (sama seperti yang sebelumnya)
// func getProfitLossStats(c *fiber.Ctx) error {
// 	req := new(StatsRequest)
// 	if err := c.BodyParser(req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
// 	}

// 	rows, err := db.Query(`
//         SELECT id, date, revenue, expense, profitloss
//         FROM profit_losses
//         WHERE user_id = ?
//         ORDER BY date ASC
//     `, req.UserID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	defer rows.Close()

// 	var result []ProfitLoss
// 	dailyRevenue := make(map[string]float64)
// 	dailyExpense := make(map[string]float64)
// 	dailyProfitloss := make(map[string]float64)

// 	monthlyRevenue := make(map[string]float64)
// 	monthlyExpense := make(map[string]float64)
// 	monthlyProfitloss := make(map[string]float64)

// 	yearlyRevenue := make(map[string]float64)
// 	yearlyExpense := make(map[string]float64)
// 	yearlyProfitloss := make(map[string]float64)

// 	// insight tambahan
// 	var totalRevenue, totalExpense, totalProfit float64
// 	var maxRevenue, maxExpense, maxProfit float64
// 	var minRevenue, minExpense, minProfit float64
// 	minRevenue, minExpense, minProfit = 999999999, 999999999, 999999999

// 	// tanggal hari ini
// 	now := time.Now()
// 	currentYear, currentMonth, _ := now.Date()
// 	loc := now.Location()

// 	// range daily (bulan ini)
// 	firstDay := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, loc)
// 	lastDay := firstDay.AddDate(0, 1, -1)

// 	for rows.Next() {
// 		var pl ProfitLoss
// 		var dateStr string
// 		if err := rows.Scan(&pl.ID, &dateStr, &pl.Revenue, &pl.Expense, &pl.ProfitLoss); err != nil {
// 			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 		}

// 		// parsing tanggal
// 		t, err := time.Parse("2006-01-02", dateStr)
// 		if err != nil {
// 			return c.Status(500).JSON(fiber.Map{"error": "invalid date format"})
// 		}

// 		pl.Date = dateStr
// 		result = append(result, pl)

// 		// daily (hanya bulan ini)
// 		if !t.Before(firstDay) && !t.After(lastDay) {
// 			dayKey := t.Format("2006-01-02")
// 			dailyRevenue[dayKey] += pl.Revenue
// 			dailyExpense[dayKey] += pl.Expense
// 			dailyProfitloss[dayKey] += pl.ProfitLoss
// 		}

// 		// monthly (hanya tahun ini)
// 		if t.Year() == currentYear {
// 			monthKey := t.Format("January 2006")
// 			monthlyRevenue[monthKey] += pl.Revenue
// 			monthlyExpense[monthKey] += pl.Expense
// 			monthlyProfitloss[monthKey] += pl.ProfitLoss
// 		}

// 		// yearly (semua tahun)
// 		yearKey := t.Format("2006")
// 		yearlyRevenue[yearKey] += pl.Revenue
// 		yearlyExpense[yearKey] += pl.Expense
// 		yearlyProfitloss[yearKey] += pl.ProfitLoss

// 		// total
// 		totalRevenue += pl.Revenue
// 		totalExpense += pl.Expense
// 		totalProfit += pl.ProfitLoss

// 		// max & min
// 		if pl.Revenue > maxRevenue {
// 			maxRevenue = pl.Revenue
// 		}
// 		if pl.Revenue < minRevenue {
// 			minRevenue = pl.Revenue
// 		}
// 		if pl.Expense > maxExpense {
// 			maxExpense = pl.Expense
// 		}
// 		if pl.Expense < minExpense {
// 			minExpense = pl.Expense
// 		}
// 		if pl.ProfitLoss > maxProfit {
// 			maxProfit = pl.ProfitLoss
// 		}
// 		if pl.ProfitLoss < minProfit {
// 			minProfit = pl.ProfitLoss
// 		}
// 	}

// 	// isi daily kosong (bulan ini)
// 	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
// 		dayKey := d.Format("2006-01-02")
// 		if _, ok := dailyRevenue[dayKey]; !ok {
// 			dailyRevenue[dayKey] = 0
// 			dailyExpense[dayKey] = 0
// 			dailyProfitloss[dayKey] = 0
// 		}
// 	}

// 	// isi monthly kosong (Januari–Desember tahun ini)
// 	for m := 1; m <= 12; m++ {
// 		d := time.Date(currentYear, time.Month(m), 1, 0, 0, 0, 0, loc)
// 		monthKey := d.Format("January 2006")
// 		if _, ok := monthlyRevenue[monthKey]; !ok {
// 			monthlyRevenue[monthKey] = 0
// 			monthlyExpense[monthKey] = 0
// 			monthlyProfitloss[monthKey] = 0
// 		}
// 	}

// 	// konversi monthly ke slice agar urut + margin
// 	type MonthlyStat struct {
// 		Month        string  `json:"month"`
// 		Revenue      float64 `json:"revenue"`
// 		Expense      float64 `json:"expense"`
// 		ProfitLoss   float64 `json:"profitloss"`
// 		ProfitMargin float64 `json:"profitMargin"`
// 	}
// 	var monthlyStats []MonthlyStat
// 	for m := 1; m <= 12; m++ {
// 		d := time.Date(currentYear, time.Month(m), 1, 0, 0, 0, 0, loc)
// 		monthKey := d.Format("January 2006")
// 		margin := 0.0
// 		if monthlyRevenue[monthKey] > 0 {
// 			margin = (monthlyProfitloss[monthKey] / monthlyRevenue[monthKey]) * 100
// 		}
// 		monthlyStats = append(monthlyStats, MonthlyStat{
// 			Month:        monthKey,
// 			Revenue:      monthlyRevenue[monthKey],
// 			Expense:      monthlyExpense[monthKey],
// 			ProfitLoss:   monthlyProfitloss[monthKey],
// 			ProfitMargin: margin,
// 		})
// 	}

// 	// yearly profit margin
// 	yearlyProfitMargin := make(map[string]float64)
// 	for year, rev := range yearlyRevenue {
// 		if rev > 0 {
// 			yearlyProfitMargin[year] = (yearlyProfitloss[year] / rev) * 100
// 		} else {
// 			yearlyProfitMargin[year] = 0
// 		}
// 	}

// 	// insight tambahan: rata-rata harian (bulan ini saja)
// 	daysCount := len(dailyRevenue)
// 	avgRevenue := 0.0
// 	avgExpense := 0.0
// 	avgProfit := 0.0
// 	if daysCount > 0 {
// 		avgRevenue = totalRevenue / float64(daysCount)
// 		avgExpense = totalExpense / float64(daysCount)
// 		avgProfit = totalProfit / float64(daysCount)
// 	}

//		return c.JSON(fiber.Map{
//			"data":               result,
//			"dailyRevenue":       dailyRevenue,
//			"dailyExpense":       dailyExpense,
//			"dailyProfitloss":    dailyProfitloss,
//			"monthlyStats":       monthlyStats,
//			"yearlyRevenue":      yearlyRevenue,
//			"yearlyExpense":      yearlyExpense,
//			"yearlyProfitloss":   yearlyProfitloss,
//			"yearlyProfitMargin": yearlyProfitMargin,
//			"avgRevenue":         avgRevenue,
//			"avgExpense":         avgExpense,
//			"avgProfit":          avgProfit,
//			"maxRevenue":         maxRevenue,
//			"minRevenue":         minRevenue,
//			"maxExpense":         maxExpense,
//			"minExpense":         minExpense,
//			"maxProfit":          maxProfit,
//			"minProfit":          minProfit,
//		})
//	}
func getProfitLossStats(c *fiber.Ctx) error {
	req := new(StatsRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	rows, err := db.Query(`
        SELECT id, date, revenue, expense, profitloss 
        FROM profit_losses 
        WHERE user_id = ? 
        ORDER BY date ASC
    `, req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var result []ProfitLoss
	dailyRevenue := make(map[string]float64)
	dailyExpense := make(map[string]float64)
	dailyProfitloss := make(map[string]float64)

	monthlyRevenue := make(map[string]float64)
	monthlyExpense := make(map[string]float64)
	monthlyProfitloss := make(map[string]float64)

	yearlyRevenue := make(map[string]float64)
	yearlyExpense := make(map[string]float64)
	yearlyProfitloss := make(map[string]float64)

	// insight tambahan
	var totalRevenue, totalExpense, totalProfit float64
	var maxRevenue, maxExpense, maxProfit float64
	var minRevenue, minExpense, minProfit float64
	minRevenue, minExpense, minProfit = 999999, 999999, 999999

	// tanggal hari ini
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	loc := now.Location()

	// range daily (bulan ini)
	firstDay := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	for rows.Next() {
		var pl ProfitLoss
		var dateStr string
		if err := rows.Scan(&pl.ID, &dateStr, &pl.Revenue, &pl.Expense, &pl.ProfitLoss); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// parsing tanggal
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "invalid date format"})
		}

		pl.Date = dateStr
		result = append(result, pl)

		// daily (hanya bulan ini)
		if !t.Before(firstDay) && !t.After(lastDay) {
			dayKey := t.Format("2006-01-02")
			dailyRevenue[dayKey] += pl.Revenue
			dailyExpense[dayKey] += pl.Expense
			dailyProfitloss[dayKey] += pl.ProfitLoss
		}

		// monthly (hanya tahun ini)
		if t.Year() == currentYear {
			monthKey := t.Format("January 2006")
			monthlyRevenue[monthKey] += pl.Revenue
			monthlyExpense[monthKey] += pl.Expense
			monthlyProfitloss[monthKey] += pl.ProfitLoss
		}

		// yearly (semua tahun)
		yearKey := t.Format("2006")
		yearlyRevenue[yearKey] += pl.Revenue
		yearlyExpense[yearKey] += pl.Expense
		yearlyProfitloss[yearKey] += pl.ProfitLoss

		// total
		totalRevenue += pl.Revenue
		totalExpense += pl.Expense
		totalProfit += pl.ProfitLoss

		// max & min
		if pl.Revenue > maxRevenue {
			maxRevenue = pl.Revenue
		}
		if pl.Revenue < minRevenue {
			minRevenue = pl.Revenue
		}
		if pl.Expense > maxExpense {
			maxExpense = pl.Expense
		}
		if pl.Expense < minExpense {
			minExpense = pl.Expense
		}
		if pl.ProfitLoss > maxProfit {
			maxProfit = pl.ProfitLoss
		}
		if pl.ProfitLoss < minProfit {
			minProfit = pl.ProfitLoss
		}
	}

	// isi daily kosong (bulan ini)
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		if _, ok := dailyRevenue[dayKey]; !ok {
			dailyRevenue[dayKey] = 0
			dailyExpense[dayKey] = 0
			dailyProfitloss[dayKey] = 0
		}
	}

	// isi monthly kosong (Januari–Desember tahun ini)
	for m := 1; m <= 12; m++ {
		d := time.Date(currentYear, time.Month(m), 1, 0, 0, 0, 0, loc)
		monthKey := d.Format("January 2006")
		if _, ok := monthlyRevenue[monthKey]; !ok {
			monthlyRevenue[monthKey] = 0
			monthlyExpense[monthKey] = 0
			monthlyProfitloss[monthKey] = 0
		}
	}

	// konversi monthly ke slice agar urut + margin
	type MonthlyStat struct {
		Month        string  `json:"month"`
		Revenue      float64 `json:"revenue"`
		Expense      float64 `json:"expense"`
		ProfitLoss   float64 `json:"profitloss"`
		ProfitMargin float64 `json:"profitMargin"`
	}
	var monthlyStats []MonthlyStat
	for m := 1; m <= 12; m++ {
		d := time.Date(currentYear, time.Month(m), 1, 0, 0, 0, 0, loc)
		monthKey := d.Format("January 2006")
		margin := 0.0
		if monthlyRevenue[monthKey] > 0 {
			margin = (monthlyProfitloss[monthKey] / monthlyRevenue[monthKey]) * 100
		}
		monthlyStats = append(monthlyStats, MonthlyStat{
			Month:        monthKey,
			Revenue:      monthlyRevenue[monthKey],
			Expense:      monthlyExpense[monthKey],
			ProfitLoss:   monthlyProfitloss[monthKey],
			ProfitMargin: margin,
		})
	}

	// yearly profit margin
	yearlyProfitMargin := make(map[string]float64)
	for year, rev := range yearlyRevenue {
		if rev > 0 {
			yearlyProfitMargin[year] = (yearlyProfitloss[year] / rev) * 100
		} else {
			yearlyProfitMargin[year] = 0
		}
	}

	// insight tambahan: rata-rata harian (bulan ini saja, hanya hari dengan transaksi)
	sumRevenue := 0.0
	sumExpense := 0.0
	sumProfit := 0.0
	countActiveDays := 0

	for day, rev := range dailyRevenue {
		exp := dailyExpense[day]
		prof := dailyProfitloss[day]

		// hanya hitung kalau ada transaksi
		if rev != 0 || exp != 0 || prof != 0 {
			sumRevenue += rev
			sumExpense += exp
			sumProfit += prof
			countActiveDays++
		}
	}

	avgRevenue := 0.0
	avgExpense := 0.0
	avgProfit := 0.0
	if countActiveDays > 0 {
		avgRevenue = sumRevenue / float64(countActiveDays)
		avgExpense = sumExpense / float64(countActiveDays)
		avgProfit = sumProfit / float64(countActiveDays)
	}

	// kalau tidak ada data, set min jadi 999999
	if len(result) == 0 {
		minRevenue = 999999
		minExpense = 0
		minProfit = 999999
	}

	return c.JSON(fiber.Map{
		"data":               result,
		"dailyRevenue":       dailyRevenue,
		"dailyExpense":       dailyExpense,
		"dailyProfitloss":    dailyProfitloss,
		"monthlyStats":       monthlyStats,
		"yearlyRevenue":      yearlyRevenue,
		"yearlyExpense":      yearlyExpense,
		"yearlyProfitloss":   yearlyProfitloss,
		"yearlyProfitMargin": yearlyProfitMargin,
		"avgRevenue":         avgRevenue, // pendapatan rata-rata harian (bulan ini)
		"avgExpense":         avgExpense, // beban rata-rata harian (bulan ini)
		"avgProfit":          avgProfit,  // laba rata-rata harian (bulan ini)
		"maxRevenue":         maxRevenue,
		"minRevenue":         minRevenue,
		"maxExpense":         maxExpense,
		"minExpense":         minExpense,
		"maxProfit":          maxProfit,
		"minProfit":          minProfit,
	})
}
