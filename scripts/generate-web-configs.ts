/**
 * Shepherd Web 配置生成器
 *
 * 从 config/web/config.yaml 生成所有前端配置文件
 * 运行: tsx scripts/generate-web-configs.ts
 */

import fs from 'fs';
import path from 'path';
import yaml from 'js-yaml';

// 配置文件路径（从脚本位置计算，而非 cwd）
const SCRIPT_DIR = path.dirname(new URL(import.meta.url).pathname);
const PROJECT_ROOT = path.dirname(SCRIPT_DIR);
const CONFIG_DIR = path.join(PROJECT_ROOT, 'config', 'web');
const WEB_DIR = path.join(PROJECT_ROOT, 'web');
const YAML_FILE = path.join(CONFIG_DIR, 'config.yaml');

// 主函数
async function main() {
  console.log('🔧 开始生成 Web 前端配置文件...\n');

  // 读取 YAML 配置
  const yamlContent = fs.readFileSync(YAML_FILE, 'utf8');
  const config = yaml.load(yamlContent) as any;

  console.log(`📄 读取配置文件: ${YAML_FILE}`);

  // 生成 TypeScript 配置
  generateTsConfig(config.typescript);

  // 生成 Vite 配置
  generateViteConfig(config.vite);

  // 生成 Tailwind 配置
  generateTailwindConfig(config.tailwind);

  // 生成 PostCSS 配置
  generatePostCSSConfig(config.postcss);

  // 生成 ESLint 配置
  generateEslintConfig(config.eslint);

  console.log('\n✅ 所有配置文件生成完成！');
}

/**
 * 生成 TypeScript 配置文件
 */
function generateTsConfig(tsConfig: any) {
  console.log('\n📝 生成 TypeScript 配置...');

  // tsconfig.app.json
  const tsConfigApp = {
    compilerOptions: {
      ...tsConfig.baseCompilerOptions,
      ...tsConfig.app.compilerOptions,
    },
    include: tsConfig.app.include,
  };
  writeFile(
    path.join(WEB_DIR, 'tsconfig.app.json'),
    JSON.stringify(tsConfigApp, null, 2)
  );
  console.log('  ✓ tsconfig.app.json');

  // tsconfig.node.json
  const tsConfigNode = {
    compilerOptions: {
      ...tsConfig.baseCompilerOptions,
      ...tsConfig.node.compilerOptions,
    },
    include: tsConfig.node.include,
  };
  writeFile(
    path.join(WEB_DIR, 'tsconfig.node.json'),
    JSON.stringify(tsConfigNode, null, 2)
  );
  console.log('  ✓ tsconfig.node.json');

  // tsconfig.json (根配置)
  const tsConfigRoot = {
    files: [],
    references: [
      { path: './tsconfig.app.json' },
      { path: './tsconfig.node.json' },
    ],
  };
  writeFile(
    path.join(WEB_DIR, 'tsconfig.json'),
    JSON.stringify(tsConfigRoot, null, 2)
  );
  console.log('  ✓ tsconfig.json');
}

/**
 * 生成 Vite 配置文件
 */
function generateViteConfig(viteConfig: any) {
  console.log('\n📝 生成 Vite 配置...');

  const content = `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '${viteConfig.resolve.alias['@']}'),
    },
  },
  server: ${JSON.stringify(viteConfig.server, null, 2)},
  build: ${JSON.stringify(viteConfig.build, null, 2)},
});
`;

  writeFile(path.join(WEB_DIR, 'vite.config.ts'), content);
  console.log('  ✓ vite.config.ts');
}

/**
 * 生成 Tailwind 配置文件
 */
function generateTailwindConfig(tailwindConfig: any) {
  console.log('\n📝 生成 Tailwind 配置...');

  const content = `/** @type {import('tailwindcss').Config} */
export default ${JSON.stringify(tailwindConfig, null, 2)}
`;

  writeFile(path.join(WEB_DIR, 'tailwind.config.js'), content);
  console.log('  ✓ tailwind.config.js');
}

/**
 * 生成 PostCSS 配置文件
 */
function generatePostCSSConfig(postcssConfig: any) {
  console.log('\n📝 生成 PostCSS 配置...');

  const content = `export default ${JSON.stringify(postcssConfig, null, 2)}
`;

  writeFile(path.join(WEB_DIR, 'postcss.config.js'), content);
  console.log('  ✓ postcss.config.js');
}

/**
 * 生成 ESLint 配置文件
 */
function generateEslintConfig(eslintConfig: any) {
  console.log('\n📝 生成 ESLint 配置...');

  const content = `import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(${JSON.stringify(eslintConfig.ignores)}),
  {
    files: ${JSON.stringify(eslintConfig.files)},
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: ${JSON.stringify(eslintConfig.languageOptions, null, 4)},
  },
])
`;

  writeFile(path.join(WEB_DIR, 'eslint.config.js'), content);
  console.log('  ✓ eslint.config.js');
}

/**
 * 写入文件（带错误处理）
 */
function writeFile(filePath: string, content: string) {
  try {
    fs.writeFileSync(filePath, content, 'utf8');
  } catch (error) {
    console.error(`❌ 写入文件失败: ${filePath}`);
    console.error(error);
    process.exit(1);
  }
}

// 运行主函数
main().catch((error) => {
  console.error('❌ 配置生成失败:', error);
  process.exit(1);
});
