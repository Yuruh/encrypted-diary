package ovh

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/ovh/go-ovh/ovh"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type PartialMe struct {
	Firstname string `json:"firstname"`
}

// https://eu.api.ovh.com/console/#/cloud/project/%7BserviceName%7D/storage/access#POST
type StorageAccess struct {
	Token string `json:"token"`
}

type ObjectTempPublicUrl struct {
	URL string `json:"getURL"`
	ExpirationDate string `json:"expirationDate"`
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
		// todo handle ovh errors
		fmt.Println(res.StatusCode)
		fmt.Println(helpers.ReadBody(res.Body))
		return errors.New("unexpected status code")
	}

	return nil
}


// Adapted from https://docs.openstack.org/swift/latest/api/temporary_url_middleware.html#hmac-sha1-signature-for-temporary-urls
func generateTempUrlSig(fileDescriptor string, duration time.Duration) ObjectTempPublicUrl {
	method := "GET"

	expires := time.Now().Add(duration)
//	durationSec := duration / time.Second
	expiresSec := strconv.Itoa(int(expires.Unix()))
	path := os.Getenv("OVH_OPENSTACK_CONTAINER_PATH") + fileDescriptor
	key := os.Getenv("OVH_OPENSTACK_TEMP_URL_KEY")
	hmacBody := fmt.Sprintf("%s\n%s\n%s", method, expiresSec, path)
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(hmacBody))
	signature := hex.EncodeToString(hash.Sum(nil))

	return ObjectTempPublicUrl{
		URL: os.Getenv("OVH_OPENSTACK_CONTAINER_URL") + fileDescriptor +
			"?temp_url_expires=" + expiresSec + "&temp_url_sig=" + signature,
		ExpirationDate: expires.Format(time.RFC3339),
	}
}


func GetFileTemporaryAccess(fileDescriptor string, duration time.Duration) (ObjectTempPublicUrl, error) {


	/*
	 *	The version below uses OVH Api call to generate signature. This is the easiest method and potentially more scalable if openstack api changes
	 *	But it is pretty slow: 0.35 seconds according to benchmark
     */
	/*
	 * Indeed, after manual setup, we go from ~350 ms to ~0.01 ms
	 *
	 *
	 */
/*
	client, err := ovh.NewDefaultClient()
	if err != nil {
		return ObjectTempPublicUrl{}, err
	}

	var url ObjectTempPublicUrl

	err = client.Post("/cloud/project/" + os.Getenv("OVH_SERVICE_NAME") + "/storage/" +
		os.Getenv("OVH_CONTAINER_ID") + "/publicUrl", map[string]interface{}{
		"expirationDate": time.Now().Add(duration).Format(time.RFC3339),
		"objectName": fileDescriptor,
		}, &url)

	if err != nil {
		return ObjectTempPublicUrl{}, err
	}*/
//	return url, nil

	return generateTempUrlSig(fileDescriptor, duration), nil
}

// Only needed to generate consumer key, commented for now
/*func GetOvhConsumerKey() (*ovh.CkValidationState, error) {
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