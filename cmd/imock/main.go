package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MiguelVivar/insta-mock/internal/server"
	"github.com/spf13/cobra"
)

var (
	port    string
	version = "0.1.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "imock",
		Short:   "🚀 Insta-Mock - Tu Backend listo en lo que tardas en parpadear",
		Version: version,
	}

	serveCmd := &cobra.Command{
		Use:   "serve <json-file>",
		Short: "Start the mock API server from a JSON file",
		Args:  cobra.ExactArgs(1),
		RunE:  runServe,
	}

	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to run the server on")

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read JSON file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ Error reading file '%s': %w", filePath, err)
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("❌ Invalid JSON in '%s': %w", filePath, err)
	}

	// Count resources
	resourceCount := 0
	for _, v := range data {
		if _, ok := v.([]interface{}); ok {
			resourceCount++
		} else if _, ok := v.(map[string]interface{}); ok {
			resourceCount++
		}
	}

	// Print startup banner
	fmt.Println()
	fmt.Println("  🚀 \033[1;36mInsta-Mock\033[0m")
	fmt.Println("  \033[90m───────────────────────────────\033[0m")
	fmt.Printf("  📁 File:      \033[33m%s\033[0m\n", filePath)
	fmt.Printf("  📦 Resources: \033[32m%d\033[0m\n", resourceCount)
	fmt.Printf("  🌐 Server:    \033[1;32mhttp://localhost:%s\033[0m\n", port)
	fmt.Println("  \033[90m───────────────────────────────\033[0m")
	fmt.Println()
	fmt.Println("  \033[90mEndpoints:\033[0m")
	for key := range data {
		fmt.Printf("    • \033[36m/%s\033[0m\n", key)
	}
	fmt.Println()
	fmt.Println("  \033[90mPress Ctrl+C to stop\033[0m")
	fmt.Println()

	// Start server
	engine := server.NewEngine(data)
	return engine.Start(":" + port)
}
