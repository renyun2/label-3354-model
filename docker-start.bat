@echo off
echo ==============================================
echo Go API Starter Docker 启动脚本
echo ==============================================

echo.
echo [1/4] 检查 Docker 是否运行...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker 未运行，请先启动 Docker Desktop!
    pause
    exit /b 1
)

echo [2/4] 停止并清理旧容器（如果存在）...
docker-compose down -v

echo [3/4] 构建并启动所有服务...
docker-compose up -d --build

echo [4/4] 等待服务启动...
timeout /t 10 /nobreak >nul

echo.
echo ==============================================
echo 服务启动完成!
echo ==============================================
echo.
echo API 地址: http://localhost:8080
echo Swagger 文档: http://localhost:8080/swagger/index.html
echo.
echo MySQL 数据库:
echo   - 主机: localhost
echo   - 端口: 3306
echo   - 用户名: root
echo   - 密码: root123456
echo   - 数据库: go_api_starter
echo.
echo Redis:
echo   - 主机: localhost
echo   - 端口: 6379
echo.
echo 查看日志: docker-compose logs -f
echo 停止服务: docker-compose down
echo 停止并清除数据: docker-compose down -v
echo.
pause
