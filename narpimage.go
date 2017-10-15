package NARPImage

// notaregularpixel package

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

type Point struct {
	X, Y uint16
}

type RGBAColor struct {
	R, G, B uint8
}

type NotARegularPixel struct {
	HSize uint8             // horizontal flood size
	VSize map[uint8][]uint8 // map has cells with count of vertical pixels of the same color and then the skip number, repeat
	//
	//	ex.: HSize=3, VSize[2]=[2,1,2] means area of the same color:
	//		X	X	X
	//			X
	//
	//			X
	//			X
	Color RGBAColor // color of pixel
}

type NARPImage struct {
	NARPixels map[Point]NotARegularPixel
	Size      Point
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

	/*
	   rect := img.Bounds()
	   rgba := image.NewRGBA(rect)
	   draw.Draw(rgba, rect, img, rect.Min, draw.Src)
	*/

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
	narpimage.NARPixels = make(map[Point]NotARegularPixel)
	narpimage.Size = Point{0, 0}
	narpimage.Version = "1"
}

func (narpimage *NARPImage) putToNarpImage(img image.Image, showprogress bool) error {
	if img == nil {
		return errors.New("putToNarpImage: Underlying image to construct from is nil!")
	}

	visited := make([][]bool, img.Bounds().Max.X+1)
	for k, _ := range visited {
		visited[k] = make([]bool, img.Bounds().Max.Y+1)
	}
	narpimage.Size.X, narpimage.Size.Y = uint16(img.Bounds().Max.X), uint16(img.Bounds().Max.Y)
	var boundsmin Point
	boundsmin.X, boundsmin.Y = uint16(img.Bounds().Min.X), uint16(img.Bounds().Min.Y)

	for y := boundsmin.Y; y <= narpimage.Size.Y; y++ {
		progress := float32(y) / float32(narpimage.Size.Y) * 100.0
		if showprogress {
			if int(progress*100)%10 == 0 {
				fmt.Printf("Progress: %.2f%% \r", progress)
			}
		}

		for x := boundsmin.X; x <= narpimage.Size.X; x++ {
			if !(visited[x][y]) {
				narp := getNARP(x, y, img, &visited)
				narpimage.NARPixels[Point{x, y}] = *narp
				//				if len(narp.VSize) == 0 {
				//					fmt.Println("                                 ", *narp)
				//				}

				//size := reflect.TypeOf(*narp).Size()
			}
		}
	}

	if showprogress {
		fmt.Println()
	}

	return nil
}

func getRGBAFFRange(imageClr color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := imageClr.RGBA()
	return uint8(r / 257), uint8(g / 257), uint8(b / 257)
}

func compareColor(imageClr color.Color, narpColor RGBAColor) bool {
	r, g, b := getRGBAFFRange(imageClr)
	if r == narpColor.R && g == narpColor.G && b == narpColor.B {
		return true
	}
	return false
}

func getNARP(x uint16, y uint16, img image.Image, visited *[][]bool) (narp *NotARegularPixel) {
	r, g, b := getRGBAFFRange(img.At(int(x), int(y)))
	narp = &NotARegularPixel{
		HSize: 0, VSize: map[uint8][]uint8{}, Color: RGBAColor{r, g, b}}

	for i := x; compareColor(img.At(int(i), int(y)), narp.Color) && i <= uint16(img.Bounds().Max.X); i++ {
		narp.HSize++
		(*visited)[i][y] = true
		verticals := getVerticalFloodCount(x, y, img, visited)
		if verticals != nil && len(*verticals) > 0 {
			narp.VSize[uint8(i-x)] = append(narp.VSize[uint8(i-x)], *verticals...)
		}
	}

	return narp
}

func appendToVerticals(verticals *[]uint8, count uint8) {
	if verticals == nil {
		verticals = &[]uint8{}
	}
	*verticals = append(*verticals, count)
}

func getVerticalFloodCount(x uint16, y uint16, img image.Image, visited *[][]bool) (verticals *[]uint8) {
	r, g, b := getRGBAFFRange(img.At(int(x), int(y)))
	color := RGBAColor{r, g, b}
	count := uint8(0)

	findingColor := true

	for i := y + 1; i <= uint16(img.Bounds().Max.Y); i++ {
		if findingColor == compareColor(img.At(int(x), int(i)), color) {
			count++
			(*visited)[x][i] = true
		} else {
			appendToVerticals(verticals, count)
			findingColor = !findingColor
		}
		if int(count+1) >= 256 {
			count = 0
			findingColor = !findingColor
		}
	}
	if count > 0 {
		appendToVerticals(verticals, count)
	}

	return verticals
}
