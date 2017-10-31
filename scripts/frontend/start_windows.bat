@echo off
3<conf/app.conf (
SET /p line1= <&3
SET /p line2= <&3
SET /p line3= <&3
)

FOR /f "tokens=1,2 delims==" %%a in ("%line1%") do SET ip=%%b
FOR /f "tokens=1,2 delims==" %%a in ("%line2%") do SET port=%%b
FOR /f "tokens=1,2 delims==" %%a in ("%line3%") do SET mode=%%b

CD ./app/apiList
SET "file=apiList.js"
SET "temp=temp.js"

SET /a Line#ToSearchUrl=18
SET "REPLACEURL=ip: '%mode%://%ip%',"
SET "URL=%REPLACEURL: =%"
(FOR /f "tokens=1*delims=:" %%a IN ('findstr /n "^" "%file%"') DO (
    SET "IPLine=%%b"
    IF %%a equ %Line#ToSearchUrl% SET "IPLine=%URL%"
    SETLOCAL ENABLEDELAYEDEXPANSION
    ECHO(!IPLine!
    ENDLOCAL
))>"%temp%"
TYPE "%temp%"
DEL %file%
)
cls

SET /a Line#ToSearchPort=19
SET "REPLACEPORT=port: '%port%'"
SET "PORT=%REPLACEPORT: =%"
(FOR /f "tokens=1*delims=:" %%a IN ('findstr /n "^" "%temp%"') DO (
    SET "PORTLine=%%b"
    IF %%a equ %Line#ToSearchPort% SET "PORTLine=%PORT%"
    SETLOCAL ENABLEDELAYEDEXPANSION
    ECHO(!PORTLine!
    ENDLOCAL
))>"%file%"
TYPE "%file%"
DEL %temp%
)
cls
cd ../../
scfrontend.exe