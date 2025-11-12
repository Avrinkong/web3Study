package controllers

import (
	"net/http"
	"strconv"
	"taskFour/config"
	"taskFour/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePostInput 创建文章输入参数
type CreatePostInput struct {
	Title   string `json:"title" binding:"required,min=1,max=200" example:"我的第一篇文章"`
	Content string `json:"content" binding:"required,min=1" example:"这是文章的内容..."`
}

// UpdatePostInput 更新文章输入参数
type UpdatePostInput struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=200" example:"更新后的文章标题"`
	Content string `json:"content" binding:"omitempty,min=1" example:"更新后的文章内容..."`
}

// PostsResponse 文章列表响应
type PostsResponse struct {
	Posts []models.Post `json:"posts"`
	Page  int           `json:"page" example:"1"`
	Limit int           `json:"limit" example:"10"`
}

// CreatePost 创建文章
// @Summary 创建文章
// @Description 创建新的博客文章（需要认证）
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body CreatePostInput true "文章内容"
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	// 原有实现保持不变...
	userID := c.MustGet("user_id").(uint)

	var input CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post := models.Post{
		Title:   input.Title,
		Content: input.Content,
		UserID:  userID,
	}

	if err := config.GetDB().Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// 重新加载以获取用户信息
	config.GetDB().Preload("User").First(&post, post.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post created successfully",
		"post":    post,
	})
}

// GetPosts 获取文章列表
// @Summary 获取文章列表
// @Description 获取分页的文章列表
// @Tags 文章
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} PostsResponse "成功获取文章列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts [get]
func GetPosts(c *gin.Context) {
	// 原有实现保持不变...
	var posts []models.Post

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	if err := config.GetDB().Preload("User").Offset(offset).Limit(limit).Order("created_at desc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
		"limit": limit,
	})
}

// GetPost 获取单篇文章
// @Summary 获取单篇文章
// @Description 根据ID获取单篇文章的详细信息
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "成功获取文章"
// @Failure 400 {object} map[string]interface{} "无效的文章ID"
// @Failure 404 {object} map[string]interface{} "文章未找到"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts/{id} [get]
func GetPost(c *gin.Context) {
	// 原有实现保持不变...
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := config.GetDB().Preload("User").Preload("Comments.User").First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}

// UpdatePost 更新文章
// @Summary 更新文章
// @Description 更新指定文章的内容（仅文章作者可操作）
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Param input body UpdatePostInput true "更新内容"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "文章未找到"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	// 原有实现保持不变...
	userID := c.MustGet("user_id").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := config.GetDB().First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	// 检查权限
	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own posts"})
		return
	}

	var input UpdatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}

	if err := config.GetDB().Model(&post).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	config.GetDB().Preload("User").First(&post, post.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    post,
	})
}

// DeletePost 删除文章
// @Summary 删除文章
// @Description 删除指定文章（仅文章作者可操作）
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "无效的文章ID"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Failure 404 {object} map[string]interface{} "文章未找到"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	// 原有实现保持不变...
	userID := c.MustGet("user_id").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := config.GetDB().First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	// 检查权限
	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own posts"})
		return
	}

	if err := config.GetDB().Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
