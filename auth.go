package s3gof3r

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
)

// Keys for an Amazon Web Services account.
// Used for signing http requests.
type Keys struct {
	AccessKey     string
	SecretKey     string
	SecurityToken string
}

type mdCreds struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

func InstanceKeys() (keys Keys, err error) {

	rolePath := "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
	var creds mdCreds

	// request the role name for the instance
	// assumes there is only one
	resp, err := http.Get(rolePath)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = newRespError(resp)
		return
	}
	role, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	// request the credential metadata for the role
	resp, err = http.Get(rolePath + string(role))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = newRespError(resp)
		return
	}
	metadata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	json.Unmarshal([]byte(metadata), &creds)
	keys = Keys{AccessKey: creds.AccessKeyId,
		SecretKey:     creds.SecretAccessKey,
		SecurityToken: creds.Token,
	}

	return
}

// Uses same environment variables as aws cli
func EnvKeys() (keys Keys, err error) {
	keys = Keys{AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	if keys.AccessKey == "" || keys.SecretKey == "" {
		err = errors.New("AWS keys not in environment.")
	}
	return
}

// This convenience function gets the AWS Keys from environment variables or the instance-based metadata on EC2
// Environment variables are attempted first, followed by the instance-based credentials.
// Assumes only one IAM role
// It returns an error if no keys are found.
func GetAWSKeys() (keys Keys, err error) {

	keys, err = EnvKeys()
	if err == nil {
		return
	}
	keys, err = InstanceKeys()
	if err == nil {
		return
	}
	err = errors.New("No AWS Keys found.")
	return
}
