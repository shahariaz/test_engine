package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"datagen/pkg"
)

func main() {
	var (
		configFile = flag.String("config", "configs/datagen.yaml", "Configuration file path")
		customers  = flag.Int("customers", 1000, "Number of customers to generate")
		contents   = flag.Int("contents", 500, "Number of contents to generate")
		batchSize  = flag.Int("batch", 100, "Batch size for processing")
		dgraphURL  = flag.String("dgraph", "localhost:9080", "Dgraph Alpha URL")
		outputDir  = flag.String("output", "./output", "Output directory for generated files")
		format     = flag.String("format", "dgraph", "Output format: dgraph, json, csv")
		validate   = flag.Bool("validate", true, "Validate generated data")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		dryRun     = flag.Bool("dry-run", false, "Generate data without uploading to Dgraph")
	)
	flag.Parse()

	// Display banner
	fmt.Println("🚀 Chorki Data Generator - Production Ready")
	fmt.Println("==========================================")

	// Setup logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize configuration
	config := &pkg.Config{
		Customers:    *customers,
		Contents:     *contents,
		BatchSize:    *batchSize,
		DgraphURL:    *dgraphURL,
		OutputDir:    *outputDir,
		Format:       *format,
		ValidateData: *validate,
		Verbose:      *verbose,
		DryRun:       *dryRun,
		ConfigFile:   *configFile,
	}

	// Load configuration file if exists
	configPath := *configFile
	if !filepath.IsAbs(configPath) {
		// If relative path, make it relative to the datagen directory
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			// Try to find the datagen directory
			for dir := exeDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
				if filepath.Base(dir) == "datagen" {
					configPath = filepath.Join(dir, configPath)
					break
				}
			}
		}
	}

	config.ConfigFile = configPath
	if err := config.LoadFromFile(); err != nil {
		log.Printf("Warning: Could not load config file %s: %v", configPath, err)
	}

	// Create generator
	generator, err := pkg.NewGenerator(config)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Generate data
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   - Customers: %d\n", config.Customers)
	fmt.Printf("   - Contents: %d\n", config.Contents)
	fmt.Printf("   - Batch Size: %d\n", config.BatchSize)
	fmt.Printf("   - Format: %s\n", config.Format)
	fmt.Printf("   - Output: %s\n", config.OutputDir)
	fmt.Printf("   - Dry Run: %t\n", config.DryRun)
	fmt.Println()

	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate data: %v", err)
	}

	fmt.Println("✅ Data generation completed successfully!")
	fmt.Printf("📁 Check output directory: %s\n", *outputDir)
}
