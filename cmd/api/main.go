package main

import (
	"chat2db/internal/api"
	"chat2db/internal/chat"
	"chat2db/internal/execution"
	"chat2db/internal/generator"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration (In production, use environment variables or a config file)
	dsn := os.Getenv("DB_DSN")
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("LLM_MODEL")

	if dsn == "" || apiKey == "" {
		log.Fatal("DB_DSN and OPENAI_API_KEY must be set")
	}

	// 2. Initialize Services
	exec, err := execution.NewExecutor(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize Executor: %v", err)
	}

	gen := generator.NewGenerator(apiKey, model)
	orch := chat.NewOrchestrator(gen, exec)
	handler := api.NewChatHandler(orch)

	// 3. Setup Router
	r := gin.Default()

	// API Routes
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/chat", handler.HandleChat)
	}

	// 4. Start Server
	log.Println("Starting Chat2DB API server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
