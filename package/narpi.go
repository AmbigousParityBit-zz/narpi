// not a regular pixels image package
package narpi

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type NARPImage struct {
	NARPixels []NotARegularPixel
	Size      struct{ X, Y uint16 }
	Codec     string
	//Colors    map[RGB8]rune
}

func (narpimage *NARPImage) init() {
	narpimage.NARPixels = []NotARegularPixel{}
	narpimage.Size = struct{ X, Y uint16 }{0, 0}
	narpimage.Codec = "NARPI0.6"
	//narpimage.Colors = map[RGB8]rune{}
}

func (narpimage *NARPImage) rgbaImage() (img *image.RGBA, err error) {
	defer timeTrack(time.Now(), "rgbaImage")
	img = image.NewRGBA(image.Rect(0, 0, int(narpimage.Size.X), int(narpimage.Size.Y)))
	var visited [][]bool
	lenvis := 0

	x, y := uint16(0), uint16(0)

	for _, v := range narpimage.NARPixels {
		v.drawNARP(img, int(x), int(y))
		v.markVisited(int(x), int(y), &visited, int(narpimage.Size.X), int(narpimage.Size.Y), &lenvis)
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

func (narpimage *NARPImage) BytesBuffer() (b *bytes.Buffer) {
	b = new(bytes.Buffer)

	b.WriteString(narpimage.Codec)
	b.WriteRune('!')

	left, right := cutBytesOfUint16(uint16(narpimage.Size.X))
	b.Write([]uint8{left, right})

	left, right = cutBytesOfUint16(uint16(narpimage.Size.Y))
	b.Write([]uint8{left, right})

	for _, v := range narpimage.NARPixels {
		b.Write(v.BytesBuffer().Bytes())
	}

	return b
}

func (narpimage *NARPImage) ReadBytesBuffer(b *bytes.Buffer) error {
	narpimage.init()

	var err error
	narpimage.Codec, err = b.ReadString(uint8('!'))
	if err != nil {
		log.Fatalf(err.Error())
	}

	v, err := b.ReadByte()
	narpimage.Size.X = uint16(v)
	if err != nil {
		log.Fatalf(err.Error())
	}
	v, err = b.ReadByte()
	narpimage.Size.Y = uint16(v)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return nil
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

func loadNotNarpi(filename string) (*image.RGBA, string, error) {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	imgd, cinf, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err, cinf)
	}

	width := imgd.Bounds().Dx()
	height := imgd.Bounds().Dy()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), imgd, img.Bounds().Min, draw.Src)

	return img, cinf, nil
}

func (narpimage *NARPImage) constructFromNotNarpi(filename string, showprogress bool) error {
	img, cinf, err := loadNotNarpi(filename)
	if err != nil {
		log.Fatal(err, cinf)
	}
	narpimage.init()
	narpimage.putToNarpImage(img)

	return nil
}

func (narpimage *NARPImage) Load(filename string, showprogress bool) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	narpimage.init()

	if filepath.Ext(filename) == FileExt {
		err = gob.NewDecoder(file).Decode(narpimage)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		narpimage.constructFromNotNarpi(filename, showprogress)
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
	lenvis := 0
	if narpimage.NARPixels == nil {
		s = "nil"
	}
	s = strconv.Itoa(len(narpimage.NARPixels))

	log.Printf("===========================================================================================")
	log.Printf("(NotARegularPixelImage):: size: %v, codec information: %v, pixels (#=%v):", narpimage.Size, narpimage.Codec, s)
	log.Printf("===========================================================================================")

	if narpimage.Size.X == 0 || narpimage.Size.Y == 0 {
		return
	}

	x, y := 0, 0
	for _, v := range narpimage.NARPixels {
		v.Print(fmt.Sprintf("(x:%v,y:%v) ", x+1, y+1))
		v.markVisited(x, y, &visited, int(narpimage.Size.X), int(narpimage.Size.Y), &lenvis)
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
