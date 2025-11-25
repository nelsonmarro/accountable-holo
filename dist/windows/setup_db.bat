@echo off
echo Setting up Accountable Holo Database...
set /p PGPASSWORD=Enter the password you set for the 'postgres' user during installation: 
set PGUSER=postgres
REM Check if psql is in PATH, otherwise try default location
where psql >nul 2>nul
if %ERRORLEVEL% NEQ 0 set PATH=%PATH%;C:\Program Files\PostgreSQL\16\bin
createdb -w accountableholodb
psql -d accountableholodb -f schema.sql
echo Done! You can now run AccountableHolo.exe
pause
