package pkg

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Exporter handles exporting data in various formats
type Exporter struct {
	config   *Config
	logger   *log.Logger
	summary  *GenerationSummary
	uploader *DgraphUploader
}

// GenerationSummary tracks statistics about generated data
type GenerationSummary struct {
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	Duration           time.Duration `json:"duration"`
	TotalCustomers     int           `json:"total_customers"`
	TotalContents      int           `json:"total_contents"`
	TotalSubscriptions int           `json:"total_subscriptions"`
	TotalDevices       int           `json:"total_devices"`
	TotalWatchHistory  int           `json:"total_watch_history"`
	TotalRecords       int           `json:"total_records"`
	BatchesProcessed   int           `json:"batches_processed"`
	ErrorsEncountered  int           `json:"errors_encountered"`
	CountriesUsed      []string      `json:"countries_used"`
	DeviceTypesUsed    []string      `json:"device_types_used"`
	PackagesUsed       []string      `json:"packages_used"`
}

// NewExporter creates a new exporter
func NewExporter(config *Config, logger *log.Logger) *Exporter {
	exporter := &Exporter{
		config: config,
		logger: logger,
		summary: &GenerationSummary{
			StartTime: time.Now(),
		},
	}

	// Initialize Dgraph uploader if not in dry-run mode
	if !config.DryRun && config.Format == "dgraph" {
		exporter.uploader = NewDgraphUploader(config.DgraphURL, logger)
	}

	return exporter
}

// ExportBatch exports a batch of data
func (e *Exporter) ExportBatch(data []map[string]interface{}, batchName string) error {
	e.summary.BatchesProcessed++

	switch e.config.Format {
	case "dgraph":
		return e.exportDgraphBatch(data, batchName)
	case "json":
		return e.exportJSONBatch(data, batchName)
	case "csv":
		return e.exportCSVBatch(data, batchName)
	default:
		return fmt.Errorf("unsupported format: %s", e.config.Format)
	}
}

// exportDgraphBatch exports data to Dgraph
func (e *Exporter) exportDgraphBatch(data []map[string]interface{}, batchName string) error {
	// Update summary statistics
	e.updateSummaryFromBatch(data)

	if e.config.DryRun {
		e.logger.Printf("DRY RUN: Would upload batch %s with %d items to Dgraph", batchName, len(data))
		return e.saveBatchToFile(data, batchName, "json")
	}

	// Test connection before uploading
	if err := e.uploader.TestConnection(); err != nil {
		return fmt.Errorf("Dgraph connection test failed: %w", err)
	}

	// Upload to Dgraph
	if err := e.uploader.UploadBatch(data); err != nil {
		e.summary.ErrorsEncountered++

		// Save failed batch to file for debugging
		if err := e.saveBatchToFile(data, fmt.Sprintf("%s_failed", batchName), "json"); err != nil {
			e.logger.Printf("Failed to save error batch: %v", err)
		}

		return fmt.Errorf("failed to upload batch to Dgraph: %w", err)
	}

	// Also save a copy to file for backup
	return e.saveBatchToFile(data, batchName, "json")
}

// exportJSONBatch exports data to JSON file
func (e *Exporter) exportJSONBatch(data []map[string]interface{}, batchName string) error {
	e.updateSummaryFromBatch(data)
	return e.saveBatchToFile(data, batchName, "json")
}

// exportCSVBatch exports data to CSV file
func (e *Exporter) exportCSVBatch(data []map[string]interface{}, batchName string) error {
	e.updateSummaryFromBatch(data)

	// Group by entity type
	entityGroups := make(map[string][]map[string]interface{})

	for _, item := range data {
		if dgraphType, exists := item["dgraph.type"]; exists {
			if typeSlice, ok := dgraphType.([]interface{}); ok && len(typeSlice) > 0 {
				if entityType, ok := typeSlice[0].(string); ok {
					entityGroups[entityType] = append(entityGroups[entityType], item)
				}
			}
		}
	}

	// Export each entity type to separate CSV
	for entityType, items := range entityGroups {
		filename := fmt.Sprintf("%s_%s.csv", batchName, entityType)
		if err := e.saveCSVFile(items, filename); err != nil {
			return fmt.Errorf("failed to export %s CSV: %w", entityType, err)
		}
	}

	return nil
}

