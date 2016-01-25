package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"
)

func main() {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"Yjwt"},
			OrganizationalUnit: []string{"YjwtU"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		DNSNames:    []string{"eastany.com"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		fmt.Printf("create ca failed err : %v\n", err)
		return
	}
	ca_f := "ca.pem"
	ioutil.WriteFile(ca_f, ca_b, 0777)
	priv_f := "ca.key"
	priv_b := x509.MarshalPKCS1PrivateKey(priv)
	ioutil.WriteFile(priv_f, priv_b, 0777)
	cert2 := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"EA"},
			OrganizationalUnit: []string{"EAU"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv2, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub2 := &priv2.PublicKey
	cert2_b, err2 := x509.CreateCertificate(rand.Reader, cert2, ca, pub2, priv)
	if err2 != nil {
		fmt.Printf("create ca2 failed err : %v\n", err)
		return
	}
	cert2_f := "cert2.pem"
	ioutil.WriteFile(cert2_f, cert2_b, 0777)
	priv2_f := "cert2.key"
	priv2_b := x509.MarshalPKCS1PrivateKey(priv2)
	ioutil.WriteFile(priv2_f, priv2_b, 0777)
}
