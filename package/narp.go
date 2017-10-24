//not a regular pixel image
package narpi

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
)

type RGB8 = struct{ R, G, B uint8 }

type NotARegularPixel struct {
	HSize uint8           // horizontal flood size
	Color RGB8            // color of pixel
	VSize map[uint8]uint8 // map has cells with count of vertical pixels of the same color
	//
	//	ex.: HSize=3, VSize[2]=[2,1,2] means area of the same color:
	//		X	X	X
	//			X
	//			X
	//			X
}

func (pixel *NotARegularPixel) RunesArray() (s [][]rune) {
	hsize := pixel.HSize + 1
	s = make([][]rune, hsize)

	vsize := uint8(0)
	for i := uint8(0); i < hsize; i++ {
		if va, ok := pixel.VSize[i]; ok {
			v := va + 1
			if v > vsize {
				vsize = v
			}
		}
	}
	if vsize == 0 {
		vsize++
	}
	for i := uint8(0); i < hsize; i++ {
		s[i] = make([]rune, vsize)
	}

	for j := uint8(0); j < vsize; j++ {
		for i := uint8(0); i < hsize; i++ {
			if val, ok := pixel.VSize[i]; ok {
				val++
				if j < val {
					s[i][j] = 'o'
				} else {
					s[i][j] = '.'
				}
			} else {
				s[i][0] = 'o'
			}
		}
	}

	return s
}

func printVerticals(m map[uint8]uint8) (r string) {
	r = ""

	for i := 0; i < len(m); i++ {
		r += fmt.Sprintf("%v:%d ", i+1, m[uint8(i)]+1)
	}

	r = ""
	if r == "" {
		r = "empty"
	}
	return r
}

func (pixel *NotARegularPixel) Print(prefix string) {
	s := pixel.RunesArray()

	log.Printf("--------------------------------------------------------------------------------------")
	log.Printf("(NotARegularPixel):: %vcolor:%v, hsize:%v, size(runes, human friendly):%vx%v, \n\nvsize:  %v\nverticals(human friendly):  %v\n",
		prefix, pixel.Color, pixel.HSize, len(s), len(s[0]), pixel.VSize, printVerticals(pixel.VSize))

	for j := 0; j < len(s[0]); j++ {
		st := ""
		for i := 0; i < len(s); i++ {
			st += string(s[i][j])
		}
		log.Println(st)
	}
	log.Printf("--------------------------------------------------------------------------------------")
}

func getRGBA8(rgba16 color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := rgba16.RGBA()
	return uint8(r / 257), uint8(g / 257), uint8(b / 257)
}

func colorsEqual(img *image.RGBA, x, y int, rgb8 RGB8) bool {
	firstb := y*img.Stride + x*4

	if img.Pix[firstb] == rgb8.R && img.Pix[firstb+1] == rgb8.G && img.Pix[firstb+2] == rgb8.B {
		return true
	}
	return false
}

func (pixel *NotARegularPixel) ReadBytesBuffer(b *bytes.Buffer) error {
	*pixel = NotARegularPixel{
		HSize: 0, VSize: map[uint8]uint8{}, Color: RGB8{0, 0, 0}}
	var err error

	pixel.Color.R, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.Color.G, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.Color.B, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pixel.HSize, err = b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}

	left, err := b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}
	right, err := b.ReadByte()
	if err != nil {
		log.Fatalf(err.Error())
	}
	vsize := putBytesToUint16(left, right)

	if vsize > 0 {
		for i := byte(0); i < pixel.HSize; i++ {
			v, err := b.ReadByte()
			if err != nil {
				log.Fatalf(err.Error())
			}
			pixel.VSize[i] = v
		}

	}

	return nil
}

func (pixel *NotARegularPixel) BytesBuffer() (b *bytes.Buffer) {
	bn := new(bytes.Buffer)

	bn.WriteByte(pixel.Color.R)
	bn.WriteByte(pixel.Color.G)
	bn.WriteByte(pixel.Color.B)
	bn.WriteByte(pixel.HSize)
	bn.WriteByte(uint8(len(pixel.VSize)))
	if pixel.VSize != nil && len(pixel.VSize) > 0 {
		for i := byte(0); i < pixel.HSize; i++ {
			bn.WriteByte(pixel.VSize[i])
		}
	}

	b = new(bytes.Buffer)
	left, right := cutBytesOfUint16(uint16(bn.Len()))
	b.Write([]uint8{left, right})

	b.Write(bn.Bytes())

	return b
}
