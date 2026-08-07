package httpx

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"unitechs.com/unios-dice/uni-base/core/log"
)

// Down2File downloads the content at url and writes it to localFile.
func Down2File(url, localFile string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return down2File(url, localFile, timeout)
}

// Down2Memory downloads the content at url and returns its bytes plus a
// best-effort file suffix derived from the response's Content-Type / Content-Disposition.
func Down2Memory(url string, timeout time.Duration) ([]byte, string, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return down2Memory(url, timeout)
}

func down2File(url, localFile string, timeout time.Duration) error {
	logPrefix := "Down2File " + url + " to " + localFile
	log.Debug("%s begin...", logPrefix)
	beginTime := time.Now()
	defer func() {
		log.Debug("%s end and duration: %s", logPrefix, time.Since(beginTime))
	}()

	res, err := fetch(url, timeout)
	if err != nil {
		log.Debug("%s fetch failed as: %s", logPrefix, err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("StatusCode: %d", res.StatusCode)
		log.Debug("%s failed as %s", logPrefix, err)
		return err
	}
	if res.Header.Get("Content-Type") == contentTypeJSON {
		return fmt.Errorf("response content type is application/json")
	}

	if dir := filepath.Dir(localFile); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Debug("%s create dir failed as: %s", logPrefix, err)
			return err
		}
	}
	out, err := os.Create(localFile)
	if err != nil {
		log.Debug("%s create localFile failed as: %s", logPrefix, err)
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, res.Body); err != nil {
		log.Debug("%s copy to localFile failed as: %s", logPrefix, err)
		return err
	}
	return nil
}

func down2Memory(url string, timeout time.Duration) ([]byte, string, error) {
	logPrefix := "Down2Memory " + url
	log.Debug("%s begin...", logPrefix)
	beginTime := time.Now()
	defer func() {
		log.Debug("%s end and duration: %s", logPrefix, time.Since(beginTime))
	}()

	res, err := fetch(url, timeout)
	if err != nil {
		log.Debug("%s fetch failed as: %s", logPrefix, err)
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("StatusCode: %d", res.StatusCode)
		content, _ := io.ReadAll(res.Body)
		log.Debug("%s failed, err: %v\ncontent: %s", logPrefix, err, content)
		return content, "", err
	}
	if res.Header.Get("Content-Type") == contentTypeJSON {
		return nil, "", fmt.Errorf("response content type is application/json")
	}

	suffix := suffixByResponse(res)

	content, err := io.ReadAll(res.Body)
	if err != nil {
		log.Debug("%s read body failed as: %s", logPrefix, err)
		return nil, "", err
	}
	return content, suffix, nil
}

func fetch(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// suffixByResponse derives a file suffix from the response's Content-Type or
// Content-Disposition header.
// Reference: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
func suffixByResponse(res *http.Response) string {
	suffix := ""
	contentType := res.Header.Get("content-type")
	log.Debug("ContentType is %s while download file", contentType)
	if len(contentType) == 0 {
		log.Warn("content-type is empty while download file")
	} else {
		switch contentType {
		case "application/vnd.ms-powerpoint":
			suffix = ".ppt"
		case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
			suffix = ".pptx"
		case "application/pdf":
			suffix = ".pdf"
		case "application/msword":
			suffix = ".doc"
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			suffix = ".docx"
		case "application/vnd.ms-excel":
			suffix = ".xls"
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
			suffix = ".xlsx"
		case "video/mp4", "video/mpeg":
			suffix = ".mp4"
		case "audio/mpeg":
			suffix = ".mp3"
		case "image/jpeg":
			suffix = ".jpg"
		case "image/png":
			suffix = ".png"
		case "image/x-icon":
			suffix = ".ico"
		case "application/json":
			suffix = ".json"
		default:
			log.Warn("unknown content-type %s while download file", contentType)
		}
	}

	if len(suffix) > 0 {
		log.Debug("get file suffix(%s) by content-type(%s)", suffix, contentType)
		return suffix
	}

	contentDisposition := res.Header.Get("content-disposition")
	log.Debug("content-disposition is %s while download file", contentDisposition)
	if len(contentDisposition) > 0 {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			log.Debug("parse content-disposition and params is %s", params)
			if filename := params["filename"]; len(filename) > 0 {
				suffix = path.Ext(filename)
			} else if filename := params["filename*"]; len(filename) > 0 {
				suffix = path.Ext(filename)
			}
		} else {
			log.Debug("parse content-disposition fail as err is %s", err)
		}
	}

	return suffix
}
