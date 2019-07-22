@ECHO OFF
:: Path Settings
set Scriptpath=%~dp0
set Binpath=%Scriptpath%Bin
set Docpath=%Scriptpath%doc
set PWSalterpath=%Scriptpath%PasswordSalter
set Examplepath=%Scriptpath%example
set Sourcepath=%Scriptpath%Source
set Outputpath=%Scriptpath%output
set ApplicationServer=%Binpath%\NetfabbApplicationServer.exe
set ApplicationService=%Binpath%\NetfabbApplicationService.exe
set TaskServer=%Binpath%\NetfabbTaskServer.exe
set ServiceBatch=%Binpath%\setup_firewall_rules.bat

:: GO Package Settings
set PackageGoSqlite=github.com/mattn/go-sqlite3
set PackageGoUuid=github.com/twinj/uuid

:: GO File Lists
set common_source=netfabbstorage_db.go netfabbstorage_types.go netfabbstorage_utils.go netfabbstorage_auth.go netfabbstorage_orm.go netfabbstorage_data.go netfabbtask_handler.go netfabbapplication.go
set taskserver_source=netfabbtaskserver.go netfabbtask_config.go 
set applicationserver_source=netfabbapplicationserver.go netfabbstorage_config.go
set applicationservice_source=netfabbapplicationservice.go netfabbstorage_config.go service.go

::echo Install required GO packages
::go get %PackageGoSqlite% %PackageGoUuid%
if %errorlevel% == 0 (
  goto buildServers
) else (
  goto END
)


:buildServers
cd /d %Binpath%
if exist NetfabbApplicationServer.exe del NetfabbApplicationServer.exe
if exist NetfabbApplicationService.exe del NetfabbApplicationService.exe
if exist NetfabbTaskServer.exe del NetfabbTaskServer.exe
if exist NetfabbStorageServer.exe del NetfabbStorageServer.exe
cd /d %Sourcepath%
echo Building Application Server
go build -o %ApplicationServer% %applicationserver_source% %common_source% 
echo Building Application Service
go build -o %ApplicationService% %applicationservice_source% %common_source% 
REM echo Building Task Server
REM go build -o %TaskServer% %taskserver_source% %common_source%
cd /d %Scriptpath%
if %errorlevel% == 0 (
  goto buildPasswordSalter
) else (
  goto END
)

:buildPasswordSalter
cd /d %PWSalterpath%
if exist PasswordSalter.exe del PasswordSalter.exe
echo Building PasswordSalter
go build
cd /d %Scriptpath%
if %errorlevel% == 0 (
  goto prepareOutputDir
) else (
  goto END
)


:prepareOutputDir
echo Preparing Output DIR
if not exist %Outputpath% mkdir %Outputpath%
cd /d %Outputpath%
for /F "delims=" %%i in ('dir /b') do (rmdir "%%i" /s/q || del "%%i" /s/q)
cd /d %Scriptpath% 
if not exist %Outputpath%Examples mkdir %Outputpath%\Examples
if not exist %Outputpath%Documentation mkdir %Outputpath%\Documentation
xcopy /s /i %Binpath%\* %Outputpath%
xcopy /s /i %PWSalterpath%\PasswordSalter.exe %Outputpath%
xcopy /s /i %Examplepath%\* %Outputpath%\Examples
xcopy /s /i %Docpath%\* %Outputpath%\Documentation

:END
echo Done creating Application Server