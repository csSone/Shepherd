package gguf

// Metadata represents the parsed GGUF model metadata
// It contains the most important fields needed for model management
type Metadata struct {
	// Basic information
	Name         string
	Architecture string
	Quantization string

	// Model parameters
	Parameters        float64 // Number of parameters (in billions)
	ContextLength     int     // Maximum context length
	EmbeddingLength   int     // Embedding dimension
	FeedForwardLength int     // Feed forward network size
	BlockSize         int     // Block/layer count
	HeadCount         int     // Number of attention heads
	HeadCountKV       int     // Number of key-value heads (for grouped query attention)
	LayerNormRMS_EPS  float64 // Layer normalization epsilon

	// Tokenizer information
	TokenCount     int    // Number of tokens in vocabulary
	TokenizerModel string // Tokenizer model type (e.g., "llama", "gpt2")
	BosTokenID     int    // Beginning of sequence token ID
	EosTokenID     int    // End of sequence token ID
	PadTokenID     int    // Padding token ID
	UncTokenID     int    // Unknown token ID

	// Rope (Rotary Position Embedding) settings
	RopeDim       int     // Rotary dimension
	RopeFreqBase  float64 // Frequency base for rotary embeddings
	RopeFreqScale float64 // Frequency scaling factor

	// Quantization settings
	QuantizationVersion uint32 // GGUF quantization version
	FileType           uint32 // GGUF file type (Q4_0, Q4_K_M, etc.)

	// Special tokens
	PreToken  string // String to prepend to tokens
	PostToken string // String to append to tokens

	// Additional metadata as key-value pairs
	Extra map[string]interface{}
}

// FileTypeMap maps GGUF file type codes to their string names
var FileTypeMap = map[uint32]string{
	0:  "ALL_F32",
	1:  "MOSTLY_F16",
	2:  "MOSTLY_Q4_0",
	3:  "MOSTLY_Q4_1",
	4:  "MOSTLY_Q4_2", // Unsupported
	5:  "MOSTLY_Q4_3", // Unsupported
	6:  "MOSTLY_Q5_0",
	7:  "MOSTLY_Q5_1",
	8:  "MOSTLY_Q8_0",
	9:  "MOSTLY_Q8_1",
	10: "MOSTLY_Q2_K", // Unsupported
	11: "MOSTLY_Q3_K",
	12: "MOSTLY_Q4_K",
	13: "MOSTLY_Q5_K",
	14: "MOSTLY_Q6_K",
	15: "MOSTLY_Q2_K_S",
	16: "MOSTLY_Q3_K_S",
	17: "MOSTLY_Q2_K_XL",
	18: "MOSTLY_Q3_K_XL",
	19: "MOSTLY_Q1_K", // Unsupported
	20: "MOSTLY_Q4_K_S",
	21: "MOSTLY_Q3_K_S_XL",
	22: "MOSTLY_Q2_K_XL",
}

// GetQuantizationString returns the human-readable quantization string
func (m *Metadata) GetQuantizationString() string {
	if m.FileType == 0 {
		return "F32"
	}
	if name, ok := FileTypeMap[m.FileType]; ok {
		// Convert to short format like "Q4_K_M"
		switch name {
		case "MOSTLY_Q4_K":
			return "Q4_K_M"
		case "MOSTLY_Q5_K":
			return "Q5_K_M"
		case "MOSTLY_Q3_K":
			return "Q3_K_M"
		case "MOSTLY_Q4_K_S":
			return "Q4_K_S"
		default:
			// Remove "MOSTLY_" prefix
			if len(name) > 7 {
				return name[7:]
			}
			return name
		}
	}
	return "UNKNOWN"
}

// GetParametersInBillions returns the parameter count in billions (B)
func (m *Metadata) GetParametersInBillions() float64 {
	if m.Parameters > 0 {
		return m.Parameters
	}
	// Estimate from layer count and embedding size if not directly specified
	if m.BlockSize > 0 && m.EmbeddingLength > 0 {
		// Rough estimation: 12 * n_layers * d_model^2 (for standard transformer)
		params := 12.0 * float64(m.BlockSize) * float64(m.EmbeddingLength*m.EmbeddingLength) / 1e9
		return params
	}
	return 0
}

// IsChatModel returns true if this appears to be a chat/instruct model
// Heuristic: checks model name for common chat model suffixes
func (m *Metadata) IsChatModel() bool {
	name := m.Name
	chatSuffixes := []string{
		"-chat", "-instruct", "-sft", "-lora", "-adapter",
		"chat", "instruct", "sft", "conversation", "dialogue",
	}
	for _, suffix := range chatSuffixes {
		if contains(name, suffix) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	// Simple case-insensitive contains
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase (simplified, ASCII only)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
