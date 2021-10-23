package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var headers = make(map[string]string)

func init() {
	headers = map[string]string{"Cookie": "69tikiman69", "User-Agent": "BOBKILL"}
}

// Returns empty string if the message matches,
// https://vm.tiktok.com/ZM8Fs6RR6/
func isShortTiktokUrl(message string) string {
	reg, _ := regexp.Compile(`https:\/\/vm\.tiktok\.com\/(.*)\/`)
	return reg.FindString(message)
}

// URL's that look like https://www.tiktok.com/@yannbernillie/video/7013334258051845381
func isLongTiktokUrl(message string) string {
	reg, _ := regexp.Compile(`https://w{3}\.tiktok\.com/@.*/\d*`)
	return reg.FindString(message)
}

func isTiktokUrl(message string, client *http.Client) (string, error) {
	if url := isShortTiktokUrl(message); url != "" { // check if url is of type vm.tiktok
		longUrl, err := transformShortUrl(url, client)
		if err != nil {
			return "", err
		}
		return longUrl, err
	}
	return isLongTiktokUrl(message), nil
}

// Takes vm.tiktok urls and converts them to long url's for video grabbing
func transformShortUrl(shortURL string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", shortURL, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	longUrl := strings.Split(res.Request.URL.String(), "?")[0]
	return longUrl, nil
}

func getTiktokVideoURL(longURL string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", longURL, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", headers["User-Agent"])
	req.Header.Set("Cookie", headers["Cookie"])
	req.Header.Set("Referer", headers["Referer"])

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	// Actual video URL from CDN somewhere in the body
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	// some bunk regex to find the video url
	bodyReplaced := strings.ReplaceAll(string(body), "amp;", "")
	videoUrlRegEx, err := regexp.Compile(`content="(https://[^w-w]{3}(.*?)tiktok\.com/video/tos(.*?)&vr=)`)
	if err != nil {
		return "", err
	}

	videoURLMatches := videoUrlRegEx.FindAllStringSubmatch(bodyReplaced, 1)
	if len(videoURLMatches) == 0 {
		err = fmt.Errorf("no matches for videoUrlRegex: %v", videoURLMatches)
		return "", err
	}
	videoURL := videoURLMatches[0][1]
	return videoURL, nil
}

// Returns bytes of Tiktok video given tiktok url of type longURL
func getTiktokVideo(url string, client *http.Client) ([]byte, error) {
	videoUrl, err := getTiktokVideoURL(url, client)
	if err != nil {
		return nil, fmt.Errorf("error getting tiktok video URL: %v", err)
	}

	req, err := http.NewRequest("GET", videoUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", headers["User-Agent"])
	req.Header.Set("Cookie", headers["Cookie"])
	req.Header.Set("Referer", headers["Referer"])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fileBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading tiktok video bytes: %v", err)
	}
	return fileBytes, err
}
