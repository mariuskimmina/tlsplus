#!/bin/bash

# step 1. get coredns
git clone https://github.com/coredns/coredns.git
cd coredns

# step 2. replace the tls plugin with tlsplus
sed -i '/tls:tls/c\tls:github.com/mariuskimmina/tlsplus' plugin.cfg

go get -u github.com/mariuskimmina/tlsplus@testing
go mod tidy
make

cp ../Corefile Corefile

# step 4. run CoreDNS
./coredns
