#!/usr/bin/expect -f
spawn argocd login 127.0.0.1:8080 --insecure
expect "Username: "
send "__User__\r"
expect "Password: "
send "__Password__\r"
interact