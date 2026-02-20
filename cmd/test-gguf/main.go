package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shepherd-project/shepherd/Shepherd/internal/gguf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: test-gguf <gguf文件路径> [gguf文件路径2...]")
		os.Exit(1)
	}

	for i, path := range os.Args[1:] {
		if i > 0 {
			fmt.Println("\n" + string(make([]byte, 60)))
		}

		fmt.Printf("\n========== 测试文件 %d: %s ==========\n", i+1, filepath.Base(path))

		// 检查文件是否存在
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("❌ 文件不存在: %s\n", path)
			continue
		}

		// 创建解析器
		parser, err := gguf.NewParser(path)
		if err != nil {
			fmt.Printf("❌ 创建解析器失败: %v\n", err)
			continue
		}
		defer parser.Close()

		// 获取元数据
		meta, err := parser.GetMetadata()
		if err != nil {
			fmt.Printf("❌ 获取元数据失败: %v\n", err)
			continue
		}

		// 显示结果
		fmt.Printf("\n📄 基本信息:\n")
		fmt.Printf("  文件名: %s\n", filepath.Base(path))
		fmt.Printf("  文件大小: %s\n", formatBytes(parser.GetFileSize()))

		fmt.Printf("\n🏗️  模型架构:\n")
		fmt.Printf("  名称: %s\n", nonEmpty(meta.Name, "(未设置)"))
		fmt.Printf("  架构: %s\n", nonEmpty(meta.Architecture, "(未设置)"))
		fmt.Printf("  文件类型: %d\n", meta.FileType)
		fmt.Printf("  量化类型: %s\n", nonEmpty(meta.Quantization, "(未知)"))
		fmt.Printf("  参数量: %.0f (%s)\n", meta.Parameters, formatParameters(meta.Parameters))

		fmt.Printf("\n📐 模型参数:\n")
		fmt.Printf("  上下文长度: %d\n", meta.ContextLength)
		fmt.Printf("  嵌入维度: %d\n", meta.EmbeddingLength)
		fmt.Printf("  块数量: %d\n", meta.BlockSize)
		fmt.Printf("  前馈维度: %d\n", meta.FeedForwardLength)
		fmt.Printf("  注意力头数: %d\n", meta.HeadCount)
		fmt.Printf("  KV 注意力头数: %d\n", meta.HeadCountKV)
		fmt.Printf("  RoPE 维度: %d\n", meta.RopeDim)
		fmt.Printf("  RoPE 频率基数: %.2f\n", meta.RopeFreqBase)
		fmt.Printf("  RoPE 频率缩放: %.4f\n", meta.RopeFreqScale)

		fmt.Printf("\n🔤 Tokenizer:\n")
		fmt.Printf("  模型: %s\n", nonEmpty(meta.TokenizerModel, "(未设置)"))
		fmt.Printf("  BOS Token ID: %d\n", meta.BosTokenID)
		fmt.Printf("  EOS Token ID: %d\n", meta.EosTokenID)
		fmt.Printf("  PAD Token ID: %d\n", meta.PadTokenID)
		fmt.Printf("  UNK Token ID: %d\n", meta.UncTokenID)
		fmt.Printf("  Pre Token: %s\n", nonEmpty(meta.PreToken, "(未设置)"))
		fmt.Printf("  Post Token: %s\n", nonEmpty(meta.PostToken, "(未设置)"))

		// 验证结果
		fmt.Printf("\n✅ 验证结果:\n")
		allOK := true

		if meta.Architecture == "" {
			fmt.Printf("  ⚠️  架构为空\n")
			allOK = false
		}
		if meta.FileType == 0 && meta.Quantization == "F32" {
			// 某些模型可能真的是 F32，不一定是错误
			fmt.Printf("  ℹ️  文件类型为 0 (F32)\n")
		}
		if meta.Parameters == 0 {
			fmt.Printf("  ⚠️  参数量为 0\n")
			allOK = false
		}

		if allOK {
			fmt.Printf("  ✅ 所有关键字段都已正确填充\n")
		}
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("\n测试完成!")
}

func nonEmpty(s string, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatParameters(params float64) string {
	if params >= 1e9 {
		return fmt.Sprintf("%.1fB", params/1e9)
	}
	if params >= 1e6 {
		return fmt.Sprintf("%.1fM", params/1e6)
	}
	if params >= 1e3 {
		return fmt.Sprintf("%.1fK", params/1e3)
	}
	return fmt.Sprintf("%.0f", params)
}
