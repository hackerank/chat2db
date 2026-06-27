package api

import (
	"chat2db/internal/chat"
	"chat2db/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatHandler holds the orchestrator dependency
type ChatHandler struct {
	orchestrator *chat.Orchestrator
}

// NewChatHandler creates a new handler instance
func NewChatHandler(o *chat.Orchestrator) *ChatHandler {
	return &ChatHandler{orchestrator: o}
}

// ChatRequest defines the expected JSON payload
type ChatRequest struct {
	Schema models.DatabaseSchema `json:"schema"`
	Prompt string                `json:"prompt"`
}

// HandleChat processes the incoming chat request
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Execute the pipeline
	results, err := h.orchestrator.ProcessPrompt(c.Request.Context(), req.Schema, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Pipeline execution failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": results,
	})
}
