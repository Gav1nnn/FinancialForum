package controllers

import (
	"exchangeapp/global"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// LikeArticle 对指定文章的点赞计数做原子自增。
func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := likeKeyFromParam(articleID)
	if err := global.RedisDB.Incr(likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully liked the article"})

}

// GetArticleLikes 查询文章当前点赞数（无记录时返回 0）。
func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := likeKeyFromParam(articleID)
	likes, err := global.RedisDB.Get(likeKey).Result()

	if err == redis.Nil {
		likes = "0"
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}

// likeKeyFromParam 构造点赞缓存键；若 ID 非数字则回退为原始字符串。
func likeKeyFromParam(articleID string) string {
	if parsedID, err := strconv.ParseUint(articleID, 10, 64); err == nil {
		return likeCacheKey(uint(parsedID))
	}
	return "article:" + articleID + ":likes"
}
