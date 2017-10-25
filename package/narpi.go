// Not A Regular Pixels image package.
// The aim is to create lossless format strictly for photos, which could be used to shrink filesizes. Work in progress.
package narpi

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
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

type nImage []byte

func (narpimage *NARPImage) init() {
	narpimage.NARPixels = []NotARegularPixel{}
	narpimage.Size = struct{ X, Y uint16 }{0, 0}
	narpimage.Codec = "NARPI0.6!"
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

	left, err := b.ReadByte()
	if err != nil {
		return err
	}
	right, err := b.ReadByte()
	if err != nil {
		return err
	}
	narpimage.Size.X = putBytesToUint16(left, right)

	left, err = b.ReadByte()
	if err != nil {
		return err
	}
	right, err = b.ReadByte()
	if err != nil {
		return err
	}
	narpimage.Size.Y = putBytesToUint16(left, right)

	narpimage.NARPixels = make([]NotARegularPixel, 0, int(narpimage.Size.X)*int(narpimage.Size.Y))
	for err == nil {
		pixel := NotARegularPixel{}
		err = pixel.ReadBytesBuffer(b)
		if err == nil {
			narpimage.NARPixels = append(narpimage.NARPixels, pixel)
		}
	}
	if err != io.EOF {
		log.Fatalf(err.Error())
	}

	return nil
}

func createBuffer(imgp *image.RGBA, filename string) {
	img := (*imgp)
	xs := img.Bounds().Dx()
	ys := img.Bounds().Dy()

	var bR bytes.Buffer
	var bG bytes.Buffer
	var bB bytes.Buffer
	var ba bytes.Buffer

	ctR := uint8(0)
	ctG := uint8(0)
	ctB := uint8(0)
	vR := uint8(img.Pix[0])
	vG := uint8(img.Pix[1])
	vB := uint8(img.Pix[2])

	for i := 0; i < xs*ys/4; i++ {
		if img.Pix[i*4] == vR {
			ctR++
		} else {
			bR.WriteByte(ctR)
			bR.WriteByte(vR)
			vR = uint8(img.Pix[i*4])
			ctR = 0
		}
	}
	if ctR != 0 {
		bR.WriteByte(ctR)
		bR.WriteByte(vR)
	}

	for i := 0; i < xs*ys/4; i++ {
		if img.Pix[i*4+1] == vG {
			ctG++
		} else {
			bG.WriteByte(ctG)
			bG.WriteByte(vG)
			vG = uint8(img.Pix[i*4+1])
			ctG = 0
		}
	}
	if ctG != 0 {
		bR.WriteByte(ctG)
		bR.WriteByte(vG)
	}

	for i := 0; i < xs*ys/4; i++ {
		if img.Pix[i*4+2] == vB {
			ctB++
		} else {
			bB.WriteByte(ctB)
			bB.WriteByte(vB)
			vB = uint8(img.Pix[i*4+2])
			ctB = 0
		}
	}
	if ctG != 0 {
		bR.WriteByte(ctG)
		bR.WriteByte(vG)
	}

	ba.WriteByte(uint8(bR.Len()))
	ba.Write(bR.Bytes())
	ba.WriteByte(uint8(bG.Len()))
	ba.Write(bG.Bytes())
	ba.WriteByte(uint8(bB.Len()))
	ba.Write(bB.Bytes())

	//log.Println(bR)
	log.Println(bR.Len())
	err := ioutil.WriteFile(filename+".raw1", ba.Bytes(), 0666)
	if err != nil {
		return
	}

	// ----------------

}

func getInfo(imgp *image.RGBA, filename string) error {
	img := (*imgp)
	xs := img.Bounds().Dx()
	ys := img.Bounds().Dy()

	err := ioutil.WriteFile(filename+".raw", img.Pix, 0666)
	if err != nil {
		return err
	}

	colors := map[RGB8]int{}
	for i := 0; i < xs*ys; i += 4 {
		r, g, b := img.Pix[i], img.Pix[i+1], img.Pix[i+2]
		color := RGB8{r, g, b}
		if _, b := colors[color]; b {
			colors[color]++
		} else {
			colors[color] = 1
		}
	}

	log.Printf("Colors: %v, size: %vx%v", len(colors), xs, ys)
	counter := 0
	for i, v := range colors {
		counter++
		if v > 200 {
			log.Printf("%v\t::: [%v]=\t%v", counter, i, v)
		}
	}
	createBuffer(imgp, filename)
	log.Fatalf("")

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

	err = getInfo(img, filename)

	return img, cinf, nil
}

func (narpimage *NARPImage) constructFromNotNarpi(filename string) error {
	img, cinf, err := loadNotNarpi(filename)
	if err != nil {
		log.Fatal(err, cinf)
	}
	narpimage.init()
	narpimage.putToNarpImage(img)

	return nil
}

func (narpimage *NARPImage) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	narpimage.init()

	if filepath.Ext(filename) == FileExt {
		//err = gob.NewDecoder(file).Decode(narpimage)
		var b bytes.Buffer
		_, err := b.ReadFrom(file)
		if err != nil {
			log.Println(err)
			return err
		}
		err = narpimage.ReadBytesBuffer(&b)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		narpimage.constructFromNotNarpi(filename)
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

	/*
		b := new(bytes.Buffer)
		err = gob.NewEncoder(b).Encode(narpimage)
		if err != nil {
			log.Println(err)
			return err
		}*/
	//file.Write(b.Bytes())
	file.Write(narpimage.BytesBuffer().Bytes())

	return err
}

func (narpimage *NARPImage) Print(detailed bool) {
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

	if detailed {
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
}
