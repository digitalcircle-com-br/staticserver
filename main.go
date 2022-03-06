package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/digitalcircle-com-br/buildinfo"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

//go:embed ca.cer
var caCerPEMBlock []byte

//go:embed ca.key
var caKeyPEMBlock []byte

//go:embed localhost.cer
var certPEMBlock []byte

//go:embed localhost.key
var keyPEMBlock []byte

func serve() error {

	root := os.Getenv("ROOT")
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// cwd := os.Getenv("CWD")
	if root == "" {
		root = "./static"
	}
	fs := http.FileServer(http.Dir(root))

	http.Handle("/", fs)
	srv := &http.Server{}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Listening at: %s, using root: %s", addr, root)

	return srv.Serve(ln)
}

func serveTLS() error {

	root := os.Getenv("ROOT")
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8443"
	}
	// cwd := os.Getenv("CWD")
	if root == "" {
		root = "./static"
	}
	fs := http.FileServer(http.Dir(root))

	log.Printf("adding handler: %s", "/.ca/ca.cer")
	log.Printf("adding handler: %s", "/.ca/ca.key")
	log.Printf("adding handler: %s", "/.ca/server.cer")
	log.Printf("adding handler: %s", "/.ca/server.key")

	http.HandleFunc("/.ca/ca.cer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-pem-file")
		w.Write(caCerPEMBlock)
	})

	http.HandleFunc("/.ca/ca.key", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-pem-file")
		w.Write(caKeyPEMBlock)
	})

	http.HandleFunc("/.ca/server.cer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-pem-file")
		w.Write(certPEMBlock)
	})

	http.HandleFunc("/.ca/server.key", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-pem-file")
		w.Write(keyPEMBlock)
	})

	http.Handle("/", fs)
	srv := &http.Server{}

	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	log.Printf("Listening TLS at: %s, using root: %s", addr, root)
	return srv.Serve(tlsListener)
}

func doGenca() error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:       []string{"Digital Circle - DEV ONLY NOT FOR PRD"},
			Country:            []string{"BR"},
			Province:           []string{"MG"},
			Locality:           []string{"BH"},
			StreetAddress:      []string{"N/A"},
			PostalCode:         []string{"N/A"},
			OrganizationalUnit: []string{"DEVOPS"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1000, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 4096)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return err
	}
	out := &bytes.Buffer{}
	pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b})
	cert := out.Bytes()
	err = ioutil.WriteFile("ca.cer", cert, 0600)
	if err != nil {
		return err
	}
	keyOut, err := os.OpenFile("ca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	return nil
}

func GenCertForDomain(d string) (keybs []byte, certbs []byte, err error) {

	catls, err := tls.LoadX509KeyPair("ca.cer", "ca.key")
	if err != nil {
		return nil, nil, err
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return nil, nil, err
	}

	time.Sleep(time.Nanosecond)
	ser := time.Now().UnixNano()

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(ser),
		Subject: pkix.Name{
			Organization: []string{"Digital Circle - DEV ONLY NOT FOR PRD"},
			Country:      []string{"BR"},
			CommonName:   "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1000, 0, 0),
		//SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		//Extensions: []pkix.Extension{
		//	{
		//		Id:       asn1.ObjectIdentifier{2, 5, 29, 17},
		//		Critical: false,
		//		Value:    rawByte,
		//	},
		//},
	}

	cert.DNSNames = append(cert.DNSNames, d)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	pub := &priv.PublicKey

	// Sign the certificate
	cert_b, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, catls.PrivateKey)
	if err != nil {
		return
	}
	// Public key
	certBuffer := &bytes.Buffer{}
	pem.Encode(certBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: cert_b})

	keyBuffer := &bytes.Buffer{}
	pem.Encode(keyBuffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	keybs = keyBuffer.Bytes()
	certbs = certBuffer.Bytes()
	return
}
func GenCertFilesForDomain(d string) error {
	key, cert, err := GenCertForDomain(d)
	if err != nil {
		return err
	}

	keyfile := d + ".key"
	certfile := d + ".cer"
	err = ioutil.WriteFile(keyfile, key, 0600)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(certfile, cert, 0600)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.Printf("%s", buildinfo.String())
	genca := flag.Bool("genca", false, "if you want to drop a new CA cert")
	gencert := flag.Bool("gencert", false, "if you want a new cert")
	usetls := flag.Bool("tls", false, "if you want ot use tls")
	certname := flag.String("certname", "localhost", "name for new cert")
	flag.Parse()
	if !*usetls {
		*usetls = os.Getenv("TLS") != ""
	}
	switch {
	case *genca:
		doGenca()
	case *gencert:
		GenCertFilesForDomain(*certname)
	case *usetls:
		err := serveTLS()
		if err != nil {
			panic(err)
		}
	default:
		err := serve()
		if err != nil {
			panic(err)
		}
	}

}
