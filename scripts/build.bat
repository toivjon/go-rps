@echo off

:: Resolve the absolute path of the project root from the script path.
set rootpath=%~dp0%
set rootpath=%rootpath:~0,-9%

:: Specify the folder where to put build results.
set binpath=%rootpath%\bin

:: Clear the old binary folder by re-creating it.
if exist "%binpath%" rd /s /q "%binpath%" || exit /B 1
mkdir %binpath%

:: Build the binaries.
echo Building the binaries. Please wait...
go build -o %binpath% %rootpath%\cmd\server || exit /B 1
go build -o %binpath% %rootpath%\cmd\client || exit /B 1

:: Show information related to compilation.
echo Build succeeded:
echo     Server    %binpath%\server
echo     Client    %binpath%\client
echo Build completed.
