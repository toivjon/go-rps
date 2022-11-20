@echo off

:: Resolve the absolute path of the project root from the script path.
set rootpath=%~dp0%
set rootpath=%rootpath:~0,-9%
cd %rootpath%

:: Run the system tests.
echo Running system tests. Please wait...
go run ./systest/client || exit /B 1
go run ./systest/server || exit /B 1

:: Show information related to test results.
echo System tests succeeded.
exit /B 0
