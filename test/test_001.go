package main

import (
	"chat2db/internal/chat"
	"chat2db/internal/execution"
	"chat2db/internal/formatter"
	"chat2db/internal/generator"
	"chat2db/internal/schema"
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	// 1. Load Configuration
	dsn := os.Getenv("DB_DSN")
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("LLM_MODEL")
	dbName := os.Getenv("DB_NAME") // Ensure you set this

	if dsn == "" || apiKey == "" || dbName == "" {
		log.Fatal("Please set DB_DSN, OPENAI_API_KEY, LLM_MODEL, and DB_NAME environment variables")
	}

	fmt.Println("--- Starting Pipeline Test ---")

	// 2. Step A: Sync Schema
	fmt.Println("Step 1: Syncing Schema...")
	cfg := schema.DBConfig{
		Host:     "127.0.0.1", // Adjust if needed
		Port:     3306,        // Adjust if needed
		User:     "root",      // Adjust if needed
		Password: "password",  // Adjust if needed
		DBName:   dbName,
	}
	
	dbSchema, err := schema.SyncDatabase(cfg)
	if err != nil {
		log.Fatalf("Schema Sync failed: %v", err)
	}
	fmt.Printf("Successfully synced %d tables.\n", len(dbSchema.Tables))

	// 3. Initialize Services
	exec, err := execution.NewExecutor(dsn)
	if err != nil {
		log.Fatalf("Executor init failed: %v", err)
	}

	gen := generator.NewGenerator(apiKey, model)
	fmt := formatter.NewFormatter(apiKey, model)
	orch := chat.NewOrchestrator(gen, exec, fmt)

	// 4. Run Test Prompt
	testPrompt := "How many users are in the database?"
	fmt.Printf("Step 2: Running Prompt: '%s'\n", testPrompt)

	ctx := context.Background()
	result, err := orch.ProcessPrompt(ctx, *dbSchema, testPrompt)
	if err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	// 5. Output Result
	fmt.Println("--- Test Result ---")
	fmt.Println(result)
	fmt.Println("--- Test Complete ---")
}