// saveBatchToFile saves batch data to a file
func (e *Exporter) saveBatchToFile(data []map[string]interface{}, batchName, format string) error {
	filename := fmt.Sprintf("%s.%s", batchName, format)
	filepath := filepath.Join(e.config.OutputDir, filename)

	var exportData interface{}
	if format == "json" && e.config.Format == "dgraph" {
		// Wrap in set for Dgraph format
		exportData = map[string]interface{}{
			"set": data,
		}
	} else {
		exportData = data
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	e.logger.Printf("Saved batch to: %s", filepath)
	return nil
}

// saveCSVFile saves data to CSV format
func (e *Exporter) saveCSVFile(data []map[string]interface{}, filename string) error {
	if len(data) == 0 {
		return nil
	}

	filepath := filepath.Join(e.config.OutputDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get all unique headers
	headers := make(map[string]bool)
	for _, item := range data {
		for key := range item {
			headers[key] = true
		}
	}

	// Convert to sorted slice
	var headerSlice []string
	for header := range headers {
		headerSlice = append(headerSlice, header)
	}

	// Write headers
	if err := writer.Write(headerSlice); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, item := range data {
		var row []string
		for _, header := range headerSlice {
			value := ""
			if val, exists := item[header]; exists {
				value = fmt.Sprintf("%v", val)
			}
			row = append(row, value)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	e.logger.Printf("Saved CSV to: %s", filepath)
	return nil
}

// updateSummaryFromBatch updates generation summary from batch data
func (e *Exporter) updateSummaryFromBatch(data []map[string]interface{}) {
	for _, item := range data {
		if dgraphType, exists := item["dgraph.type"]; exists {
			if typeSlice, ok := dgraphType.([]interface{}); ok && len(typeSlice) > 0 {
				if entityType, ok := typeSlice[0].(string); ok {
					switch entityType {
					case "chorki_customers":
						e.summary.TotalCustomers++
						// Track countries
						if country, exists := item["chorki_customers.country"]; exists {
							if countryStr, ok := country.(string); ok {
								e.addUniqueString(&e.summary.CountriesUsed, countryStr)
							}
						}
						// Track device types
						if device, exists := item["chorki_customers.device"]; exists {
							if deviceStr, ok := device.(string); ok {
								e.addUniqueString(&e.summary.DeviceTypesUsed, deviceStr)
							}
						}

						// Count nested subscriptions
						if subscriptions, exists := item["chorki_customers.subscriptions"]; exists {
							if subSlice, ok := subscriptions.([]map[string]interface{}); ok {
								e.summary.TotalSubscriptions += len(subSlice)
								// Track packages
								for _, sub := range subSlice {
									if pkg, exists := sub["chorki_subscriptions.package"]; exists {
										if pkgStr, ok := pkg.(string); ok {
											e.addUniqueString(&e.summary.PackagesUsed, pkgStr)
										}
									}
								}
							}
						}

						// Count nested devices
						if devices, exists := item["chorki_customers.devices"]; exists {
							if devSlice, ok := devices.([]map[string]interface{}); ok {
								e.summary.TotalDevices += len(devSlice)
							}
						}

						// Count nested watch histories
						if watchHistories, exists := item["chorki_customers.watch_histories"]; exists {
							if watchSlice, ok := watchHistories.([]map[string]interface{}); ok {
								e.summary.TotalWatchHistory += len(watchSlice)
							}
						}

					case "chorki_contents":
						e.summary.TotalContents++
					case "chorki_subscriptions":
						e.summary.TotalSubscriptions++
						// Track packages
						if pkg, exists := item["chorki_subscriptions.package"]; exists {
							if pkgStr, ok := pkg.(string); ok {
								e.addUniqueString(&e.summary.PackagesUsed, pkgStr)
							}
						}
					case "chorki_devices":
						e.summary.TotalDevices++
					case "chorki_watch_histories":
						e.summary.TotalWatchHistory++
					}
				}
			}
		}
		e.summary.TotalRecords++
	}
}

// addUniqueString adds a string to slice if not already present
func (e *Exporter) addUniqueString(slice *[]string, value string) {
	for _, existing := range *slice {
		if existing == value {
			return
		}
	}
	*slice = append(*slice, value)
}

// ExportSummary exports the final generation summary
func (e *Exporter) ExportSummary() error {
	e.summary.EndTime = time.Now()
	e.summary.Duration = e.summary.EndTime.Sub(e.summary.StartTime)

	// Create summary report
	summaryPath := filepath.Join(e.config.OutputDir, "generation_summary.json")
	summaryData, err := json.MarshalIndent(e.summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := os.WriteFile(summaryPath, summaryData, 0644); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	// Create human-readable report
	reportPath := filepath.Join(e.config.OutputDir, "generation_report.txt")
	report := e.generateTextReport()

	if err := os.WriteFile(reportPath, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	e.logger.Printf("Summary saved to: %s", summaryPath)
	e.logger.Printf("Report saved to: %s", reportPath)

	return nil
}

// generateTextReport creates a human-readable report
func (e *Exporter) generateTextReport() string {
	var report strings.Builder

	report.WriteString("=== DATA GENERATION REPORT ===\n\n")
	report.WriteString(fmt.Sprintf("Generated at: %s\n", e.summary.EndTime.Format(time.RFC3339)))
	report.WriteString(fmt.Sprintf("Duration: %v\n", e.summary.Duration.Round(time.Second)))
	report.WriteString(fmt.Sprintf("Batches processed: %d\n", e.summary.BatchesProcessed))

	if e.summary.ErrorsEncountered > 0 {
		report.WriteString(fmt.Sprintf("Errors encountered: %d\n", e.summary.ErrorsEncountered))
	}

	report.WriteString("\n=== ENTITY COUNTS ===\n")
	report.WriteString(fmt.Sprintf("Customers: %d\n", e.summary.TotalCustomers))
	report.WriteString(fmt.Sprintf("Contents: %d\n", e.summary.TotalContents))
	report.WriteString(fmt.Sprintf("Subscriptions: %d\n", e.summary.TotalSubscriptions))
	report.WriteString(fmt.Sprintf("Devices: %d\n", e.summary.TotalDevices))
	report.WriteString(fmt.Sprintf("Watch histories: %d\n", e.summary.TotalWatchHistory))
	report.WriteString(fmt.Sprintf("Total records: %d\n", e.summary.TotalRecords))

	if len(e.summary.CountriesUsed) > 0 {
		report.WriteString(fmt.Sprintf("\nCountries (%d): %s\n",
			len(e.summary.CountriesUsed), strings.Join(e.summary.CountriesUsed, ", ")))
	}

	if len(e.summary.DeviceTypesUsed) > 0 {
		report.WriteString(fmt.Sprintf("Device types (%d): %s\n",
			len(e.summary.DeviceTypesUsed), strings.Join(e.summary.DeviceTypesUsed, ", ")))
	}

	if len(e.summary.PackagesUsed) > 0 {
		report.WriteString(fmt.Sprintf("Packages (%d): %s\n",
			len(e.summary.PackagesUsed), strings.Join(e.summary.PackagesUsed, ", ")))
	}

	report.WriteString("\n=== CONFIGURATION ===\n")
	report.WriteString(fmt.Sprintf("Output format: %s\n", e.config.Format))
	report.WriteString(fmt.Sprintf("Output directory: %s\n", e.config.OutputDir))
	report.WriteString(fmt.Sprintf("Batch size: %d\n", e.config.BatchSize))
	report.WriteString(fmt.Sprintf("Validation enabled: %t\n", e.config.ValidateData))
	report.WriteString(fmt.Sprintf("Dry run: %t\n", e.config.DryRun))

	if !e.config.DryRun && e.config.Format == "dgraph" {
		report.WriteString(fmt.Sprintf("Dgraph URL: %s\n", e.config.DgraphURL))
	}

	return report.String()
}
