@echo off

set old_pwd=%CD%

set curr_path=%~dp0

set curr_path=%curr_path:~0,-1%
set curr_path=%curr_path:\=/%

set instantclient_path=%curr_path%/instantclient_12_2

echo prefix=%instantclient_path%>%instantclient_path%/oci8.pc
echo exec_prefix=%instantclient_path%>>%instantclient_path%/oci8.pc
echo libdir=%instantclient_path%/sdk/lib/msvc>>%instantclient_path%/oci8.pc
echo includedir=%instantclient_path%/sdk/include>>%instantclient_path%/oci8.pc
echo. >>%instantclient_path%/oci8.pc
echo glib_genmarshal=glib-genmarshal>>%instantclient_path%/oci8.pc
echo gobject_query=gobject-query>>%instantclient_path%/oci8.pc
echo glib_mkenums=glib-mkenums>>%instantclient_path%/oci8.pc
echo. >>%instantclient_path%/oci8.pc
echo Name: oci8>>%instantclient_path%/oci8.pc
echo Description: oci8 library>>%instantclient_path%/oci8.pc
echo Libs: -L${libdir} -loci>>%instantclient_path%/oci8.pc
echo Cflags: -I${includedir}>>%instantclient_path%/oci8.pc
echo Version: 12.2>>%instantclient_path%/oci8.pc


if defined MSYS_HOME (
  set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig
) else (
  set PKG_CONFIG_PATH=%MSYS_HOME%\mingw64\lib\pkgconfig
)

copy /Y %instantclient_path:/=\%\oci8.pc %PKG_CONFIG_PATH%

SET PKG_CONFIG_PATH=%PKG_CONFIG_PATH%

echo %PKG_CONFIG_PATH%

SET CGO_ENABLED=1
SET GO111MODULE=on

cd %curr_path%/..


if "%1" == "386" (
SET GOARCH=386
SET BUILD_TAGS="windows"
) else (
SET GOARCH=amd64
SET BUILD_TAGS="windows"
)

SET GOOS=windows

go build --tags %BUILD_TAGS%

cd %old_pwd%
