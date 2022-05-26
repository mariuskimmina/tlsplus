package acme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
    "encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/mholt/acmez/acme"
)

const (
    etcDir = "/etc/coredns/"
)

func encodeKey(privateKey *ecdsa.PrivateKey) ([]byte) {
    x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
    pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

    return pemEncoded
}

func NewTLSConfigWithACMEFromArgs(conf *dnsserver.Config, domainName string) (*tls.Config, error) {
    fmt.Println("NewTLSConfigWithACMEFromArgs")
    fmt.Printf("Let's get a cert for: %s \n", domainName)

    domains := []string{domainName}

    ctx := context.Background()

    fmt.Println("Creating PKey")
    certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating certificate key: %v", err)
	}

    fmt.Println("Creating CSR")
    csrTemplate := &x509.CertificateRequest{DNSNames: domains}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, certPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("generating CSR: %v", err)
	}
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return nil, fmt.Errorf("parsing generated CSR: %v", err)
	}

    fmt.Println("Creating Account PKey")
    accountPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating account key: %v", err)
	}

	account := acme.Account{
		Contact:              []string{"mailto:test@test.test"},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

    client := &acme.Client{
		Directory: "https://127.0.0.1:14000/dir", // default pebble endpoint
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // we're just tinkering locally - REMOVE THIS FOR PRODUCTION USE!
				},
			},
		},
	}


    fmt.Println("Creating Account")
    account, err = client.NewAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("new account: %v", err)
	}

    // now we can actually get a cert; first step is to create a new order
    fmt.Println("Creating Order")
	var ids []acme.Identifier
	for _, domain := range domains {
		ids = append(ids, acme.Identifier{Type: "dns", Value: domain})
	}
	order := acme.Order{Identifiers: ids}
	order, err = client.NewOrder(ctx, account, order)
	if err != nil {
		return nil, fmt.Errorf("creating new order: %v", err)
	}

    // each identifier on the order should now be associated with an
	// authorization object; we must make the authorization "valid"
	// by solving any of the challenges offered for it
	for _, authzURL := range order.Authorizations {
        fmt.Println("Getting Challenge")
		authz, err := client.GetAuthorization(ctx, account, authzURL)
		if err != nil {
			return nil, fmt.Errorf("getting authorization %q: %v", authzURL, err)
		}

		// pick any available challenge to solve
        var challenge acme.Challenge

        i := 0
        for {
            challenge = authz.Challenges[i]
            if challenge.Type != "dns-01" {
                fmt.Println("Not DNS")
            } else {
                fmt.Println("DNS")
                break
            }
            i++
        }

        solver := DNSSolver{
            Addr: "127.0.0.1:53",
            Config: conf,
        }
        ctx := context.Background()

        //Prepare to solve the challenge
        solver.Present(ctx, challenge)

        fmt.Println("Challenge URL:", challenge.URL)
        fmt.Println(challenge.DNS01TXTRecordName())
        fmt.Println(challenge.Identifier)

		// once you are ready to solve the challenge, let the ACME
		// server know it should begin
        fmt.Println("Starting Challenge Now!")
		challenge, err = client.InitiateChallenge(ctx, account, challenge)
		if err != nil {
			return nil, fmt.Errorf("initiating challenge %q: %v", challenge.URL, err)
		}

        // wait until the challenge has been solved
        solver.Wait(ctx, challenge)

		// now the challenge should be under way; at this point, we can
		// continue initiating all the other challenges so that they are
		// all being solved in parallel (this saves time when you have a
		// large number of SANs on your certificate), but this example is
		// simple, so we will just do one at a time; we wait for the ACME
		// server to tell us the challenge has been solved by polling the
		// authorization status
		authz, err = client.PollAuthorization(ctx, account, authz)
		if err != nil {
			return nil, fmt.Errorf("solving challenge: %v", err)
		}

		// if we got here, then the challenge was solved successfully, hurray!
        fmt.Println("HAPPY SUCCESS! - Let's clean up")
        solver.CleanUp(ctx, challenge)
	}

    // to request a certificate, we finalize the order; this function
	// will poll the order status for us and return once the cert is
	// ready (or until there is an error)
    fmt.Println("Finalizing Order")
	order, err = client.FinalizeOrder(ctx, account, order, csr.Raw)
	if err != nil {
		return nil, fmt.Errorf("finalizing order: %v", err)
	}

	// we can now download the certificate; the server should actually
	// provide the whole chain, and it can even offer multiple chains
	// of trust for the same end-entity certificate, so this function
	// returns all of them; you can decide which one to use based on
	// your own requirements
	certChains, err := client.GetCertificateChain(ctx, account, order.Certificate)
	if err != nil {
		return nil, fmt.Errorf("downloading certs: %v", err)
	}

    var tlsCerts []tls.Certificate


    err = os.MkdirAll(etcDir, os.ModePerm)
    if err != nil {
        fmt.Println("Error during os.MkdirAll")
    }

    certPrivKeyPem := encodeKey(certPrivateKey)
    fmt.Println(string(certPrivKeyPem))

	// all done! store it somewhere safe, along with its key
	for _, cert := range certChains {
		fmt.Printf("Certificate %q:\n%s\n\n", cert.URL, cert.ChainPEM)
        err := os.WriteFile(etcDir + "cert.pem", cert.ChainPEM, 0644)
        if err != nil {
            fmt.Println("Error Writing cert.pem")
            fmt.Println(err)
        }
        err = os.WriteFile(etcDir + "key.pem", encodeKey(certPrivateKey), 0644)
        if err != nil {
            fmt.Println("Error Writing key.pem")
            fmt.Println(err)
        }

        //tlsCert, err := tls.LoadX509KeyPair(string(cert.ChainPEM), certPrivateKey.D.String())
        //if err != nil {
            //fmt.Println("Error Loading Certificate!!!")
            //fmt.Println(err)
        //}
        //tlsCerts = append(tlsCerts, tlsCert)
	}


    tls := &tls.Config{
        Certificates: tlsCerts,
    }
    //var err error
    fmt.Println("End of NewTLSConfigWithACMEFromArgs")
	return tls, nil
}

// exists returns whether the given file or directory exists

