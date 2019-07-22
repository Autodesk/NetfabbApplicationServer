@echo OFF
set Scriptpath=%~dp0
set PATH=%PATH%;%Scriptpath%

set APPLICATIONPORT=8650

echo Do you want to add or remove the firewall rules:
echo 1: Add
echo 2: Remove
set INPUT=
set /P INPUT=Type input # and enter: %=%

if /I %INPUT% == 1 set installOrUninstall=install
if /I %INPUT% == 2 set installOrUninstall=uninstall

net session >nul 2>&1
if %ERRORLEVEL% EQU 0 (
	goto installUninstallService
) else (
    echo This batch requires Administrator rights.
	echo Exiting...
	PING 127.0.0.1 > NUL 2>&1
	exit /B 1
)


:installUninstallService
if "%installOrUninstall%" == "install" (
  netsh advfirewall firewall show rule name="Netfabb Application Server TCP %APPLICATIONPORT%" | findstr "no rules"
  netsh advfirewall firewall show rule name="Netfabb Application Server UDP %APPLICATIONPORT%" | findstr "no rules"
  if %errorlevel% NEQ 0 (
	echo Firewall rules already existing.
	echo Done.
	PING 127.0.0.1 > NUL 2>&1
	echo Exiting...
	exit /b 1
  ) 
  netsh advfirewall firewall add rule name="Netfabb Application Server TCP %APPLICATIONPORT%" dir=in action=allow protocol=TCP localport=%APPLICATIONPORT%
  netsh advfirewall firewall add rule name="Netfabb Application Server UDP %APPLICATIONPORT%" dir=in action=allow protocol=UDP localport=%APPLICATIONPORT%
  echo Done adding firewall rules
  PING 127.0.0.1 > NUL 2>&1
  echo Exiting...
  exit /b 1
  )
) else (
  echo Removing Firewall rules...
  echo -------------------------------------
  netsh advfirewall firewall delete rule name="Netfabb Application Server TCP %APPLICATIONPORT%"
  netsh advfirewall firewall delete rule name="Netfabb Application Server UDP %APPLICATIONPORT%"
  echo Done.
  PING 127.0.0.1 > NUL 2>&1
  exit /b 1
)


