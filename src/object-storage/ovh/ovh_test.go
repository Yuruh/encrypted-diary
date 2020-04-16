package ovh

import (
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
	Utility to generate ovh consumer key,
func TestGetOvhConsumerKey(t *testing.T) {
	key, err := GetOvhConsumerKey()
	if assert.Nil(t, err) {
		assert.Equal(t, "12", key.ConsumerKey)
		assert.Equal(t, "pending", key.State)
		assert.Equal(t, "yolo", key.ValidationURL)
	}
}
*/

func TestUploadFileToPrivateObjectStorage(t *testing.T) {
	assert := asserthelper.New(t)

	file, err := os.Open("testdata/joe_le_pangolin.jpg")
	if err != nil {
		t.Fatal(err.Error())
	}

	err = UploadFileToPrivateObjectStorage("test_file_upload", file)
	assert.Nil(err)
}

func TestGetFileTemporaryAccess(t *testing.T) {
	assert := asserthelper.New(t)

	// We assume the file "test_file_upload" is already stored on the server
	fileAccess, err := GetFileTemporaryAccess("test_file_upload", time.Second * 2)
	assert.Nil(err)

	parsedTime, err := time.Parse(time.RFC3339, fileAccess.ExpirationDate)
	assert.Nil(err)

	assert.Greater(parsedTime.Second(), time.Now().Second())
	assert.Greater(time.Now().Add(time.Second * 4).Second(), parsedTime.Second())

	req, err := http.NewRequest(http.MethodGet, fileAccess.URL, nil)
	assert.Nil(err)

	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode)

	time.Sleep(time.Second * 3)
	res, err = client.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusUnauthorized, res.StatusCode)
}