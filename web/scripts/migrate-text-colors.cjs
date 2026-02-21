#!/usr/bin/env node
/**
 * 文本颜色语义化迁移脚本
 *
 * 将硬编码的 text-gray-* dark:text-* 模式替换为语义化的 CSS 变量类
 */

const fs = require('fs');
const path = require('path');

// 定义颜色映射规则
const TEXT_MAPPINGS = [
  // 主要文本 → text-foreground
  { from: 'text-gray-300 dark:text-white', to: 'text-foreground' },
  { from: 'text-gray-300 dark:text-gray-100', to: 'text-foreground' },
  { from: 'text-gray-700 dark:text-gray-200', to: 'text-foreground' },
  { from: 'text-gray-700 dark:text-white', to: 'text-foreground' },
  { from: 'text-gray-900 dark:text-white', to: 'text-foreground' },
  { from: 'text-gray-900 dark:text-gray-100', to: 'text-foreground' },
  { from: 'text-gray-800 dark:text-gray-200', to: 'text-foreground' },

  // 次要文本/提示 → text-muted-foreground
  { from: 'text-gray-400 dark:text-gray-300', to: 'text-muted-foreground' },
  { from: 'text-gray-500 dark:text-gray-400', to: 'text-muted-foreground' },
  { from: 'text-gray-500 dark:text-gray-300', to: 'text-muted-foreground' },
  { from: 'text-gray-600 dark:text-gray-300', to: 'text-muted-foreground' },
  { from: 'text-gray-600 dark:text-gray-400', to: 'text-muted-foreground' },
  { from: 'text-gray-600 dark:text-gray-200', to: 'text-muted-foreground' },
  { from: 'text-gray-400 dark:text-gray-200', to: 'text-muted-foreground' },

  // 卡片内文本 → text-card-foreground
  { from: 'text-gray-100 dark:text-gray-100', to: 'text-card-foreground' },

  // 错误/破坏性状态
  { from: 'text-red-700 dark:text-red-400', to: 'text-destructive' },
];

// 需要保留的功能性颜色（不替换）
const PRESERVE_PATTERNS = [
  /text-green-\d+ dark:text-green-/,
  /text-blue-\d+ dark:text-blue-/,
  /text-yellow-\d+ dark:text-yellow-/,
  /text-red-\d+ dark:text-red-/,
  /text-purple-\d+ dark:text-purple-/,
  /text-primary/,
  /text-accent/,
];

// 统计信息
const stats = {
  filesProcessed: 0,
  filesModified: 0,
  replacementsMade: 0,
  skippedFiles: [],
};

/**
 * 检查字符串是否应保留
 */
function shouldPreserve(text) {
  // 检查是否包含需要保留的功能性颜色
  for (const pattern of PRESERVE_PATTERNS) {
    if (pattern.test(text)) {
      return true;
    }
  }
  return false;
}

/**
 * 处理单个文件
 */
function processFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, 'utf8');
    let modified = false;
    let replacements = 0;

    // 应用所有映射规则
    for (const mapping of TEXT_MAPPINGS) {
      const regex = new RegExp(mapping.from.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g');
      const matches = content.match(regex);

      if (matches) {
        // 检查是否应该保留
        if (!shouldPreserve(content.substring(content.indexOf(matches[0]) - 50, content.indexOf(matches[0]) + 50))) {
          content = content.replace(regex, mapping.to);
          modified = true;
          replacements += matches.length;
          console.log(`  ✓ ${mapping.from} → ${mapping.to} (${matches.length} 次)`);
        }
      }
    }

    if (modified) {
      fs.writeFileSync(filePath, content, 'utf8');
      stats.filesModified++;
      stats.replacementsMade += replacements;
      return true;
    }

    return false;
  } catch (error) {
    console.error(`  ✗ 处理文件失败: ${error.message}`);
    stats.skippedFiles.push({ file: filePath, error: error.message });
    return false;
  }
}

/**
 * 递归遍历目录
 */
function processDirectory(dir, extensions = ['.tsx', '.ts', '.jsx', '.js']) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      // 跳过 node_modules 和其他不需要的目录
      if (!['node_modules', '.git', 'dist', 'build'].includes(entry.name)) {
        processDirectory(fullPath, extensions);
      }
    } else if (entry.isFile()) {
      const ext = path.extname(entry.name);
      if (extensions.includes(ext)) {
        stats.filesProcessed++;
        console.log(`处理: ${fullPath.replace(process.cwd(), '.')}`);
        processFile(fullPath);
      }
    }
  }
}

// 主函数
function main() {
  console.log('='.repeat(60));
  console.log('文本颜色语义化迁移脚本');
  console.log('='.repeat(60));
  console.log('');

  const srcDir = path.join(__dirname, '../src');

  if (!fs.existsSync(srcDir)) {
    console.error(`错误: 源目录不存在: ${srcDir}`);
    process.exit(1);
  }

  console.log(`开始处理: ${srcDir}`);
  console.log('');

  processDirectory(srcDir);

  console.log('');
  console.log('='.repeat(60));
  console.log('迁移完成统计');
  console.log('='.repeat(60));
  console.log(`处理文件数: ${stats.filesProcessed}`);
  console.log(`修改文件数: ${stats.filesModified}`);
  console.log(`替换次数: ${stats.replacementsMade}`);
  console.log('');

  if (stats.skippedFiles.length > 0) {
    console.log('跳过的文件:');
    stats.skippedFiles.forEach(({ file, error }) => {
      console.log(`  - ${file}: ${error}`);
    });
    console.log('');
  }

  console.log('✓ 迁移完成！');
  console.log('');
  console.log('后续步骤:');
  console.log('1. 运行 npm run dev 启动开发服务器');
  console.log('2. 在浏览器中测试浅色和深色主题');
  console.log('3. 检查所有页面的文本颜色是否正确');
}

if (require.main === module) {
  main();
}

module.exports = { TEXT_MAPPINGS, processFile, processDirectory };
