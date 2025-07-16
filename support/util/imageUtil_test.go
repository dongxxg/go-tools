package util

import (
	"fmt"
	"image"
	"testing"
)

func TestImage2Bytes(t *testing.T) {
	var canvasRectTemp = image.Rect(0, 0, 10, 10)
	var image = image.NewRGBA(canvasRectTemp)
	_, err := Image2Bytes(image)
	if err != nil {
		fmt.Printf("ERR,%v", err)
	}
}

func TestImage2Bytes2(t *testing.T) {
	var canvasRectTemp = image.Rect(10, 10, 10, 9)
	var image = image.NewRGBA(canvasRectTemp)
	_, err := Image2Bytes(image)
	if err != nil {
		fmt.Printf("ERR,%v", err)
	}

	println(image.Bounds().Dx(), image.Bounds().Dy())
}

// func TestImage2Bytes3(t *testing.T) {
// 	var canvasRectTemp = image.Rect(0, 0, 0, 0)
// 	var image = image.NewRGBA(canvasRectTemp)
// 	_, err := Image2Bytes(image)
// 	if err != nil {
// 		fmt.Printf("ERR,%v", err)
// 	}
// 	if canvasRectTemp.Bounds().Dy()
// }
