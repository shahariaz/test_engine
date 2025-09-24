package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for data generation
type Config struct {
	// Generation parameters
	Customers int `yaml:"customers" json:"customers"`
	Contents  int `yaml:"contents" json:"contents"`
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// Output configuration
	DgraphURL string `yaml:"dgraph_url" json:"dgraph_url"`
	OutputDir string `yaml:"output_dir" json:"output_dir"`
	Format    string `yaml:"format" json:"format"` // dgraph, json, csv

	// Processing options
	ValidateData bool `yaml:"validate" json:"validate"`
	Verbose      bool `yaml:"verbose" json:"verbose"`
	DryRun       bool `yaml:"dry_run" json:"dry_run"`

	// Data distribution
	Distribution DataDistribution `yaml:"distribution" json:"distribution"`

	// Internal
	ConfigFile string `yaml:"-" json:"-"`
}

// DataDistribution defines how data should be distributed
type DataDistribution struct {
	MinSubscriptionsPerCustomer  int     `yaml:"min_subscriptions_per_customer" json:"min_subscriptions_per_customer"`
	MaxSubscriptionsPerCustomer  int     `yaml:"max_subscriptions_per_customer" json:"max_subscriptions_per_customer"`
	MinWatchHistoriesPerCustomer int     `yaml:"min_watch_histories_per_customer" json:"min_watch_histories_per_customer"`
	MaxWatchHistoriesPerCustomer int     `yaml:"max_watch_histories_per_customer" json:"max_watch_histories_per_customer"`
	MinDevicesPerCustomer        int     `yaml:"min_devices_per_customer" json:"min_devices_per_customer"`
	MaxDevicesPerCustomer        int     `yaml:"max_devices_per_customer" json:"max_devices_per_customer"`
	ActiveCustomerRatio          float64 `yaml:"active_customer_ratio" json:"active_customer_ratio"`
	PremiumContentRatio          float64 `yaml:"premium_content_ratio" json:"premium_content_ratio"`
}

