package main

import (
	"narpimage"
	"testing"
)

func TestGetImage(t *testing.T) {
	teststring := "./testitem.jpg"

	image, err := narpimage.GetFromJpegFile(teststring)
	if err != nil {
		t.Error(err)
	}

	if image.Bounds().Max.X == 0 || image.Bounds().Max.Y == 0 {
		t.Error("Image null!")
	}
}
