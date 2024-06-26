@echo off
setlocal enabledelayedexpansion

:: Display the directory structure
echo Directory structure:
@REM tree /F
tree /F
echo.

echo ==========================================
echo Displaying contents of each file
echo ==========================================
echo.

:: Loop through each file in the directory and subdirectories
for /r %%f in (*) do (
    :: Skip files in the venv directory and exclude .txt, .bat, .jpg, .vec, .svm files
    echo %%f | findstr /I /C:"venv" >nul
    if errorlevel 1 (
        echo %%f | findstr /I /E /C:".txt" /C:".bat" /C:".jpg" /C:".vec" /C:".svm" >nul
        if errorlevel 1 (
            :: Check if the file has content
            if exist "%%f" (
                echo.
                echo File: %%f
                echo ---------------------
                :: Display the content of the file, excluding .jpg, .vec, .svm files
                if not "%%~xf"==".jpg" if not "%%~xf"==".vec" if not "%%~xf"==".svm" (
                    type "%%f"
                )
                echo.
                echo ==========================================
            )
        )
    )
)

endlocal
pause