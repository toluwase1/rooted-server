package admin

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
}

func NewAuthHandler(db *pgxpool.Pool, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

type AdminUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (h *AuthHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/admin/auth/login", h.Login)
	app.Get("/admin/auth/me", h.AdminJWTAuth(), h.GetMe)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	var user AdminUser
	var passwordHash string
	err := h.db.QueryRow(c.Context(), `
		SELECT id, email, name, role, password_hash
		FROM admin_users WHERE email = $1
	`, body.Email).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &passwordHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// Update last login
	h.db.Exec(c.Context(), "UPDATE admin_users SET last_login_at = NOW() WHERE id = $1", user.ID)

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		log.Printf("ERROR generating JWT: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{"token": tokenString, "user": user})
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	user := c.Locals("admin_user").(AdminUser)
	return c.JSON(fiber.Map{"user": user})
}

// AdminJWTAuth middleware validates admin JWT tokens.
func (h *AuthHandler) AdminJWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{"error": "missing authorization"})
		}

		tokenString := auth[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token claims"})
		}

		user := AdminUser{
			ID:    claims["sub"].(string),
			Email: claims["email"].(string),
			Name:  claims["name"].(string),
			Role:  claims["role"].(string),
		}

		c.Locals("admin_user", user)
		return c.Next()
	}
}
