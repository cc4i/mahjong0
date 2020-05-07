#!/usr/bin/expect -f
spawn argocd login 127.0.0.1:8080 --insecure
expect "Username: "
send "${username}\r"
expect "Password: "
send "${password}\r"
interact