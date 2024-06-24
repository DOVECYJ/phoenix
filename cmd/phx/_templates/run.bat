@echo off

set server=_build\platform.exe
:: read server pid
for /F %%i in ('type pid') do (set pid=%%i)
:: check process status
for /F "skip=3 tokens=2" %%i in ('tasklist /fi "PID eq %pid%"') do (
    if %%i==%pid% set isrunning=yes
)

if "%1" == "start" (
    if "%isrunning%" == "yes" (
        echo server is running
    ) else (
        echo starting server...
        goto do-starting
    )
    exit /b 0
) ^
else if "%1" == "restart" (
    if "%isrunning%" == "yes" (
        echo stopping server...
        taskkill /f /pid %pid%
    )
    echo restarting server...
    goto do-starting
    exit /b 0
) ^
else if "%1" == "stop" (
    if "%isrunning%" == "yes" (
        echo stopping server...
        taskkill /f /pid %pid%
    ) else (
        echo server isnot running
    )
    exit /b 0
) ^
else if "%1" == "status" (
    echo server status...
    if "%isrunning%" == "yes" (echo running) else (echo gone)
    exit /b 0
) ^
else (
    echo "USAGE: run.bat [start|restart|stop|status]"
    exit /b 0
)
exit

:do-starting
powershell.exe -command "& {Start-Process -WindowStyle Hidden -WorkingDirectory '%~dp0' -FilePath '%server%' -RedirectStandardOutput logs\stdout.log}"
timeout /T 1 /NOBREAK > nul
for /F %%i in ('type pid') do (set pid=%%i)
for /F "skip=3 tokens=2" %%i in ('tasklist /fi "PID eq %pid%"') do (
    if %%i==%pid% set isrunning=yes
)
if "%isrunning%" == "yes" (echo success) else (echo failed)
exit /b 0