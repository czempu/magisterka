@echo off
cd /d "%~dp0"
title RCP Scheduler - Backend (Python) & Frontend (Go)
echo ==============================================================
echo Uruchamianie calego projektu (Backend FastAPI + Frontend Go)...
echo ==============================================================
call .venv\Scripts\python.exe run_all.py
pause
