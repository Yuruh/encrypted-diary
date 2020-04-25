package authentication

import (
	"encoding/base32"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/skip2/go-qrcode"
	"math/rand"
	"net/url"
)

func GenerateQRCodeFromURI(uri string) ([]byte, error) {
	png, err := qrcode.Encode(uri, qrcode.Medium, 256)
	if err != nil {
		err = fmt.Errorf("could not encore: %v", err)
		return nil, err
	}
	return png, err
}

func GenerateRandomSecret() string {
	const possibilities = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const secretLength = 30

	b := make([]byte, secretLength)
	for i := range b {
		b[i] = possibilities[rand.Intn(len(possibilities))]
	}
	return base32.StdEncoding.EncodeToString(b)
}

func BuildGAuthURI(userEmail string, secret string) string {
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format

	secretBase32 := base32.StdEncoding.EncodeToString([]byte(secret))

	account := userEmail
	issuer := "EncryptedDiary"

	URL, err := url.Parse("otpauth://totp")
	if err != nil {
		panic(err)
	}

	URL.Path += "/" + url.PathEscape(issuer) + ":" + url.PathEscape(account)

	params := url.Values{}
	params.Add("secret", secretBase32)
	params.Add("issuer", issuer)
	params.Add("algorithm", "SHA512")

	URL.RawQuery = params.Encode()

	return URL.String()
}

func Authorize(passCode string, secret string) (bool, error) {
	otpc := &dgoogauth.OTPConfig{
		Secret:      secret,
		WindowSize:  2,
		HotpCounter: 0,
	}
	val, err := otpc.Authenticate(passCode)
	if err != nil {
		err = fmt.Errorf("could not authenticate: %v", err)
		return false, err
	}
	return val, err
}