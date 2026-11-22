// Command devtoken makes a local ES256 keypair and signs development tokens.
//
//	devtoken keygen DIR
//	devtoken sign -key FILE -iss ISSUER -aud AUDIENCE -email EMAIL [-ttl 1h]
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "devtoken:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: devtoken keygen DIR | devtoken sign -key FILE -iss ISSUER -aud AUDIENCE -email EMAIL")
	}
	switch args[0] {
	case "keygen":
		if len(args) != 2 {
			return errors.New("usage: devtoken keygen DIR")
		}
		return keygen(args[1])
	case "sign":
		return sign(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func keygen(dir string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	priv, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "dev.key"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv}), 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "dev.pub"), pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub}), 0o644)
}

func sign(args []string) error {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	keyFile := fs.String("key", "", "PKCS#8 EC private key")
	iss := fs.String("iss", "", "issuer")
	aud := fs.String("aud", "", "audience")
	email := fs.String("email", "", "actor email")
	ttl := fs.Duration("ttl", time.Hour, "lifetime")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *keyFile == "" || *iss == "" || *aud == "" || *email == "" {
		return errors.New("sign needs -key, -iss, -aud and -email")
	}
	raw, err := os.ReadFile(*keyFile)
	if err != nil {
		return err
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(raw)
	if err != nil {
		return err
	}
	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": *iss, "aud": *aud, "email": *email,
		"iat": now.Unix(), "exp": now.Add(*ttl).Unix(),
	}).SignedString(key)
	if err != nil {
		return err
	}
	fmt.Println(token)
	return nil
}
