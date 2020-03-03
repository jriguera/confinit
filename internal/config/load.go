package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"

	"gopkg.in/yaml.v2"
)

func LoadResource(resource string) (doc interface{}, err error) {
	var result interface{}
	if ValidUrl(resource) {
		result, err = GetHttpJSON(resource)
	} else {
		exist, filetype := ValidFile(resource)
		if !exist {
			err = fmt.Errorf("File '%s' not found", resource)
			return
		}
		if filetype == "yaml" || filetype == "yml" {
			result, err = FileYAML(resource)
		} else if filetype == "json" {
			result, err = FileJSON(resource)
		} else {
			err = fmt.Errorf("File extension '%s' not supported (not Json or Yaml)", resource)
		}
	}
	if err != nil {
		return
	}
	doc = MapI2S(result)
	return
}

func ValidUrl(testurl string) bool {
	_, err := url.ParseRequestURI(testurl)
	if err != nil {
		return false
	}
	u, err := url.Parse(testurl)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func ValidFile(testfile string) (exist bool, ext string) {
	exist = false
	if _, err := os.Stat(testfile); os.IsNotExist(err) {
		return
	}
	ext = strings.ToLower(filepath.Ext(filepath.Base(testfile))[1:])
	return true, ext
}

func GetHttpJSON(url string) (result interface{}, err error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    time.Second * 30,
		DisableCompression: false,
	}
	client := http.Client{
		Transport: tr,
		Timeout:   time.Second * 30,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", ConfigUserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 || res.StatusCode == 202 || res.StatusCode == 204 {
		body, errRead := ioutil.ReadAll(res.Body)
		if errRead != nil {
			err = errRead
			return
		} else {
			err = json.Unmarshal([]byte(body), &result)
		}
	} else {
		err = fmt.Errorf("HTTP status code %s: %v ", res.Status, res.Body)
	}
	return
}

func FileJSON(filename string) (result interface{}, err error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return
	}
	defer jsonFile.Close()
	content, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(content), &result)
	return
}

func FileYAML(filename string) (result interface{}, err error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return
	}
	defer jsonFile.Close()
	content, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal([]byte(content), &result)
	return
}

// From https://github.com/icza/dyno
// Package dyno is a utility to work with dynamic objects at ease.
// Apache license 2.0
//
// MapI2S walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func MapI2S(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string:
				m[k2] = MapI2S(v2)
			default:
				m[fmt.Sprint(k)] = MapI2S(v2)
			}
		}
		v = m
	case []interface{}:
		for i, v2 := range x {
			x[i] = MapI2S(v2)
		}
	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = MapI2S(v2)
		}
	}
	return v
}
