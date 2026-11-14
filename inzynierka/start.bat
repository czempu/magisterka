@echo off
cd /d "%~dp0"
title ERPBadger (Inzynierka) - System RCP
echo ==============================================================
echo    URUCHAMIANIE SYSTEMU ERPBADGER (INZYNIERKA - DOCKER)
echo ==============================================================

:: 1. Upewnij sie, ze Docker CLI jest w sciezce PATH
set "DOCKER_BIN=C:\Users\czemp\AppData\Local\Programs\DockerDesktop\resources\bin"
if exist "%DOCKER_BIN%" (
    set "PATH=%DOCKER_BIN%;%PATH%"
)

:: 2. Sprawdz czy Docker demon dziala, jesli nie - uruchom Docker Desktop
docker ps >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [INFO] Uruchamianie silnika Docker Desktop...
    start "" "C:\Users\czemp\AppData\Local\Programs\DockerDesktop\Docker Desktop.exe"
    echo [INFO] Oczekiwanie na gotowosc demona Dockera...
    :wait_docker
    timeout /t 3 /nobreak >nul
    docker ps >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo [INFO] Czekam na start Dockera...
        goto wait_docker
    )
)

echo [OK] Silnik Docker jest aktywny.
echo.
echo ==============================================================
echo Budowanie i uruchamianie kontenerow bazy MySQL i backendu Go...
echo ==============================================================
docker compose up --build -d

if %ERRORLEVEL% NEQ 0 (
    echo [BLAD] Nie udalo sie uruchomic kontenerow.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ==============================================================
echo   [OK] System ERPBadger dziala pomyslnie!
echo.
echo   * Adres logowania: http://localhost:8080/login
echo   * Administrator:   login: admin   haslo: admin123
echo   * Pracownik:       login: user    haslo: user123
echo   * Baza MySQL:      port 3306 (RCP, kalendarz)
echo.
echo   Aby zatrzymac kontenery: docker compose down
echo ==============================================================
echo.
start http://localhost:8080/login
pause
