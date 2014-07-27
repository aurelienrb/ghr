package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	RELEASE_URL = "https://api.github.com/repos/%s/%s/releases"
	UPLOAD_URL  = "https://uploads.github.com/repos/%s/%s/releases/%s/assets"
)

type Releases struct {
	ID      int    `json:"id"`
	TagName string `json:"tag_name"`
}

type ReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

func debugResponseBody(body io.ReadCloser) {
	if os.Getenv("DEBUG") != "" {
		body, _ := ioutil.ReadAll(body)
		log.Println(string(body))
	}
}

func GetReleaseID(info Info) (int, error) {

	url := fmt.Sprintf(RELEASE_URL, info.OwnerName, info.RepoName)
	debug(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	debug(res.Status)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	debug(string(body))

	var releases []Releases
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return -1, err
	}

	for _, release := range releases {
		if release.TagName == info.TagName {
			return release.ID, nil
		}
	}

	return -1, nil
}

func CreateNewRelease(info Info) error {

	url := fmt.Sprintf(RELEASE_URL, info.OwnerName, info.RepoName)
	debug(url)

	params := ReleaseRequest{
		TagName:         info.TagName,
		TargetCommitish: info.TargetCommitish,
		Draft:           info.Draft,
		Prerelease:      info.Prerelease,
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return err
	}
	debug(string(payload))

	reader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", info.Token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	debug(res.Status)
	debugResponseBody(res.Body)

	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == 422 {
			return fmt.Errorf("Github returned %s (this is probably because the release already exists)", res.Status)
		}
		return fmt.Errorf("Github returned %s", res.Status)
	}
	return nil
}