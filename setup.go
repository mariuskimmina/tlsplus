package tlsplus

import (
	ctls "crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mariuskimmina/tlsplus/tls"
	"github.com/mariuskimmina/tlsplus/acme"
)


func init() { plugin.Register("tls", setup) }

func setup(c *caddy.Controller) error {
	err := parseTLS(c)
	if err != nil {
		return plugin.Error("tls", err)
	}

	return nil
}

const (
    acmeCertFile = "/etc/coredns/cert.pem"
    acmeKeyFile = "/etc/coredns/key.pem"
)

func fileExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

func acmeCertPresent() (bool, error) {
    present, err := fileExists(acmeCertFile)
    if err != nil {
        return false, err
    }
    return present, nil
}

func acmeCertValid(cert *ctls.Certificate) (bool, error) {
    parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
    if err != nil {
        return false, fmt.Errorf("could not parse TLS cert: %s", err.Error())
    }
    valid := parsedCert.NotAfter.After(time.Now())

    return valid, nil
}

func parseTLS(c *caddy.Controller) (error) {
    //args := c.RemainingArgs()
    //fmt.Printf("starting to parse tls config - args: %s \n", args)
	config := dnsserver.GetConfig(c)
    var tlsconf *ctls.Config
    var err error
    clientAuth := ctls.NoClientCert

	if config.TLSConfig != nil {
		return plugin.Error("tls", c.Errf("TLS already configured for this server instance"))
	}
    i := 1
	for c.Next() {
        fmt.Printf("Run number: %d \n", i)
        i++
        args := c.RemainingArgs()
        fmt.Printf("remaining args: %s \n", args)

        if args[0] == "acme" {
            // start of the acme flow,
            // first check if a certificate is already present
            fmt.Println("Starting ACME")
            certPresent := false
            //certValid := false
            certPresent, err := acmeCertPresent()
            if err != nil {
                return err
            }
            if certPresent {
                // TODO: check if the certificate is valid
                fmt.Println("Loading existing certificate")
                tlsconf, err = tls.NewTLSConfig(acmeCertFile, acmeKeyFile, "")
                fmt.Println("Certificate aleady there")

                configureTLS(config, tlsconf, clientAuth)
                return nil

            }
            fmt.Println("No valid Certificate found, creating a new one")
            var domainNameACME string
            for c.NextBlock() {
                fmt.Println("ACME Block Found")
                switch c.Val() {
                case "domain":
                    fmt.Println("Found Keyword Domain")
                    domainArgs := c.RemainingArgs()
                    if len(domainArgs) > 1 {
                        return plugin.Error("tls", c.Errf("To many arguments to domain"))
                    }
                    domainNameACME = domainArgs[0]
                    fmt.Println(domainNameACME)
                }
            }
            config := dnsserver.GetConfig(c)
            tlsconf, err = acme.NewTLSConfigWithACMEFromArgs(config, domainNameACME)
            if err != nil {
                fmt.Println("Error during TLS Config with ACME")
                fmt.Println(err)
            }
            tlsconf, err = tls.NewTLSConfig(acmeCertFile, acmeKeyFile, "")
            fmt.Println("Certificate aleady there")
        } else {
            fmt.Println("Uing manually conigured certificate")
            if len(args) < 2 || len(args) > 3 {
                return plugin.Error("tls", c.ArgErr())
            }
            for c.NextBlock() {
                fmt.Println("Next Block")
                switch c.Val() {
                case "client_auth":
                    authTypeArgs := c.RemainingArgs()
                    if len(authTypeArgs) != 1 {
                        return c.ArgErr()
                    }
                    switch authTypeArgs[0] {
                    case "nocert":
                        clientAuth = ctls.NoClientCert
                    case "request":
                        clientAuth = ctls.RequestClientCert
                    case "require":
                        clientAuth = ctls.RequireAnyClientCert
                    case "verify_if_given":
                        clientAuth = ctls.VerifyClientCertIfGiven
                    case "require_and_verify":
                        clientAuth = ctls.RequireAndVerifyClientCert
                    default:
                        return c.Errf("unknown authentication type '%s'", authTypeArgs[0])
                    }
                default:
                    fmt.Println("Default error")
                    return c.Errf("unknown option '%s'", c.Val())
                }
            }
            tlsconf, err = tls.NewTLSConfigFromArgs(args...)
            if err != nil {
                return err
            }
        }
	}
    configureTLS(config, tlsconf, clientAuth)
	return nil
}
