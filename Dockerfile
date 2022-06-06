from golang

RUN git clone https://github.com/coredns/coredns.git
WORKDIR coredns

# step 2. replace the tls plugin with tlsplus
RUN sed -i '/tls:tls/c\tls:github.com/mariuskimmina/tlsplus' plugin.cfg

RUN go get -u github.com/mariuskimmina/tlsplus@testing
RUN go mod tidy
RUN make

Copy ./test/Corefile Corefile

CMD ["./coredns"]
