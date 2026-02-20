// Package modelrepo provides model repository integration for HuggingFace and ModelScope
package modelrepo

import (
	"testing"
)

// TestNewClient tests creating a new model repository client
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.httpClient == nil {
		t.Error("client.httpClient is nil")
	}
}

// TestSetHFToken tests setting HuggingFace token
func TestSetHFToken(t *testing.T) {
	client := NewClient()
	testToken := "test_token_123"

	client.SetHFToken(testToken)
	if client.hfToken != testToken {
		t.Errorf("SetHFToken() = %v, want %v", client.hfToken, testToken)
	}
}

// TestGenerateHuggingFaceURL tests HuggingFace URL generation
func TestGenerateHuggingFaceURL(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		repoID   string
		fileName string
		want     string
	}{
		{
			name:     "with filename",
			repoID:   "Qwen/Qwen2-7B-Instruct",
			fileName: "model.gguf",
			want:     "https://huggingface.co/Qwen/Qwen2-7B-Instruct/resolve/main/model.gguf",
		},
		{
			name:     "without filename",
			repoID:   "meta-llama/Llama-2-7b",
			fileName: "",
			want:     "https://huggingface.co/meta-llama/Llama-2-7b/resolve/main/",
		},
		{
			name:     "nested filename",
			repoID:   "owner/model",
			fileName: "path/to/model.gguf",
			want:     "https://huggingface.co/owner/model/resolve/main/path/to/model.gguf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GenerateDownloadURL(SourceHuggingFace, tt.repoID, tt.fileName)
			if err != nil {
				t.Errorf("GenerateDownloadURL() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateModelScopeURL tests ModelScope URL generation
func TestGenerateModelScopeURL(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		repoID   string
		fileName string
		want     string
	}{
		{
			name:     "with filename",
			repoID:   "AI-ModelScope/qwen-7b",
			fileName: "model.gguf",
			want:     "https://www.modelscope.cn/api/v1/models/AI-ModelScope/qwen-7b/repo?Revision=master&FilePath=model.gguf",
		},
		{
			name:     "without filename",
			repoID:   "owner/model",
			fileName: "",
			want:     "https://www.modelscope.cn/api/v1/models/owner/model/repo?Revision=master&FilePath=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GenerateDownloadURL(SourceModelScope, tt.repoID, tt.fileName)
			if err != nil {
				t.Errorf("GenerateDownloadURL() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateDownloadURLUnsupported tests unsupported source
func TestGenerateDownloadURLUnsupported(t *testing.T) {
	client := NewClient()

	_, err := client.GenerateDownloadURL("unsupported", "repo", "file")
	if err == nil {
		t.Error("expected error for unsupported source, got nil")
	}
}

// TestParseRepoID tests repository ID parsing
func TestParseRepoID(t *testing.T) {
	tests := []struct {
		name      string
		repoID    string
		wantOwner string
		wantModel string
		wantErr   bool
	}{
		{
			name:      "valid repo ID",
			repoID:    "Qwen/Qwen2-7B-Instruct",
			wantOwner: "Qwen",
			wantModel: "Qwen2-7B-Instruct",
			wantErr:   false,
		},
		{
			name:      "single segment",
			repoID:    "Qwen2",
			wantOwner: "",
			wantModel: "",
			wantErr:   true,
		},
		{
			name:      "three segments",
			repoID:    "org/user/model",
			wantOwner: "",
			wantModel: "",
			wantErr:   true,
		},
		{
			name:      "empty string",
			repoID:    "",
			wantOwner: "",
			wantModel: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, model, err := ParseRepoID(tt.repoID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseRepoID() owner = %v, want %v", owner, tt.wantOwner)
			}
			if model != tt.wantModel {
				t.Errorf("ParseRepoID() model = %v, want %v", model, tt.wantModel)
			}
		})
	}
}

// TestIsGGUFFile tests GGUF file detection
func TestIsGGUFFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"model.gguf", true},
		{"model.GGUF", true},
		{"path/to/model.gguf", true},
		{"model.safetensors", false},
		{"config.json", false},
		{"README.md", false},
		{"model.gguf.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isGGUFFile(tt.path); got != tt.want {
				t.Errorf("isGGUFFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFileInfo tests FileInfo structure
func TestFileInfo(t *testing.T) {
	file := FileInfo{
		Name:        "model.gguf",
		Size:        1024000,
		DownloadURL: "https://example.com/model.gguf",
	}

	if file.Name != "model.gguf" {
		t.Errorf("Name = %v, want model.gguf", file.Name)
	}
	if file.Size != 1024000 {
		t.Errorf("Size = %v, want 1024000", file.Size)
	}
	if file.DownloadURL != "https://example.com/model.gguf" {
		t.Errorf("DownloadURL = %v, want https://example.com/model.gguf", file.DownloadURL)
	}
}

// TestSourceConstants tests source constants
func TestSourceConstants(t *testing.T) {
	if SourceHuggingFace != "huggingface" {
		t.Errorf("SourceHuggingFace = %v, want huggingface", SourceHuggingFace)
	}
	if SourceModelScope != "modelscope" {
		t.Errorf("SourceModelScope = %v, want modelscope", SourceModelScope)
	}
}
