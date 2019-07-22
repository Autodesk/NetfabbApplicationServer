@ECHO OFF
if exist Bin\NetfabbStorageServer.exe del Bin\NetfabbStorageServer.exe
if exist Bin\NetfabbTaskServer.exe del Bin\NetfabbTaskServer.exe
if exist Bin\NetfabbApplicationServer.exe del Bin\NetfabbApplicationServer.exe

echo Building Application Server
go build -o Bin/NetfabbApplicationServer.exe Source/netfabbapplicationserver.go Source/netfabbstorage_db.go Source/netfabbstorage_types.go Source/netfabbstorage_utils.go  Source/netfabbstorage_config.go Source/netfabbstorage_auth.go Source/netfabbstorage_orm.go Source/netfabbstorage_data.go Source/netfabbtask_handler.go Source/netfabbtask_pan.go Source/netfabbapplication.go 

echo Building Application Service
go build -o Bin/NetfabbApplicationService.exe Source/netfabbapplicationservice.go Source/netfabbstorage_db.go Source/netfabbstorage_types.go Source/netfabbstorage_utils.go  Source/netfabbstorage_config.go Source/netfabbstorage_auth.go Source/netfabbstorage_orm.go Source/netfabbstorage_data.go Source/netfabbtask_handler.go Source/netfabbtask_pan.go Source/netfabbapplication.go Source/service.go
