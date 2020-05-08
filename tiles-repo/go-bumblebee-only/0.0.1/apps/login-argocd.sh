#!/usr/bin/expect -f
spawn argocd login 127.0.0.1:8080 --insecure
expect "Username: "
send "x_user_x\r"
expect "Password: "
send "x_password_x\r"
interact