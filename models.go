package main

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID               int    `json:"id"`
	RoleID           int    `json:"role_id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Organization     string `json:"organization"`
	Whatsapp         string `json:"whatsapp"`
	AccessToProduct1 bool   `json:"access_to_product_1"`
	AppKey           string `json:"app_key"`
}

type ProfitLoss struct {
	ID         int     `json:"id"`
	UserID     int     `json:"user_id"`
	Date       string  `json:"date"`
	Revenue    float64 `json:"revenue"`
	Expense    float64 `json:"expense"`
	ProfitLoss float64 `json:"profitloss"`
	AppKey     string  `json:"app_key"`
}

type StatsRequest struct {
	UserID int `json:"user_id"`
}

type UserRequest struct {
	UserID int64 `json:"user_id"`
}

type Ticket struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	ProductID   int    `json:"product_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	AppKey      string `json:"app_key"` // dari front-end
}
