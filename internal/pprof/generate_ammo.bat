@echo off
for /f "tokens=1,2 delims==" %%a in (.env) do set %%a=%%b
docker exec -i postgres-l0 psql -U postgres -d "orders_db" -t -c "SELECT order_uid FROM \"order\" ORDER BY RANDOM() LIMIT 100 > order_ids_temp.txt
echo. > ammo.txt
for /f "tokens=*" %%i in (order_ids_temp.txt) do (
    set "line=%%i"
    setlocal enabledelayedexpansion
    set "line=!line: =!"
    if not "!line!"=="" echo /order/!line! >> ammo.txt
    endlocal
)
del order_ids_temp.txt
echo Ammo file generated