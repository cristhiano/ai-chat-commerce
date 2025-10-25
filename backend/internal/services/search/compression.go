package search

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// CompressionService handles search result compression
type CompressionService struct {
	compressionLevel int
}

// NewCompressionService creates a new compression service
func NewCompressionService(compressionLevel int) *CompressionService {
	if compressionLevel < 1 || compressionLevel > 9 {
		compressionLevel = 6 // Default compression level
	}
	return &CompressionService{
		compressionLevel: compressionLevel,
	}
}

// CompressSearchResults compresses search results
func (cs *CompressionService) CompressSearchResults(results *SearchResult) ([]byte, error) {
	// Serialize results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search results: %w", err)
	}

	// Compress the JSON data
	compressedData, err := cs.compressData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress search results: %w", err)
	}

	return compressedData, nil
}

// DecompressSearchResults decompresses search results
func (cs *CompressionService) DecompressSearchResults(compressedData []byte) (*SearchResult, error) {
	// Decompress the data
	jsonData, err := cs.decompressData(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress search results: %w", err)
	}

	// Deserialize JSON data
	var results SearchResult
	err = json.Unmarshal(jsonData, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return &results, nil
}

// CompressSuggestions compresses suggestions
func (cs *CompressionService) CompressSuggestions(suggestions []string) ([]byte, error) {
	// Serialize suggestions to JSON
	jsonData, err := json.Marshal(suggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestions: %w", err)
	}

	// Compress the JSON data
	compressedData, err := cs.compressData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress suggestions: %w", err)
	}

	return compressedData, nil
}

// DecompressSuggestions decompresses suggestions
func (cs *CompressionService) DecompressSuggestions(compressedData []byte) ([]string, error) {
	// Decompress the data
	jsonData, err := cs.decompressData(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress suggestions: %w", err)
	}

	// Deserialize JSON data
	var suggestions []string
	err = json.Unmarshal(jsonData, &suggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal suggestions: %w", err)
	}

	return suggestions, nil
}

// CompressFilterOptions compresses filter options
func (cs *CompressionService) CompressFilterOptions(options *FilterOptions) ([]byte, error) {
	// Serialize options to JSON
	jsonData, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter options: %w", err)
	}

	// Compress the JSON data
	compressedData, err := cs.compressData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress filter options: %w", err)
	}

	return compressedData, nil
}

// DecompressFilterOptions decompresses filter options
func (cs *CompressionService) DecompressFilterOptions(compressedData []byte) (*FilterOptions, error) {
	// Decompress the data
	jsonData, err := cs.decompressData(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress filter options: %w", err)
	}

	// Deserialize JSON data
	var options FilterOptions
	err = json.Unmarshal(jsonData, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal filter options: %w", err)
	}

	return &options, nil
}

// compressData compresses data using gzip
func (cs *CompressionService) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// Create gzip writer with compression level
	gzWriter, err := gzip.NewWriterLevel(&buf, cs.compressionLevel)
	if err != nil {
		return nil, err
	}
	defer gzWriter.Close()

	// Write data to gzip writer
	_, err = gzWriter.Write(data)
	if err != nil {
		return nil, err
	}

	// Close to flush remaining data
	err = gzWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses data using gzip
func (cs *CompressionService) decompressData(compressedData []byte) ([]byte, error) {
	// Create gzip reader
	gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	// Read all decompressed data
	decompressedData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	return decompressedData, nil
}

// GetCompressionRatio calculates compression ratio
func (cs *CompressionService) GetCompressionRatio(originalSize, compressedSize int) float64 {
	if originalSize == 0 {
		return 0.0
	}
	return float64(compressedSize) / float64(originalSize) * 100
}

// GetCompressionStats returns compression statistics
func (cs *CompressionService) GetCompressionStats(originalSize, compressedSize int) *CompressionStats {
	return &CompressionStats{
		OriginalSize:      originalSize,
		CompressedSize:    compressedSize,
		CompressionRatio:  cs.GetCompressionRatio(originalSize, compressedSize),
		SpaceSaved:        originalSize - compressedSize,
		SpaceSavedPercent: float64(originalSize-compressedSize) / float64(originalSize) * 100,
		GeneratedAt:       time.Now(),
	}
}

// ShouldCompress determines if data should be compressed
func (cs *CompressionService) ShouldCompress(dataSize int, threshold int) bool {
	if threshold <= 0 {
		threshold = 1024 // Default 1KB threshold
	}
	return dataSize > threshold
}

// CompressionStats represents compression statistics
type CompressionStats struct {
	OriginalSize      int       `json:"original_size"`
	CompressedSize    int       `json:"compressed_size"`
	CompressionRatio  float64   `json:"compression_ratio"`
	SpaceSaved        int       `json:"space_saved"`
	SpaceSavedPercent float64   `json:"space_saved_percent"`
	GeneratedAt       time.Time `json:"generated_at"`
}
