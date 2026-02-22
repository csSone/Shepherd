package config

import (
	"fmt"
	"testing"
)

// TestDefaultConfigInTestEnv 验证测试环境中 DefaultConfig() 返回空路径
func TestDefaultConfigInTestEnv(t *testing.T) {
	if !testing.Testing() {
		t.Fatal("此测试必须在测试环境中运行")
	}

	cfg := DefaultConfig()

	// 验证路径为空
	if len(cfg.Model.Paths) != 0 {
		t.Errorf("测试环境中 Model.Paths 应该为空,但得到 %d 个路径: %v",
			len(cfg.Model.Paths), cfg.Model.Paths)
	}

	// 验证自动扫描被禁用
	if cfg.Model.AutoScan {
		t.Error("测试环境中 Model.AutoScan 应该为 false")
	}

	fmt.Printf("✅ 测试环境检测正确:\n")
	fmt.Printf("   - Model.Paths = %v\n", cfg.Model.Paths)
	fmt.Printf("   - Model.AutoScan = %v\n", cfg.Model.AutoScan)
}
