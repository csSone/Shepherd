@echo off
REM Shepherd 运行脚本 (Windows)
REM 支持 standalone, master, client 三种模式

setlocal enabledelayedexpansion

REM 获取脚本所在目录
set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "BUILD_DIR=%PROJECT_DIR%\build"
set "BINARY_NAME=shepherd.exe"

REM 颜色设置 (Windows 10+)
set "INFO=[INFO]"
set "SUCCESS=[SUCCESS]"
set "WARNING=[WARNING]"
set "ERROR=[ERROR]"

REM 显示帮助信息
:show_help
echo 🐏 Shepherd 运行脚本
echo.
echo 用法: %~nx0 [模式] [选项]
echo.
echo 模式:
echo     standalone     单机模式 (默认)
echo     master         Master 模式 - 管理多个 Client 节点
echo     client         Client 模式 - 作为工作节点
echo.
echo 通用选项:
echo     -h, --help     显示此帮助信息
echo     -b, --build    运行前先编译
echo     -v, --version  显示版本信息
echo.
echo Master 模式选项:
echo     --port PORT    Web 服务器端口 (默认: 9190)
echo     --scan         启动时自动扫描网络
echo.
echo Client 模式选项:
echo     --master URL   Master 地址 (必需)
echo     --name NAME    Client 名称 (可选)
echo     --tags TAGS    Client 标签，逗号分隔 (可选)
echo.
echo 示例:
echo     # 单机模式
echo     %~nx0 standalone
echo.
echo     # Master 模式
echo     %~nx0 master --port 9190 --scan
echo.
echo     # Client 模式
echo     %~nx0 client --master http://192.168.1.100:9190 --name client-1
echo.
echo     # 运行前先编译
echo     %~nx0 master -b
echo.
goto :eof

REM 检查二进制文件是否存在
:check_binary
if not exist "%BUILD_DIR%\%BINARY_NAME%" (
    echo %WARNING% 二进制文件不存在: %BUILD_DIR%\%BINARY_NAME%
    set /p BUILD_NOW="是否现在编译? (y/N): "
    if /i "!BUILD_NOW!"=="y" (
        cd /d "%SCRIPT_DIR%"
        call build.bat
        cd /d "%PROJECT_DIR%"
    ) else (
        echo %ERROR% 无法继续，请先编译项目
        exit /b 1
    )
)
goto :eof

REM 显示版本信息
:show_version
if exist "%BUILD_DIR%\%BINARY_NAME%" (
    "%BUILD_DIR%\%BINARY_NAME%" --version
) else (
    echo %ERROR% 二进制文件不存在，请先编译
    exit /b 1
)
exit /b 0

REM 主函数
:main
set "MODE="
set "BUILD_FIRST=0"
set "MASTER_ADDR="
set "CLIENT_NAME="
set "CLIENT_TAGS="
set "WEB_PORT=9190"
set "AUTO_SCAN=0"

REM 解析参数
:parse_args
if "%~1"=="" goto :args_done
if /i "%~1"=="standalone" (
    set "MODE=standalone"
    shift
    goto :parse_args
)
if /i "%~1"=="master" (
    set "MODE=master"
    shift
    goto :parse_args
)
if /i "%~1"=="client" (
    set "MODE=client"
    shift
    goto :parse_args
)
if /i "%~1"=="-h" goto :show_help
if /i "%~1"=="--help" goto :show_help
if /i "%~1"=="-b" (
    set "BUILD_FIRST=1"
    shift
    goto :parse_args
)
if /i "%~1"=="-v" goto :show_version
if /i "%~1"=="--version" goto :show_version
if /i "%~1"=="--master" (
    set "MASTER_ADDR=%~2"
    shift /2
    goto :parse_args
)
if /i "%~1"=="--name" (
    set "CLIENT_NAME=%~2"
    shift /2
    goto :parse_args
)
if /i "%~1"=="--tags" (
    set "CLIENT_TAGS=%~2"
    shift /2
    goto :parse_args
)
if /i "%~1"=="--port" (
    set "WEB_PORT=%~2"
    shift /2
    goto :parse_args
)
if /i "%~1"=="--scan" (
    set "AUTO_SCAN=1"
    shift
    goto :parse_args
)
echo %ERROR% 未知参数: %~1
goto :show_help

:args_done

REM 默认模式
if "%MODE%"=="" set "MODE=standalone"

REM 编译（如果需要）
if "%BUILD_FIRST%"=="1" (
    echo %INFO% 编译项目...
    cd /d "%SCRIPT_DIR%"
    call build.bat
    cd /d "%PROJECT_DIR%"
    echo %SUCCESS% 编译完成
)

REM 检查二进制文件
call :check_binary

REM 构建命令参数
set "ARGS=--mode=%MODE%"

if /i "%MODE%"=="master" (
    echo %INFO% 启动 Master 模式...
    set "ARGS=!ARGS! --master-addr=0.0.0.0:%WEB_PORT%"

    if "!AUTO_SCAN!"=="1" (
        echo %INFO% 启用自动网络扫描
    )
)

if /i "%MODE%"=="client" (
    if "%MASTER_ADDR%"=="" (
        echo %ERROR% Client 模式需要指定 Master 地址 (--master)
        echo %INFO% 示例: %~nx0 client --master http://192.168.1.100:9190
        exit /b 1
    )
    echo %INFO% 启动 Client 模式...
    echo %INFO% Master 地址: %MASTER_ADDR%
    set "ARGS=!ARGS! --master-address=%MASTER_ADDR%"

    if not "%CLIENT_NAME%"=="" (
        echo %INFO% Client 名称: %CLIENT_NAME%
        set "SHEPHERD_CLIENT_NAME=%CLIENT_NAME%"
    )

    if not "%CLIENT_TAGS%"=="" (
        echo %INFO% Client 标签: %CLIENT_TAGS%
        set "SHEPHERD_CLIENT_TAGS=%CLIENT_TAGS%"
    )
)

if /i "%MODE%"=="standalone" (
    echo %INFO% 启动单机模式...
)

REM 显示启动信息
echo.
echo ==========================================
echo   🐏 Shepherd %MODE%
echo ==========================================
echo   模式: %MODE%
if /i "%MODE%"=="master" (
    echo   端口: %WEB_PORT%
)
if /i "%MODE%"=="client" (
    echo   Master: %MASTER_ADDR%
)
echo ==========================================
echo.

REM 启动程序
cd /d "%PROJECT_DIR%"
"%BUILD_DIR%\%BINARY_NAME%" %ARGS%

goto :eof

REM 运行主函数
call :main %*
