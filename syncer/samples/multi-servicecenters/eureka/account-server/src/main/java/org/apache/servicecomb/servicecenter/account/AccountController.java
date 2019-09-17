/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.apache.servicecomb.servicecenter.account;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/")
public class AccountController {
    @GetMapping(value = "/checkHealth", produces = "application/json;charset=UTF-8")
    public ResponseEntity<String> checkHealth() {
        return new ResponseEntity<>("I am healthy", HttpStatus.OK);
    }

    @PostMapping(value = "/login", produces = "application/json;charset=UTF-8")
    @ResponseBody
    public ResponseEntity<String> login(@RequestBody LoginVO loginVO) {
        if (loginVO.user.isEmpty()) {
            return new ResponseEntity<>("user is empty", HttpStatus.UNAUTHORIZED);
        }

        if (loginVO.password.isEmpty()) {
            return new ResponseEntity<>("password is empty", HttpStatus.UNAUTHORIZED);
        }

        // todo: check user and password
        return new ResponseEntity<>("welcome " + loginVO.user, HttpStatus.OK);
    }

    static class LoginVO {
        String user;
        String password;

        public String getUser() {
            return user;
        }

        public void setUser(String user) {
            this.user = user;
        }

        public String getPassword() {
            return password;
        }

        public void setPassword(String password) {
            this.password = password;
        }
    }
}
