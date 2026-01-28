package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MiguelVivar/insta-mock/internal/generator"
	"github.com/MiguelVivar/insta-mock/internal/server"
	"github.com/spf13/cobra"
)

var (
	port    string
	count   int
	watch   bool
	chaos   bool
	version = "0.2.0"
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
	serveCmd.Flags().IntVarP(&count, "count", "c", 0, "Generate N additional fake items per resource")
	serveCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch JSON file for changes (hot-reload)")
	serveCmd.Flags().BoolVar(&chaos, "chaos", false, "Enable chaos mode (random failures/latency)")

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

	// Generate additional fake data
	if count > 0 {
		data = generator.ExpandData(data, count)
	}

	// Count resources and items
	resourceCount := 0
	totalItems := 0
	for _, v := range data {
		if arr, ok := v.([]interface{}); ok {
			resourceCount++
			totalItems += len(arr)
		} else if _, ok := v.(map[string]interface{}); ok {
			resourceCount++
			totalItems++
		}
	}

	// Create engine with config
	config := server.EngineConfig{
		EnableLogger: true,
		ChaosMode:    chaos,
		ChaosPercent: 15,
	}
	engine := server.NewEngineWithConfig(data, config)

	// Print banner
	fmt.Println()
	fmt.Println("  🚀 \033[1;36mInsta-Mock\033[0m \033[90mv" + version + "\033[0m")
	fmt.Println("  \033[90m─────────────────────────────────────\033[0m")
	fmt.Printf("  📁 File:      \033[33m%s\033[0m\n", filePath)
	fmt.Printf("  📦 Resources: \033[32m%d\033[0m\n", resourceCount)
	fmt.Printf("  📊 Items:     \033[32m%d\033[0m", totalItems)
	if count > 0 {
		fmt.Printf(" \033[90m(+%d generated)\033[0m", count*resourceCount)
	}
	fmt.Println()
	fmt.Printf("  🌐 Server:    \033[1;32mhttp://localhost:%s\033[0m\n", port)

	// Feature flags
	features := []string{}
	if watch {
		features = append(features, "🔄 hot-reload")
	}
	if chaos {
		features = append(features, "💥 chaos")
	}
	if len(features) > 0 {
		fmt.Printf("  ⚡ Features:  %s\n", features[0])
		for i := 1; i < len(features); i++ {
			fmt.Printf("              %s\n", features[i])
		}
	}

	fmt.Println("  \033[90m─────────────────────────────────────\033[0m")
	fmt.Println()
	fmt.Println("  \033[1mEndpoints:\033[0m")
	for key, v := range data {
		itemCount := 0
		if arr, ok := v.([]interface{}); ok {
			itemCount = len(arr)
		}
		fmt.Printf("    \033[36m%-12s\033[0m \033[90m%d items\033[0m\n", "/"+key, itemCount)
	}

	fmt.Println()
	fmt.Println("  \033[1mQuery Parameters:\033[0m")
	fmt.Println("    \033[90m?_page=1&_limit=10  Pagination\033[0m")
	fmt.Println("    \033[90m?_sort=name&_order=desc  Sorting\033[0m")
	fmt.Println("    \033[90m?q=keyword  Full-text search\033[0m")
	fmt.Println("    \033[90m?field=value  Filter by field\033[0m")
	fmt.Println()
	fmt.Println("  \033[90mPress Ctrl+C to stop\033[0m")
	fmt.Println()

	// Setup hot-reload watcher
	if watch {
		watcher, err := server.NewWatcher(filePath, engine)
		if err != nil {
			fmt.Printf("  ⚠️  \033[33mHot-reload unavailable: %v\033[0m\n", err)
		} else {
			watcher.SetOnChange(func(msg string) {
				fmt.Printf("  %s\n", msg)
			})
			if err := watcher.Start(); err != nil {
				fmt.Printf("  ⚠️  \033[33mHot-reload failed: %v\033[0m\n", err)
			} else {
				defer watcher.Stop()
			}
		}
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\n  \033[33mShutting down...\033[0m")
		engine.Shutdown()
	}()

	// Start server
	return engine.Start(":" + port)
}
