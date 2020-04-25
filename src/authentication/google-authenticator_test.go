package authentication

import (
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildGAuthURI(t *testing.T) {
	assert := asserthelper.New(t)
	assert.Equal("otpauth://totp/EncryptedDiary:antoine.lempereur@epitech.eu?algorithm=SHA512&issuer=EncryptedDiary&secret=KNKGW6KMMVHEINTLJQ3VO2ZROVUEQ3CKJFBWUZLIJFBDKZCLJBSQ%3D%3D%3D%3D",
		BuildGAuthURI("antoine.lempereur@epitech.eu"))
}

func TestGenerateQRCodeFromURI(t *testing.T) {
	assert := asserthelper.New(t)
	png, err := GenerateQRCodeFromURI("otpauth://totp/EncryptedDiary:antoine.lempereur@epitech.eu?algorithm=SHA512&issuer=EncryptedDiary&secret=KNKGW6KMMVHEINTLJQ3VO2ZROVUEQ3CKJFBWUZLIJFBDKZCLJBSQ%3D%3D%3D%3D")

	assert.Nil(err)
	assert.Equal(1062, len(png))
	assert.Equal("\u0089", string(png[0]))

	_, err = GenerateQRCodeFromURI("")

	assert.NotNil(err)
}

func TestGenerateRandomSecret(t *testing.T) {
	assert := asserthelper.New(t)

	s1 := GenerateRandomSecret()
	assert.Equal(60, len(s1))
	s2 := GenerateRandomSecret()
	assert.Equal(len(s1), len(s2))
	assert.NotEqual(s1, s2)
}