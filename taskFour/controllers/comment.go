package controllers

import (
	"net/http"
	"strconv"
	"taskFour/config"
	"taskFour/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCommentInput 创建评论输入参数
type CreateCommentInput struct {
	Content string `json:"content" binding:"required,min=1" example:"这是一条评论"`
	PostID  uint   `json:"post_id" binding:"required" example:"1"`
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 对文章发表评论（需要认证）
// @Tags 评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body CreateCommentInput true "评论内容"
// @Success 201 {object} map[string]interface{} "评论创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "文章未找到"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /comments [post]
func CreateComment(c *gin.Context) {
	// 原有实现保持不变...
	userID := c.MustGet("user_id").(uint)

	var input CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查文章是否存在
	var post models.Post
	if err := config.GetDB().First(&post, input.PostID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	comment := models.Comment{
		Content: input.Content,
		UserID:  userID,
		PostID:  input.PostID,
	}

	if err := config.GetDB().Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// 重新加载以获取用户信息
	config.GetDB().Preload("User").First(&comment, comment.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment created successfully",
		"comment": comment,
	})
}

// GetPostComments 获取文章评论
// @Summary 获取文章评论列表
// @Description 获取指定文章的所有评论
// @Tags 评论
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "成功获取评论列表"
// @Failure 400 {object} map[string]interface{} "无效的文章ID"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts/{id}/comments [get]
func GetPostComments(c *gin.Context) {
	// 原有实现保持不变...
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var comments []models.Comment
	if err := config.GetDB().Preload("User").Where("post_id = ?", id).Order("created_at desc").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
