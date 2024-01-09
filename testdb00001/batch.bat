set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_PASSWORD=p05tgres
set SERVER_PORT=3456
set DB_NAME=db
set InjectTestData=InjectTestData
set TEST_SERVER_PORT=3123

echo DB_HOST set to %DB_HOST%
echo DB_PORT set to %DB_PORT%
echo DB_USER set to %DB_USER%
echo DB_PASSWORD set to %DB_PASSWORD%
echo SERVER_PORT set to %SERVER_PORT%
echo DB_NAME set to %DB_NAME%
echo InjectTestData set to %InjectTestData%

@REM #!/bin/bash
@REM #for mac and linux
@REM export DB_HOST=dbhostname
@REM export DB_PORT=5432
@REM export DB_USER=postgres
@REM export DB_PASSWORD=p05tgres
@REM export SERVER_PORT=3003
@REM export DB_NAME=db
@REM export InjectTestData=InjectTestData

@REM echo "DB_HOST set to $DB_HOST"
@REM echo "DB_PORT set to $DB_PORT"
@REM echo "DB_USER set to $DB_USER"
@REM echo "DB_PASSWORD set to $DB_PASSWORD"
@REM echo "SERVER_PORT set to $SERVER_PORT"
@REM echo "DB_NAME set to $DB_NAME"
@REM echo "InjectTestData set to $InjectTestData"