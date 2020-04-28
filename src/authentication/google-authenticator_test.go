package authentication

import (
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildGAuthURI(t *testing.T) {
	assert := asserthelper.New(t)
	assert.Equal("otpauth://totp/EncryptedDiary:antoine.lempereur@epitech.eu?algorithm=SHA512&issuer=EncryptedDiary&secret=STkyLeND6kL7Wk1uhHlJICjehIB5dKHe",
		BuildGAuthURI("antoine.lempereur@epitech.eu", "STkyLeND6kL7Wk1uhHlJICjehIB5dKHe"))
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

	assert.Equal(16, len(s1))
	s2 := GenerateRandomSecret()
	assert.Equal(len(s1), len(s2))
	assert.NotEqual(s1, s2)
}

func TestAuthorize(t *testing.T) {
	assert := asserthelper.New(t)

	validate, err := Authorize("123456", "secret")
	assert.Nil(err)
	assert.Equal(false, validate)

	validate, err = Authorize("12345azerazer6", "secret")
	assert.Equal(false, validate)
	assert.NotNil(err)

}