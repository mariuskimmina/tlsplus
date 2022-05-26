# Clone CoreDNS
git clone https://github.com/coredns/coredns
cd coredns

# Add acme:github.com/chinzhiweiblank/coredns-acme into the plugin configuration
echo "tlsp:github.com/mariuskimmina/tlsplus" >> plugin.cfg

# Get the modules
go get github.com/mariuskimmina/tlsplus

# Generate Files
go generate

# Tidy the modules
go mod tidy

# Compile
go build
