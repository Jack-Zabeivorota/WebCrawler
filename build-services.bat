set ROOT_DIR=%~dp0

cd "%ROOT_DIR%aggregator"
call builder.bat

cd "%ROOT_DIR%controller"
call builder.bat

cd "%ROOT_DIR%main"
call builder.bat

cd "%ROOT_DIR%worker"
call builder.bat