// LoadFromFile loads configuration from YAML file
func (c *Config) LoadFromFile() error {
	if c.ConfigFile == "" {
		return nil
	}

	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// SaveToFile saves current configuration to YAML file
func (c *Config) SaveToFile(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	if c.DgraphURL == "" {
		c.DgraphURL = "localhost:9080"
	}
	if c.OutputDir == "" {
		c.OutputDir = "./output"
	}
	if c.Format == "" {
		c.Format = "dgraph"
	}

	// Set default distribution
	if c.Distribution.MinSubscriptionsPerCustomer == 0 {
		c.Distribution.MinSubscriptionsPerCustomer = 1
	}
	if c.Distribution.MaxSubscriptionsPerCustomer == 0 {
		c.Distribution.MaxSubscriptionsPerCustomer = 3
	}
	if c.Distribution.MinWatchHistoriesPerCustomer == 0 {
		c.Distribution.MinWatchHistoriesPerCustomer = 2
	}
	if c.Distribution.MaxWatchHistoriesPerCustomer == 0 {
		c.Distribution.MaxWatchHistoriesPerCustomer = 15
	}
	if c.Distribution.MinDevicesPerCustomer == 0 {
		c.Distribution.MinDevicesPerCustomer = 1
	}
	if c.Distribution.MaxDevicesPerCustomer == 0 {
		c.Distribution.MaxDevicesPerCustomer = 4
	}
	if c.Distribution.ActiveCustomerRatio == 0 {
		c.Distribution.ActiveCustomerRatio = 0.85
	}
	if c.Distribution.PremiumContentRatio == 0 {
		c.Distribution.PremiumContentRatio = 0.4
	}
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	if c.Customers <= 0 {
		return fmt.Errorf("customers must be positive, got %d", c.Customers)
	}
	if c.Contents <= 0 {
		return fmt.Errorf("contents must be positive, got %d", c.Contents)
	}
	if c.BatchSize <= 0 || c.BatchSize > 10000 {
		return fmt.Errorf("batch_size must be between 1 and 10000, got %d", c.BatchSize)
	}
	if c.Format != "dgraph" && c.Format != "json" && c.Format != "csv" {
		return fmt.Errorf("format must be 'dgraph', 'json', or 'csv', got %s", c.Format)
	}

	// Validate distribution
	dist := c.Distribution
	if dist.MinSubscriptionsPerCustomer > dist.MaxSubscriptionsPerCustomer {
		return fmt.Errorf("min_subscriptions_per_customer cannot be greater than max_subscriptions_per_customer")
	}
	if dist.MinWatchHistoriesPerCustomer > dist.MaxWatchHistoriesPerCustomer {
		return fmt.Errorf("min_watch_histories_per_customer cannot be greater than max_watch_histories_per_customer")
	}
	if dist.MinDevicesPerCustomer > dist.MaxDevicesPerCustomer {
		return fmt.Errorf("min_devices_per_customer cannot be greater than max_devices_per_customer")
	}
	if dist.ActiveCustomerRatio < 0 || dist.ActiveCustomerRatio > 1 {
		return fmt.Errorf("active_customer_ratio must be between 0 and 1, got %f", dist.ActiveCustomerRatio)
	}
	if dist.PremiumContentRatio < 0 || dist.PremiumContentRatio > 1 {
		return fmt.Errorf("premium_content_ratio must be between 0 and 1, got %f", dist.PremiumContentRatio)
	}

	return nil
}

// Generator is the main data generation engine
type Generator struct {
	config    *Config
	progress  *ProgressTracker
	validator *Validator
	exporter  *Exporter
	logger    *log.Logger
}

// NewGenerator creates a new data generator
func NewGenerator(config *Config) (*Generator, error) {
	config.SetDefaults()
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup logger
	logger := log.New(os.Stdout, "[DATAGEN] ", log.LstdFlags)
	if !config.Verbose {
		logger.SetOutput(io.Discard)
	}

	// Create components
	progress := NewProgressTracker(logger)
	validator := NewValidator(logger)
	exporter := NewExporter(config, logger)

	return &Generator{
		config:    config,
		progress:  progress,
		validator: validator,
		exporter:  exporter,
		logger:    logger,
	}, nil
}

// Generate orchestrates the data generation process
func (g *Generator) Generate() error {
	ctx := context.Background()
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	g.logger.Printf("Starting data generation with config: %+v", g.config)
	
	totalItems := g.config.Customers + g.config.Contents
	g.progress.Start(totalItems)
	defer g.progress.Finish()

	// Generate customers with relationships
	if err := g.generateCustomers(ctx); err != nil {
		return fmt.Errorf("failed to generate customers: %w", err)
	}

	// Generate contents
	if err := g.generateContents(ctx); err != nil {
		return fmt.Errorf("failed to generate contents: %w", err)
	}

	// Export final summary
	if err := g.exporter.ExportSummary(); err != nil {
		g.logger.Printf("Warning: failed to export summary: %v", err)
	}

	return nil
}

// generateCustomers generates customer data in batches
func (g *Generator) generateCustomers(ctx context.Context) error {
	g.logger.Printf("Generating %d customers in batches of %d", g.config.Customers, g.config.BatchSize)

	batches := (g.config.Customers + g.config.BatchSize - 1) / g.config.BatchSize
	
	for batch := 0; batch < batches; batch++ {
		start := batch * g.config.BatchSize
		end := start + g.config.BatchSize
		if end > g.config.Customers {
			end = g.config.Customers
		}

		batchData := g.generateCustomerBatch(start, end)
		
		// Validate batch
		if g.config.ValidateData {
			if err := g.validator.ValidateBatch(batchData); err != nil {
				return fmt.Errorf("validation failed for batch %d: %w", batch+1, err)
			}
		}

		// Export batch
		if err := g.exporter.ExportBatch(batchData, fmt.Sprintf("customers_batch_%d", batch+1)); err != nil {
			return fmt.Errorf("failed to export batch %d: %w", batch+1, err)
		}

		g.progress.Update(end - start)
		g.logger.Printf("Completed customer batch %d/%d (%d-%d)", batch+1, batches, start+1, end)
	}

	return nil
}

// generateContents generates content data in batches
func (g *Generator) generateContents(ctx context.Context) error {
	g.logger.Printf("Generating %d contents in batches of %d", g.config.Contents, g.config.BatchSize)

	batches := (g.config.Contents + g.config.BatchSize - 1) / g.config.BatchSize
	
	for batch := 0; batch < batches; batch++ {
		start := batch * g.config.BatchSize
		end := start + g.config.BatchSize
		if end > g.config.Contents {
			end = g.config.Contents
		}

		batchData := g.generateContentBatch(start, end)
		
		// Validate batch
		if g.config.ValidateData {
			if err := g.validator.ValidateBatch(batchData); err != nil {
				return fmt.Errorf("validation failed for batch %d: %w", batch+1, err)
			}
		}

		// Export batch
		if err := g.exporter.ExportBatch(batchData, fmt.Sprintf("contents_batch_%d", batch+1)); err != nil {
			return fmt.Errorf("failed to export batch %d: %w", batch+1, err)
		}

		g.progress.Update(end - start)
		g.logger.Printf("Completed content batch %d/%d (%d-%d)", batch+1, batches, start+1, end)
	}

	return nil
}

// generateCustomerBatch generates a batch of customers with all relationships
func (g *Generator) generateCustomerBatch(start, end int) []map[string]interface{} {
	var batch []map[string]interface{}
	
	for i := start; i < end; i++ {
		customer := g.generateSingleCustomer(i + 1)
		batch = append(batch, customer)
	}
	
	return batch
}

// generateContentBatch generates a batch of contents
func (g *Generator) generateContentBatch(start, end int) []map[string]interface{} {
	var batch []map[string]interface{}
	
	for i := start; i < end; i++ {
		content := g.generateSingleContent(i + 1)
		batch = append(batch, content)
	}
	
	return batch
}

// Helper methods for random data generation
func (g *Generator) randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.Intn(max)
}

func (g *Generator) randomFloat() float64 {
	return rand.Float64()
}

// DgraphUploader handles uploading data to Dgraph
type DgraphUploader struct {
	baseURL string
	client  *http.Client
	logger  *log.Logger
}

// NewDgraphUploader creates a new Dgraph uploader
func NewDgraphUploader(baseURL string, logger *log.Logger) *DgraphUploader {
	return &DgraphUploader{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// UploadBatch uploads a batch of data to Dgraph
func (du *DgraphUploader) UploadBatch(data []map[string]interface{}) error {
	mutation := map[string]interface{}{
		"set": data,
	}

	jsonData, err := json.Marshal(mutation)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	url := fmt.Sprintf("http://%s/mutate?commitNow=true", du.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := du.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	du.logger.Printf("Successfully uploaded batch with %d items", len(data))
	return nil
}

// TestConnection tests the connection to Dgraph
func (du *DgraphUploader) TestConnection() error {
	url := fmt.Sprintf("http://%s/health", du.baseURL)
	resp, err := du.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Dgraph health check failed with status %d", resp.StatusCode)
	}

	return nil
}