@echo off
REM Shepherd Web 前端运行脚本 (Windows)
REM 此脚本应从 scripts/ 目录运行，操作 web/ 目录

setlocal enabledelayedexpansion

REM 获取项目根目录
set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "WEB_DIR=%PROJECT_DIR%\web"

REM 显示帮助信息
:show_help
echo 🐏 Shepherd Web 前端
echo.
echo 用法: %~nx0 [命令] [选项]
echo.
echo 命令:
echo     dev         启动开发服务器 (默认)
echo     build       构建生产版本
echo     preview     预览生产构建
echo     install     安装依赖
echo     clean       清理构建文件
echo.
echo 选项:
echo     -h, --help     显示此帮助信息
echo     -p, --port PORT    指定端口 (开发模式默认: 3000)
echo.
echo 示例:
echo     %~nx0 dev                 # 启动开发服务器
echo     %~nx0 dev -p 4000         # 在端口 4000 启动
echo     %~nx0 build              # 构建生产版本
echo     %~nx0 preview            # 预览构建结果
echo.
goto :eof

REM 检查依赖
:check_dependencies
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Node.js 未安装
    exit /b 1
)

where npm >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] npm 未安装
    exit /b 1
)

if not exist "%WEB_DIR%\node_modules\" (
    echo [WARNING] 依赖未安装，正在安装...
    call :install_dependencies
)
goto :eof

REM 安装依赖
:install_dependencies
echo [INFO] 安装 Web 前端依赖...
cd /d "%WEB_DIR%"
call npm install
echo [SUCCESS] 依赖安装完成
goto :eof

REM 清理构建文件
:clean_build
echo [INFO] 清理 Web 构建文件...
cd /d "%WEB_DIR%"
if exist "dist\" rmdir /s /q dist
if exist "node_modules\.vite" rmdir /s /q node_modules\.vite
echo [SUCCESS] 清理完成
goto :eof

REM 启动开发服务器
:run_dev
set "PORT=%~1"
if "%PORT%"=="" set "PORT=3000"
echo [INFO] 启动开发服务器 (端口: %PORT%)...
cd /d "%WEB_DIR%"
call npm run dev -- --port %PORT%
goto :eof

REM 构建生产版本
:run_build
echo [INFO] 构建 Web 生产版本...
cd /d "%WEB_DIR%"
call npm run build
echo [SUCCESS] 构建完成，输出目录: web\dist\
goto :eof

REM 预览生产构建
:run_preview
echo [INFO] 预览 Web 生产构建...
cd /d "%WEB_DIR%"
call npm run preview
goto :eof

REM 主函数
:main
set "COMMAND="
set "PORT="

REM 解析参数
:parse_args
if "%~1"=="" goto :args_done
if /i "%~1"=="dev" (
    set "COMMAND=dev"
    shift
    goto :parse_args
)
if /i "%~1"=="build" (
    set "COMMAND=build"
    shift
    goto :parse_args
)
if /i "%~1"=="preview" (
    set "COMMAND=preview"
    shift
    goto :parse_args
)
if /i "%~1"=="install" (
    set "COMMAND=install"
    shift
    goto :parse_args
)
if /i "%~1"=="clean" (
    set "COMMAND=clean"
    shift
    goto :parse_args
)
if /i "%~1"=="-h" goto :show_help
if /i "%~1"=="--help" goto :show_help
if /i "%~1"=="-p" (
    set "PORT=%~2"
    shift /2
    goto :parse_args
)
echo [ERROR] 未知参数: %~1
goto :show_help

:args_done

REM 默认命令
if "%COMMAND%"=="" set "COMMAND=dev"

REM 检查依赖
call :check_dependencies

REM 执行命令
if /i "%COMMAND%"=="dev" (
    call :run_dev %PORT%
) else if /i "%COMMAND%"=="build" (
    call :run_build
) else if /i "%COMMAND%"=="preview" (
    call :run_preview
) else if /i "%COMMAND%"=="install" (
    call :install_dependencies
) else if /i "%COMMAND%"=="clean" (
    call :clean_build
)

goto :eof

call :main %*
