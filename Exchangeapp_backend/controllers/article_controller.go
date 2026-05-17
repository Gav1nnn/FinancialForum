package controllers

import (
	"encoding/json"
	"errors"
	"exchangeapp/global"
	"exchangeapp/models"
	"exchangeapp/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var cacheKey = "articles"

// articlePayload 是文章创建/更新接口的请求体。
type articlePayload struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Preview string `json:"preview" binding:"required"`
}

// CreateArticle 创建新文章，并清理相关文章缓存与索引缓存。
func CreateArticle(ctx *gin.Context) {
	username := ctx.GetString("username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload articlePayload

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article := models.Article{
		Title:          payload.Title,
		Content:        payload.Content,
		Preview:        payload.Preview,
		AuthorUsername: username,
	}

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := clearArticleCaches(article.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

// GetArticles 查询文章列表，优先读取 Redis 缓存。
func GetArticles(ctx *gin.Context) {
	cachedData, err := global.RedisDB.Get(cacheKey).Result()

	if err == redis.Nil {
		var articles []models.Article

		if err := global.Db.Find(&articles).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		articleJSON, err := json.Marshal(articles)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := global.RedisDB.Set(cacheKey, articleJSON, 10*time.Minute).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, articles)

	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		var articles []models.Article
		if err := json.Unmarshal([]byte(cachedData), &articles); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, articles)
	}
}

// GetMyArticles 查询当前登录用户发布的文章。
func GetMyArticles(ctx *gin.Context) {
	username := ctx.GetString("username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var articles []models.Article
	if err := global.Db.Where("author_username = ?", username).Order("updated_at desc").Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, articles)
}

// GetArticlesByID 根据 ID 查询单篇文章。
func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article

	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, article)
}

// UpdateArticle 更新当前用户自己的文章，并失效缓存。
func UpdateArticle(ctx *gin.Context) {
	username := ctx.GetString("username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	article, err := findOwnedArticle(ctx.Param("id"), username)
	if err != nil {
		handleArticleOwnershipError(ctx, err)
		return
	}

	var payload articlePayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.Title = payload.Title
	article.Content = payload.Content
	article.Preview = payload.Preview

	if err := global.Db.Save(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := clearArticleCaches(article.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, article)
}

// DeleteArticle 删除当前用户自己的文章，并失效缓存。
func DeleteArticle(ctx *gin.Context) {
	username := ctx.GetString("username")
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	article, err := findOwnedArticle(ctx.Param("id"), username)
	if err != nil {
		handleArticleOwnershipError(ctx, err)
		return
	}

	if err := global.Db.Delete(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := clearArticleCaches(article.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "article deleted"})
}

// findOwnedArticle 根据文章 ID + 用户名校验所有权。
func findOwnedArticle(id string, username string) (models.Article, error) {
	var article models.Article
	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		return article, err
	}

	if article.AuthorUsername != username {
		return article, errors.New("forbidden")
	}

	return article, nil
}

// handleArticleOwnershipError 统一处理文章归属相关错误。
func handleArticleOwnershipError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
	case err != nil && err.Error() == "forbidden":
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you can only manage your own articles"})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// clearArticleCaches 清理文章列表/点赞缓存，并触发 RAG 索引失效。
func clearArticleCaches(articleID uint) error {
	keys := []string{cacheKey}
	if articleID > 0 {
		keys = append(keys, likeCacheKey(articleID))
	}

	if err := global.RedisDB.Del(keys...).Err(); err != nil {
		return err
	}

	services.RAG.Invalidate()
	return nil
}

// likeCacheKey 生成文章点赞缓存键。
func likeCacheKey(articleID uint) string {
	return "article:" + strconv.FormatUint(uint64(articleID), 10) + ":likes"
}
