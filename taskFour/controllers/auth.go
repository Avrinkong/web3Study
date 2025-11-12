package controllers

import (
	"net/http"
	"taskFour/config"
	"taskFour/middleware"
	"taskFour/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterInput 注册输入参数
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=100" example:"testuser"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`
}

// LoginInput 登录输入参数
type LoginInput struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Message string `json:"message" example:"Login successful"`
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    struct {
		ID       uint   `json:"id" example:"1"`
		Username string `json:"username" example:"testuser"`
		Email    string `json:"email" example:"test@example.com"`
	} `json:"user"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param input body RegisterInput true "注册信息"
// @Success 201 {object} map[string]interface{} "注册成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	// 原有实现保持不变...
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已存在
	var existingUser models.User
	if err := config.GetDB().Where("username = ? OR email = ?", input.Username, input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	user := models.User{
		Username: input.Username,
		Password: input.Password,
		Email:    input.Email,
	}

	if err := config.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param input body LoginInput true "登录信息"
// @Success 200 {object} LoginResponse "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "认证失败"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	// 原有实现保持不变...
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Message: "Login successful",
		Token:   token,
	}
	response.User.ID = user.ID
	response.User.Username = user.Username
	response.User.Email = user.Email

	c.JSON(http.StatusOK, response)
}
