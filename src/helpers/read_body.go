package helpers

import (
	"io"
	"io/ioutil"
	"log"
)

func ReadBody(body io.ReadCloser) string {
	defer body.Close()

	result, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return string(result)
}

func ContainsString(src []string, value string) bool {
	for _, elem := range src {
		if elem == value {
			return true
		}
	}
	return false
}
