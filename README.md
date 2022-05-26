# tlsplus

## Name

*tlsplus* generates and manages TLS certificates for DNS over TLS or HTTPS. You can also optionally configure the certificates yourself.

## Description

CoreDNS supports queries that are encrypted using TLS (DNS over Transport Layer Security, RFC 7858)
or are using gRPC (https://grpc.io/, not an IETF standard). Normally DNS traffic isn't encrypted at
all (DNSSEC only signs resource records).

The *tls* "plugin" allows you to either have CoreDNS generate and manage certificates for itself or configure the cryptographic keys that are needed for both
DNS-over-TLS and DNS-over-gRPC. If the *tls* plugin is omitted, then no encryption takes place.

## Syntax

### Automatic

~~~ txt
tlsplus acme {
    domain example.com
}
~~~

### Manual

~~~ txt
tls CERT KEY [CA]
~~~

Parameter CA is optional. If not set, system CAs can be used to verify the client certificate

~~~ txt
tls CERT KEY [CA] {
    client_auth nocert|request|require|verify_if_given|require_and_verify
}
~~~

If client\_auth option is specified, it controls the client authentication policy.
The option value corresponds to the [ClientAuthType values of the Go tls package](https://golang.org/pkg/crypto/tls/#ClientAuthType): NoClientCert, RequestClientCert, RequireAnyClientCert, VerifyClientCertIfGiven, and RequireAndVerifyClientCert, respectively.
The default is "nocert".  Note that it makes no sense to specify parameter CA unless this option is
set to verify\_if\_given or require\_and\_verify.

## Examples

Start a DNS-over-TLS server that picks up incoming DNS-over-TLS queries on port 5553 and uses the
nameservers defined in `/etc/resolv.conf` to resolve the query. This proxy path uses plain old DNS.

~~~
tls://.:5553 {
	tls cert.pem key.pem ca.pem
	forward . /etc/resolv.conf
}
~~~

Start a DNS-over-gRPC server that is similar to the previous example, but using DNS-over-gRPC for
incoming queries.

~~~
grpc://. {
	tls cert.pem key.pem ca.pem
	forward . /etc/resolv.conf
}
~~~

Start a DoH server on port 443 that is similar to the previous example, but using DoH for incoming queries.
~~~
https://. {
	tls cert.pem key.pem ca.pem
	forward . /etc/resolv.conf
}
~~~

Only Knot DNS' `kdig` supports DNS-over-TLS queries, no command line client supports gRPC making
debugging these transports harder than it should be.

## See Also

RFC 7858 and https://grpc.io.
