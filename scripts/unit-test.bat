@echo off
setlocal enabledelayedexpansion

:: Resolve the absolute path of the project root from the script path.
set rootpath=%~dp0%
set rootpath=%rootpath:~0,-9%
cd %rootpath%

:: Run the unit tests.
echo Running unit tests. Please wait...
go test -failfast -short -coverprofile coverage.out ./... || exit /B 1

:: Find the unit tests coverage.
set coveragethreshold=95.0%
set coverage=0.0%
for /f "tokens=3" %%i in ('go tool cover -func ./coverage.out') do set coverage=%%i
call :percentage_string_gte %coverage% %coveragethreshold% coveragepassed
if %coveragepassed% equ 0 (
  echo Checking test coverage failed. Coverage %coverage% is less than %coveragethreshold%.
  exit /B 1
)

:: Show information related to test results.
echo Unit tests succeeded:
echo    Coverage           %coverage%
echo    Coverage Threshold %coveragethreshold%%%
echo Unit tests completed.
exit /B 0

:: Compare two percentage strings by checking whether the lhs is greater or equal than rhs. Both
:: strings must be in a numeric string format with one decimal number and with a percentage suffix.
:percentage_string_gte
set lhs=%~1
set rhs=%~2
set lhs_decimal=%lhs:~-1%
set rhs_decimal=%rhs:~-1%
set lhs=%lhs:~0,-2%
set rhs=%rhs:~0,-2%
set lhs=%lhs%%lhs_decimal%
set rhs=%rhs%%rhs_decimal%
set /a lhs=%lhs%
set /a rhs=%rhs%
if %lhs% geq %rhs% ( set %~3=1 ) else ( set %~3=0 )
exit /B 0
