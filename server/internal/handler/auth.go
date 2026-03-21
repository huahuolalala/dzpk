package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dz-poker/server/internal/model"
	"github.com/dz-poker/server/internal/service"
)

// AuthHandler handles authentication HTTP endpoints
type AuthHandler struct {
	authService *service.AuthService
	hub         *Hub // 用于推送消息
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SetHub sets the hub for the auth handler
func (h *AuthHandler) SetHub(hub *Hub) {
	h.hub = hub
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": err.Error()})
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"code": "user_exists", "message": "Username already taken"})
		case service.ErrInvalidAvatar:
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_avatar", "message": "Invalid avatar selection"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": err.Error()})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"code": "invalid_credentials", "message": "Invalid username or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserInfo handles getting user info (requires auth)
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Not authenticated"})
		return
	}

	resp, err := h.authService.GetUserInfo(userID.(string))
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": "user_not_found", "message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetGameHistory handles getting user game history (requires auth)
func (h *AuthHandler) GetGameHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Not authenticated"})
		return
	}

	resp, err := h.authService.GetUserGameHistory(userID.(string), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllUsers handles getting all users (admin only)
func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Not authenticated"})
		return
	}

	// Check admin
	if !h.authService.IsAdmin(userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"code": "forbidden", "message": "Admin access required"})
		return
	}

	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		return
	}

	resp := make([]model.AdminUserResp, len(users))
	for i, u := range users {
		resp[i] = model.AdminUserResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
			Chips:    u.Chips,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateUserChips handles updating user chips (admin only)
func (h *AuthHandler) UpdateUserChips(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Not authenticated"})
		return
	}

	// Check admin
	if !h.authService.IsAdmin(userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"code": "forbidden", "message": "Admin access required"})
		return
	}

	var req model.UpdateChipsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": err.Error()})
		return
	}

	if err := h.authService.UpdateUserChips(req.UserID, req.Chips); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
		return
	}

	// 通过 WebSocket 通知用户筹码更新
	if h.hub != nil {
		h.hub.NotifyUserChipsUpdate(req.UserID, req.Chips)
	}

	c.JSON(http.StatusOK, model.UpdateChipsResp{
		UserID: req.UserID,
		Chips:  req.Chips,
	})
}

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			tokenString = authHeader
		}

		result, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "invalid_token", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", result.User.ID)
		c.Set("user", result.User)
		c.Next()
	}
}
