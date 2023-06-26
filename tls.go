package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	r := randBytes(n)

	b := make([]byte, n)
	l := len(letters)
	for i := range b {
		b[i] = letters[int(r[i])%l]
	}
	return string(b)
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	_, _ = rand.Read(r)
	return r
}

func RandBigInt(max *big.Int) *big.Int {
	r, _ := rand.Int(rand.Reader, max)
	return r
}

func genPair(keysize int) (caCert []byte, caCertKey []byte, cert []byte, certKey []byte) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	ca := &x509.Certificate{
		SerialNumber: RandBigInt(serialNumberLimit),
		Subject: pkix.Name{
			Country:            []string{randString(16)},
			Organization:       []string{randString(16)},
			OrganizationalUnit: []string{randString(16)},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          randBytes(5),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	private, _ := rsa.GenerateKey(rand.Reader, keysize)
	pub := &private.PublicKey
	caBin, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, private)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}

	cert2 := &x509.Certificate{
		SerialNumber: RandBigInt(serialNumberLimit),
		Subject: pkix.Name{
			Country:            []string{randString(16)},
			Organization:       []string{randString(16)},
			OrganizationalUnit: []string{randString(16)},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: randBytes(6),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	private2, _ := rsa.GenerateKey(rand.Reader, keysize)
	pub2 := &private2.PublicKey
	cert2Bin, err2 := x509.CreateCertificate(rand.Reader, cert2, ca, pub2, private)
	if err2 != nil {
		log.Println("create cert2 failed", err2)
		return
	}

	privateBin := x509.MarshalPKCS1PrivateKey(private)
	private2Bin := x509.MarshalPKCS1PrivateKey(private2)

	return caBin, privateBin, cert2Bin, private2Bin

}

func getPEMs(cert []byte, key []byte) (pemcert []byte, pemkey []byte) {
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: key,
	})

	return certPem, keyPem
}

func getTLSPair(certPem []byte, keyPem []byte) (tls.Certificate, error) {
	tlspair, errt := tls.X509KeyPair(certPem, keyPem)
	if errt != nil {
		return tlspair, errt
	}
	return tlspair, nil
}

func getRandomTLS(keysize int) (tls.Certificate, error) {
	_, _, cert, certkey := genPair(keysize)
	certPem, keyPem := getPEMs(cert, certkey)
	tlspair, err := getTLSPair(certPem, keyPem)
	return tlspair, err
}
