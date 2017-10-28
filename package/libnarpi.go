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
const NarpiCodecInformation = "NARPI0.75!"

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("(NARPImage):: ")
	log.SetOutput(os.Stderr)
}

func createLightBuffer(pix *[]uint8, xs int, colorindex uint8) ([]uint8, error) {
	if pix == nil {
		log.Fatalf("pix argument is nil")
	}
	var ba bytes.Buffer
	var bh bytes.Buffer

	ys := len(*pix) / 4 / xs

	//log.Println(pix)
	cts := uint8(1)
	pv := uint8((*pix)[colorindex])
	v := uint8(0)
	eq := true

	for i := 1; i < xs*ys; {
		//		log.Printf("table index=%v,pv=%v,nv=%v,cts=%v,bh=%v,eq=%v", i*4+int(colorindex), pv, (*pix)[i*4+int(colorindex)],
		//			cts, bh.Bytes(), eq)
		v = (*pix)[i*4+int(colorindex)]
		if i == 1 {
			eq = pv == v
		}
		if pv == v {
			if !eq {
				err := ba.WriteByte(127 + uint8(bh.Len()))
				if err != nil {
					return ba.Bytes(), err
				}
				_, err = ba.Write(bh.Bytes())
				if err != nil {
					return ba.Bytes(), err
				}
				bh.Reset()
				eq = true
				//				log.Printf("ba=%v,eq=%v", ba.Bytes(), eq)
			}
			cts++
		} else {
			if eq {
				err := ba.WriteByte(cts)
				if err != nil {
					return ba.Bytes(), err
				}
				cts = 1
				err = ba.WriteByte(pv)
				if err != nil {
					return ba.Bytes(), err
				}
				eq = false
			} else {
				err := bh.WriteByte(pv)
				if err != nil {
					return ba.Bytes(), err
				}
			}
		}
		pv = v
		i++
	}

	if int(cts)*(bh.Len()+1) > 1 {
		if !eq {
			err := ba.WriteByte(127 + uint8(bh.Len()) + 1)
			if err != nil {
				return ba.Bytes(), err
			}

			_, err = ba.Write(bh.Bytes())
			if err != nil {
				return ba.Bytes(), err
			}

			err = ba.WriteByte(pv)
			if err != nil {
				return ba.Bytes(), err
			}
			//log.Printf("ba=%v,eq=%v", ba.Bytes(), eq)
		} else {
			err := ba.WriteByte(cts)
			if err != nil {
				return ba.Bytes(), err
			}
			err = ba.WriteByte(pv)
			if err != nil {
				return ba.Bytes(), err
			}
			//log.Printf("ba=%v,cts=%v,eq=%v", ba.Bytes(), cts, eq)
		}
	}

	var bb bytes.Buffer
	l := uint32(ba.Len())
	//log.Println("ColorIndex write::", colorindex, l)
	err := bb.WriteByte(uint8(l >> 24))
	if err != nil {
		return bb.Bytes(), err
	}

	err = bb.WriteByte(uint8(l << 8 >> 24))
	if err != nil {
		return bb.Bytes(), err
	}

	err = bb.WriteByte(uint8(l << 16 >> 24))
	if err != nil {
		return bb.Bytes(), err
	}

	err = bb.WriteByte(uint8(l << 24 >> 24))
	if err != nil {
		return bb.Bytes(), err
	}

	//log.Println("len::", uint8(l>>24), uint8(l<<8>>24), uint8(l<<16>>24), uint8(l<<24>>24))
	//log.Println("len back::", uint32(uint8(l>>24))<<24|uint32(uint8(l<<8>>24))<<16|uint32(uint8(l<<16>>24))<<8|uint32(uint8(l<<24>>24)))
	_, err = bb.Write(ba.Bytes())
	if err != nil {
		return bb.Bytes(), err
	}

	return bb.Bytes(), nil
}

func drawLightBuffer(bI *bytes.Buffer, pix *[]uint8, colorindex uint8) {
	//log.Printf("buff=%v", bI.Bytes()[:50])

	v1, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer. " + err.Error())
	}
	v2, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer. " + err.Error())
	}
	v3, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer. " + err.Error())
	}
	v4, err := bI.ReadByte()
	if err != nil {
		log.Fatalf("Couldn't read Narpi buffer. " + err.Error())
	}

	l := uint32(v1)<<24 | uint32(v2)<<16 | uint32(v3)<<8 | uint32(v4)
	//log.Println("ColorIndex read::", colorindex, l)
	offset := uint32(0)
	for offsetbuff := uint32(0); offsetbuff < uint32(l); {
		ct, _ := bI.ReadByte()
		offsetbuff++
		if ct < 128 {
			v, _ := bI.ReadByte()
			offsetbuff++
			for j := uint32(0); j < uint32(ct); j++ {
				(*pix)[offset*4+uint32(colorindex)] = v
				(*pix)[offset*4+3] = 255
				offset++
				//log.Printf("ct=%v,v=%v, offset=%v, pix=%v", ct, v, offset, *pix)
			}
		} else {
			for j := uint32(0); j < uint32(ct-127); j++ {
				v, _ := bI.ReadByte()
				offsetbuff++
				(*pix)[offset*4+uint32(colorindex)] = v
				(*pix)[offset*4+3] = 255
				offset++
				//log.Printf("ct=%v,v=%v, offset=%v, pix=%v", ct, v, offset, *pix)
			}
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
	drawLightBuffer(&bI, &(img.Pix), 0)
	drawLightBuffer(&bI, &(img.Pix), 1)
	drawLightBuffer(&bI, &(img.Pix), 2)

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

	fileOut.Close()
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

	fileOut.Close()
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
	var bt []byte

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

	bt, err = createLightBuffer(&(img.Pix), img.Stride, 0)
	ba.Write(bt)
	if err != nil {
		return err
	}

	bt, err = createLightBuffer(&(img.Pix), img.Stride, 1)
	ba.Write(bt)
	if err != nil {
		return err
	}

	bt, err = createLightBuffer(&(img.Pix), img.Stride, 2)
	if err != nil {
		return err
	}
	_, err = ba.Write(bt)
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
