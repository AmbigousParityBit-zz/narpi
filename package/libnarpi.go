// Not A Regular Pixels image package.
// The aim is to create lossless format strictly for photos, which could be used to shrink filesizes. Work in progress.
package libnarpi

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const NarpiFileExt = ".narpi"
const NarpiCodecInformation = "NARPI0.6!"

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("(NARPImage):: ")
	log.SetOutput(os.Stderr)
}

func createLightBuffer(imgp *image.RGBA, colorindex uint8) (bytes.Buffer, error) {
	var ba bytes.Buffer
	var bb bytes.Buffer
	img := (*imgp)
	xs := img.Bounds().Dx()
	ys := img.Bounds().Dy()

	ct := uint8(0)
	v := uint8(img.Pix[0])

	for i := 0; i < xs*ys; {
		if img.Pix[i*4+int(colorindex)] == v {
			ct++
			i++
		} else {
			err := ba.WriteByte(ct)
			if err != nil {
				return bb, err
			}
			err = ba.WriteByte(v)
			if err != nil {
				return bb, err
			}
			v = uint8(img.Pix[i*4+int(colorindex)])
			ct = 0
		}
	}
	if ct != 0 {
		err := ba.WriteByte(ct)
		if err != nil {
			return bb, err
		}
		err = ba.WriteByte(v)
		if err != nil {
			return bb, err
		}
	}

	l := uint32(ba.Len())
	log.Println("ColorIndex write::", colorindex, l)
	err := bb.WriteByte(uint8(l >> 24))
	if err != nil {
		return bb, err
	}

	err = bb.WriteByte(uint8(l << 8 >> 24))
	if err != nil {
		return bb, err
	}

	err = bb.WriteByte(uint8(l << 16 >> 24))
	if err != nil {
		return bb, err
	}

	err = bb.WriteByte(uint8(l << 24 >> 24))
	if err != nil {
		return bb, err
	}

	_, err = bb.Write(ba.Bytes())
	if err != nil {
		return bb, err
	}

	return bb, nil
}

func drawLightBuffer(bI *bytes.Buffer, imgp *image.RGBA, colorindex uint8) {
	img := *imgp

	v1, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	v2, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	v3, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	v4, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}

	l := uint32(v1)<<24 | uint32(v2)<<16 | uint32(v3)<<8 | uint32(v4)
	log.Println("ColorIndex read::", colorindex, l)
	offset := uint32(0)
	for i := uint32(0); i < uint32(l/2); i++ {
		ct, _ := bI.ReadByte()
		v, _ := bI.ReadByte()
		for j := uint32(0); j < uint32(ct); j++ {
			img.Pix[offset*4+uint32(colorindex)] = v
			img.Pix[offset*4+3] = 255
			offset++
		}
	}
}

func getRGBA(filenameIn, filenameOut string, overwrite bool) (*image.RGBA, *os.File, error) {
	if !overwrite {
		if _, err := os.Stat(filenameOut); !os.IsNotExist(err) {
			log.Fatalf("Save: error, file <%s> already exists", filenameOut)
			return nil, nil, err
		}
	}

	fileOut, err := os.OpenFile(filenameOut, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, nil, err
	}
	//defer fileOut.Close()

	var bI bytes.Buffer
	fileIn, err := os.OpenFile(filenameIn, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer fileIn.Close()

	_, err = bI.ReadFrom(fileIn)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	cinfo, err := bI.ReadString('!')
	if cinfo != NarpiCodecInformation || err != nil {
		log.Fatalf("Given file <%s> is not Narpi type (returned <%s> from header).", filepath.Base(filenameIn), cinfo)
	}

	v1, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	v2, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	xs := uint16(v1)<<8 | uint16(v2)

	v1, err = bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	v2, err = bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer.")
	}
	ys := uint16(v1)<<8 | uint16(v2)

	img := image.NewRGBA(image.Rect(0, 0, int(xs), int(ys)))
	drawLightBuffer(&bI, img, 0)
	drawLightBuffer(&bI, img, 1)
	drawLightBuffer(&bI, img, 2)

	return img, fileOut, nil
}

func Png(filenameIn string, filenameOut string, overwrite bool) error {
	img, fileOut, err := getRGBA(filenameIn, filenameOut, overwrite)
	if err != nil {
		return err
	}
	err = png.Encode(fileOut, img)
	if err != nil {
		return err
	}

	return nil
}

func Jpg(filenameIn string, filenameOut string, overwrite bool) error {
	img, fileOut, err := getRGBA(filenameIn, filenameOut, overwrite)
	if err != nil {
		return err
	}
	opt := jpeg.Options{Quality: 100}
	err = jpeg.Encode(fileOut, img, &opt)
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

func Narpi(filenameIn string, filenameOut string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(filenameOut); !os.IsNotExist(err) {
			log.Fatalf("Save: error, file <%s> already exists", filenameOut)
			return err
		}
	}

	img, _, err := loadNotNarpi(filenameIn)

	var ba bytes.Buffer
	var bt bytes.Buffer

	_, err = ba.WriteString(NarpiCodecInformation)
	if err != nil {
		return err
	}

	xs := uint16(img.Bounds().Max.X)
	ys := uint16(img.Bounds().Max.Y)
	err = ba.WriteByte(uint8(xs >> 8))
	if err != nil {
		return err
	}
	err = ba.WriteByte(uint8(xs << 8 >> 8))
	if err != nil {
		return err
	}

	err = ba.WriteByte(uint8(ys >> 8))
	if err != nil {
		return err
	}
	err = ba.WriteByte(uint8(ys << 8 >> 8))
	if err != nil {
		return err
	}

	bt, err = createLightBuffer(img, 0)
	ba.Write(bt.Bytes())
	if err != nil {
		return err
	}
	bt, err = createLightBuffer(img, 1)
	ba.Write(bt.Bytes())
	if err != nil {
		return err
	}

	bt, err = createLightBuffer(img, 2)
	if err != nil {
		return err
	}
	_, err = ba.Write(bt.Bytes())
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filenameOut, ba.Bytes(), 0666)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filenameOut+".raw", img.Pix, 0666)
	if err != nil {
		return err
	}

	return nil
}
