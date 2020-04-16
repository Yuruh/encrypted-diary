package ovh

import (
	asserthelper "github.com/stretchr/testify/assert"
	"os"
	"testing"
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

// là, dans part, j'ai tout ce qu'il faut
	err = UploadFileToPrivateObjectStorage("test_file_upload", file)
	assert.Nil(err)
}