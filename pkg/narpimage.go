// notaregularpixel package
package NARPImage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strconv"
	"time"
)

type NARPImage struct {
	NARPixels []NotARegularPixel
	Size      struct{ X, Y uint16 }
	Version   string
}

func (narpimage *NARPImage) rgbaImage() (img *image.RGBA, err error) {
	defer timeTrack(time.Now(), "rgbaImage")
	img = image.NewRGBA(image.Rect(0, 0, int(narpimage.Size.X), int(narpimage.Size.Y)))
	var visited [][]bool

	x, y := uint16(0), uint16(0)

	for _, v := range narpimage.NARPixels {
		v.drawNARP(img, int(x), int(y))
		v.markVisited(int(x), int(y), &visited, int(narpimage.Size.X), int(narpimage.Size.Y))
		end := false

		for !end && visited[x][y] {
			x++
			if x >= narpimage.Size.X {
				x = 0
				y++
				if y >= narpimage.Size.Y {
					end = true
				}
			}
		}
	}
	return img, nil
}

func (narpimage *NARPImage) Png(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := narpimage.rgbaImage()
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	if err != nil {
		return err
	}

	return nil
}

func (narpimage *NARPImage) Jpg(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := narpimage.rgbaImage()
	if err != nil {
		return err
	}

	opt := jpeg.Options{Quality: 100}
	err = jpeg.Encode(f, img, &opt)
	if err != nil {
		return err
	}

	return nil
}

func (narpimage *NARPImage) LoadPng(filename string, showprogress bool) error {
	reader, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer reader.Close()

	img, err := png.Decode(reader)
	if err != nil {
		log.Println(err)
		return err
	}

	narpimage.initNARPImage()
	narpimage.putToNarpImage(img, showprogress)

	return nil
}

func (narpimage *NARPImage) LoadJpg(filename string, showprogress bool) error {
	reader, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer reader.Close()

	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Println(err)
		return err
	}

	narpimage.initNARPImage()
	narpimage.putToNarpImage(img, showprogress)

	return nil
}

func (narpimage *NARPImage) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	narpimage.initNARPImage()

	err = gob.NewDecoder(file).Decode(narpimage)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

func (narpimage *NARPImage) Save(filename string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			log.Fatalf("Save: error, file <%s> already exists", filename)
			return err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	b := new(bytes.Buffer)
	err = gob.NewEncoder(b).Encode(narpimage)
	if err != nil {
		log.Println(err)
		return err
	}
	file.Write(b.Bytes())

	return err
}

func (narpimage *NARPImage) Print() {
	s := ""
	var visited [][]bool
	if narpimage.NARPixels == nil {
		s = "nil"
	}
	s = strconv.Itoa(len(narpimage.NARPixels))

	log.Printf("===========================================================================================")
	log.Printf("(NotARegularPixelImage):: size: %v, codec version: %v, pixels (#=%v):", narpimage.Size, narpimage.Version, s)
	log.Printf("===========================================================================================")

	if narpimage.Size.X == 0 || narpimage.Size.Y == 0 {
		return
	}

	x, y := 0, 0
	for _, v := range narpimage.NARPixels {
		v.Print(fmt.Sprintf("(x:%v,y:%v) ", x+1, y+1))
		v.markVisited(x, y, &visited, int(narpimage.Size.X), int(narpimage.Size.Y))
		end := false

		for !end && visited[x][y] {
			x++
			if x >= int(narpimage.Size.X) {
				x = 0
				y++
				if y >= int(narpimage.Size.Y) {
					end = true
				}
			}
		}
	}
}
