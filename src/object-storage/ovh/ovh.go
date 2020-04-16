package ovh

import (
	"errors"
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"io"
	"net/http"
	"os"
	"time"
)

type PartialMe struct {
	Firstname string `json:"firstname"`
}

// https://eu.api.ovh.com/console/#/cloud/project/%7BserviceName%7D/storage/access#POST
type StorageAccess struct {
	Token string `json:"token"`
}

func getStorageAccess() (StorageAccess, error) {
	// Uses env variable for client configuration
	client, err := ovh.NewDefaultClient()

	if err != nil {
		return StorageAccess{}, fmt.Errorf("could not create ovh client: %v", err)
	}
	access := StorageAccess{}
	err = client.Post("/cloud/project/" + os.Getenv("OVH_SERVICE_NAME") + "/storage/access", nil, &access)
	if err != nil {
		return StorageAccess{}, fmt.Errorf("could not get storage access: %v", err)
	}
	return access, nil
}

func UploadFileToPrivateObjectStorage(fileDescriptor string, file io.Reader) error {
	// It does require access every time, but is a singleton the only way to prevent this ?
	access, err := getStorageAccess()
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, os.Getenv("OVH_OPENSTACK_CONTAINER_URL") + fileDescriptor, file)
	if err != nil {
		return fmt.Errorf("could not create http request: %v", err)
	}
	req.Header.Add("X-Auth-Token", access.Token)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload file failed: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		// todo handle
		return errors.New("unexpected status code")
	}

	return nil
}

// the duration could be computed from the time left in access token
func GetFileTemporaryAccess(duration time.Duration) {

}

/* Only needed to generate consumer key, commented for now
func GetOvhConsumerKey() (*ovh.CkValidationState, error) {
	fmt.Println("allo", os.Getenv("OVH_ENDPOINT"))
	client, err := ovh.NewDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("Ovh consumer key: %q\n", err)
	}
	ckReq := client.NewCkRequest()

	ckReq.AddRules(ovh.ReadWriteSafe, "/cloud/project/" + os.Getenv("OVH_SERVICE_NAME") + "/storage/access")
	ckReq.AddRules(ovh.ReadWriteSafe, "/cloud/project/" + os.Getenv("OVH_SERVICE_NAME") + "/storage/" +
		os.Getenv("OVH_CONTAINER_ID") + "/publicUrl")

	return ckReq.Do()
}*/