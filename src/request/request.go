package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/KJHJason/Cultured-Downloader-CLI/utils"
)

// CallRequest is used to make a request to a URL and return the response
//
// If the request fails, it will retry the request again up 
// to the defined max retries in the constants.go in utils package
func CallRequest(method, url string, timeout int, cookies []http.Cookie, additionalHeaders, params map[string]string, checkStatus bool) (*http.Response, error) {
	// sends a request to the website
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// add cookies to the request
	for _, cookie := range cookies {
		if strings.Contains(url, cookie.Domain) {
			req.AddCookie(&cookie)
		}
	}

	// add headers to the request
	for key, value := range additionalHeaders {
		req.Header.Add(key, value)
	}
	req.Header.Add(
		"User-Agent", utils.USER_AGENT,
	)

	// add params to the request
	if len(params) > 0 {
		query := req.URL.Query()
		for key, value := range params {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	// send the request
	client := &http.Client{}
	client.Timeout = time.Duration(timeout) * time.Second
	for i := 1; i <= utils.RETRY_COUNTER; i++ {
		resp, err := client.Do(req)
		if err == nil {
			if !checkStatus {
				return resp, nil
			} else if resp.StatusCode == 200 {
				return resp, nil
			}
		}
		time.Sleep(utils.GetRandomDelay())
	}
	errorMessage := fmt.Sprintf("failed to send a request to %s after %d retries", url, utils.RETRY_COUNTER)
	utils.LogError(err, errorMessage, false)
	return nil, err
}

// DownloadURL is used to download a file from a URL
//
// Note: If the file already exists, the download process will be skipped
func DownloadURL(fileURL, filePath string, cookies []http.Cookie, headers, params map[string]string) error {
	downloadTimeout := 25 * 60  // 25 minutes in seconds as downloads 
								// can take quite a while for large files (especially for Pixiv)
								// However, the average max file size on these platforms is around 300MB.
								// Note: Fantia do have a max file size per post of 3GB if one paid extra for it.
	res, err := CallRequest("GET", fileURL, downloadTimeout, cookies, headers, params, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// check if filepath already have a filename attached
	if filepath.Ext(filePath) == "" {
		os.MkdirAll(filePath, 0755)
		filename, err := url.PathUnescape(res.Request.URL.String())
		if err != nil {
			panic(err)
		}
		filename = utils.GetLastPartOfURL(filename)
		filenameWithoutExt := utils.RemoveExtFromFilename(filename)
		filePath = filepath.Join(filePath, filenameWithoutExt + strings.ToLower(filepath.Ext(filename)))
	} else {
		filePathDir := filepath.Dir(filePath)
		os.MkdirAll(filePathDir, 0755)
		filePathWithoutExt := utils.RemoveExtFromFilename(filePath)
		filePath = filePathWithoutExt + strings.ToLower(filepath.Ext(filePath))
	}

	// check if the file already exists
	if empty, _ := utils.CheckIfFileIsEmpty(filePath); !empty {
		return nil
	}

	// create the file
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	// write the body to file
	// https://stackoverflow.com/a/11693049/16377492
	_, err = io.Copy(file, res.Body)
	if err != nil {
		file.Close()
		os.Remove(filePath)
		errorMsg := fmt.Sprintf("failed to download %s due to %v", fileURL, err)
		utils.LogError(err, errorMsg, false)
		return nil
	}
	file.Close()
	return nil
}

// DownloadURLsParallel is used to download multiple files from URLs in parallel
//
// Note: If the file already exists, the download process will be skipped
func DownloadURLsParallel(urls []map[string]string, maxConcurrency int, cookies []http.Cookie, headers, params map[string]string) {
	if len(urls) < maxConcurrency {
		maxConcurrency = len(urls)
	}

	bar := utils.GetProgressBar(
		len(urls),
		"Downloading...",
		utils.GetCompletionFunc(
			fmt.Sprintf("Downloaded %d files", len(urls)),
		),
	)
	var wg sync.WaitGroup
	queue := make(chan struct{}, maxConcurrency)
	for _, url := range urls {
		wg.Add(1)
		queue <- struct{}{}
		go func(fileUrl, filePath string) {
			defer wg.Done()
			DownloadURL(fileUrl, filePath, cookies, headers, params)
			bar.Add(1)
			<-queue
		}(url["url"], url["filepath"])
	}
	close(queue)
	wg.Wait()
}