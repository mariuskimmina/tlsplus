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
	"github.com/coredns/coredns/plugin/pkg/tls"
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

func acmeCertValid() (bool, error) {
    cert, err := ctls.LoadX509KeyPair(acmeCertFile, acmeKeyFile)
    if err != nil {
        return false, fmt.Errorf("could not load TLS cert: %s", err.Error())
    }
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

	if config.TLSConfig != nil {
		return plugin.Error("tls", c.Errf("TLS already configured for this server instance"))
	}
    i := 1
	for c.Next() {
        var tlsconf *ctls.Config
        fmt.Printf("Run number: %d \n", i)
        i++
        args := c.RemainingArgs()
        fmt.Printf("remaining args: %s \n", args)
		clientAuth := ctls.NoClientCert

        if args[0] == "acme" {
            certPresent := false
            certValid := false
            fmt.Println("Starting ACME")
            certPresent, err := acmeCertPresent()
            if err != nil {
                return err
            }
            if certPresent {
                certValid, err = acmeCertValid()
                if err != nil {
                    return err
                }
                if certValid {
                    fmt.Println("Valid Cert found")
                } else {
                    fmt.Println("Cert found but expired")
                }
            }
            fmt.Println(certPresent)
            fmt.Println(certValid)
            if !certPresent || !certValid {
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
                tlsconf, err = tls.NewTLSConfigWithACMEFromArgs(config, domainNameACME)
                if err != nil {
                    fmt.Println("Error during TLS Config with ACME")
                    fmt.Println(err)
                }
            }
            fmt.Println("Certificate aleady there")
            config.TLSConfig = tlsconf
        } else {
            fmt.Println("NOOO ACME")
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
            tlsconf, err := tls.NewTLSConfigFromArgs(args...)
            if err != nil {
                return err
            }
            tlsconf.ClientAuth = clientAuth
            // NewTLSConfigFromArgs only sets RootCAs, so we need to let ClientCAs refer to it.
            tlsconf.ClientCAs = tlsconf.RootCAs

            config.TLSConfig = tlsconf
        }
	}
	return nil
}
