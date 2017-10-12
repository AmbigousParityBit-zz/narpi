package narpimage

// notaregularpixel package

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"reflect"
)

type RGBAColor struct {
	R, G, B, A uint32
}

type NotARegularPixel struct {
	Mode int
	// 1: horizontal line
	// 2: vertical line
	Size  int       // size of pixel
	Color RGBAColor // color of pixel
	// coordinates of (in case of):
	// left (horizontal line)
	// up (vertical line)
}

type NARPImage struct {
	NARPixels map[image.Point]NotARegularPixel
	Size      image.Point
	Version   string
}

func (narpimage *NARPImage) ConstructFromJpgFile(s string, showprogress bool) error {
	reader, err := os.Open(s)
	defer reader.Close()
	if err != nil {
		return err
	}

	img, err := jpeg.Decode(reader)
	if err != nil {
		return err
	}

	narpimage.initialize()
	narpimage.putToNarpImage(img, showprogress)

	return nil
}

func (narpimage *NARPImage) Load(s string) error {
	file, err := os.Open(s)
	defer file.Close()
	if err != nil {
		return err
	}

	narpimage.initialize()

	err = gob.NewDecoder(file).Decode(narpimage)

	return err
}

func (narpimage *NARPImage) Save(s string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(s); !os.IsNotExist(err) {
			return fmt.Errorf("Save: error, file <%s> already exists!", s)
		}
	}

	file, err := os.Create(s)
	defer file.Close()
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)
	err = gob.NewEncoder(b).Encode(narpimage)
	if err != nil {
		return err
	}
	file.Write(b.Bytes())

	return err
}

func (narpimage *NARPImage) initialize() {
	narpimage.NARPixels = make(map[image.Point]NotARegularPixel)
	narpimage.Size = image.Point{0, 0}
	narpimage.Version = "1"
}

func (narpimage *NARPImage) putToNarpImage(img image.Image, showprogress bool) error {
	if img == nil {
		return errors.New("putToNarpImage: Underlying image to construct from is nil!")
	}

	narpimage.Size = img.Bounds().Max

	boundsmin := img.Bounds().Min

	for y := boundsmin.Y; y <= narpimage.Size.Y; y++ {
		progress := float32(y) / float32(narpimage.Size.Y) * 100.0
		if showprogress {
			if int(progress*100)%10 == 0 {
				fmt.Printf("Progress: %.2f%% \r", progress)
			}
		}

		for x := boundsmin.X; x <= narpimage.Size.X; x++ {
			narp := getNARP(x, y, img)
			narpimage.NARPixels[image.Point{x, y}] = *narp
			x = x + narp.Size
		}
	}

	if showprogress {
		fmt.Println()
	}

	return nil
}

func getNARP(x int, y int, img image.Image) (narp *NotARegularPixel) {
	r, g, b, a := img.At(x, y).RGBA()
	narp = &NotARegularPixel{
		Mode: 1, Size: 0, Color: RGBAColor{r, g, b, a}}
	for x := x; reflect.DeepEqual(img.At(x, y), narp.Color) && x <= img.Bounds().Max.X; x++ {
		narp.Size++
	}

	return narp
}
