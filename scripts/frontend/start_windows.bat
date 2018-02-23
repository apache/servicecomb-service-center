# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
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