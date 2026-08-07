package util

import (
	"bytes"
	logger "unitechs.com/unios-dice/uni-base/core/log"
	"image"
	"image/png"
)

func Image2Bytes(m image.Image) (data []byte, errinfo error) {
	b := new(bytes.Buffer)
	err := png.Encode(b, m)
	if err != nil {
		log.Error("convert image to bytes ,error: %s", err)
		return []byte(""), err
	}
	return b.Bytes(), nil
}
