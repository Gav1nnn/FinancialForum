package controllers

import (
	"exchangeapp/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// assistantChatRequest 是助手问答接口的请求体。
type assistantChatRequest struct {
	Question string                      `json:"question" binding:"required"`
	History  []services.AssistantMessage `json:"history"`
}

// AssistantChat 调用 RAG 服务完成问答。
func AssistantChat(ctx *gin.Context) {
	var request assistantChatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply, err := services.RAG.AnswerQuestion(ctx.Request.Context(), request.Question, request.History)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, reply)
}

// AssistantStatus 返回助手可用性与索引状态。
func AssistantStatus(ctx *gin.Context) {
	status, err := services.RAG.Status(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, status)
}
