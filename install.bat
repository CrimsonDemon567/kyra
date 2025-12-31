@echo off
set VERSION=1.0.0
set INSTALLDIR=%ProgramFiles%\Kyra

echo Installing Kyra SDK %VERSION%...

if not exist "%INSTALLDIR%" (
    mkdir "%INSTALLDIR%"
)

copy /Y kyra-%VERSION%-windows-amd64.exe "%INSTALLDIR%\kyra.exe"
copy /Y kyrac-%VERSION%-windows-amd64.exe "%INSTALLDIR%\kyrac.exe"

REM Add to PATH if missing
echo %PATH% | find /I "%INSTALLDIR%" >nul
if errorlevel 1 (
    echo Adding Kyra to PATH...
    setx PATH "%PATH%;%INSTALLDIR%"
)

echo Kyra SDK installed successfully.
echo You can now run:
echo   kyra -kbc file.kbc
echo   kyrac -kbc file.kyra
