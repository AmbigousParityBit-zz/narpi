package libnarpi

import (
	"bytes"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func getTestImagesFilenames(t *testing.T, ext string) []string {
	//defer timeTrack(time.Now(), "getTestImagesFilenames")
	filenames, err := filepath.Glob("./testimages/*." + ext)
	if err != nil {
		t.Fatal(err)
	}

	fileNames := []string{}
	for _, f := range filenames {
		filename, _ := filepath.Abs(f)
		extension := path.Ext(filename)
		if extension == "."+ext {
			filename := filename[0:len(filename)-len(extension)] + "."
			fileNames = append(fileNames, filename)
		}
	}
	return fileNames
}

func deleteExtFiles(t *testing.T, ext string) {
	fileNames := getTestImagesFilenames(t, ext)
	for _, s := range fileNames {
		err := os.Remove(s + ext)
		if err != nil {
			t.Fatalf("couldn't delete file %v", s+ext)
		}
	}
	fileNames = getTestImagesFilenames(t, ext)
	if len(fileNames) > 0 {
		t.Fatalf("couldn't delete files <%v> during preparations to tests", ext)
	}
	log.Printf("Successfully deleted all <%v> files during preparations to tests.", ext)
}

func getFileSize(filename string) int64 {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	return stat.Size()
}

func compareJpgPngFiles(t *testing.T, prfx string) {
	ext1, ext2 := "png", "jpg"

	_, fn1 := filepath.Split(prfx + ext1)
	_, fn2 := filepath.Split(prfx + ext2)

	if _, err := os.Stat(prfx + ext1); os.IsNotExist(err) {
		t.Fatalf("File <%s> doesn't exists.", fn1)
	}
	if _, err := os.Stat(prfx + ext2); os.IsNotExist(err) {
		t.Fatalf("File <%s> doesn't exists.", fn2)
	}

	img1, cinf, err := loadNotNarpi(prfx + ext1)
	if err != nil {
		t.Fatal(err, cinf)
	}
	img2, cinf, err := loadNotNarpi(prfx + ext2)
	if err != nil {
		t.Fatal(err, cinf)
	}

	maxx := img1.Bounds().Max.X
	maxy := img1.Bounds().Max.Y

	if !reflect.DeepEqual(img1.Bounds(), img2.Bounds()) {
		t.Fatalf("Images <%s>, <%s> have different sizes.\n\t%s:%v\n\t%s:%v", fn1, fn2, fn1, img1.Bounds(), fn2, img2.Bounds())
	}

	count := 0
	accumError := int64(0)
	for i := 0; i < len(img1.Pix); i++ {
		d := uint64(img1.Pix[i]) - uint64(img2.Pix[i])
		d *= d
		accumError += int64(d)
		if d != 0 {
			count++
		}
	}

	if accumError == 0 {
		log.Printf("Success: images <%s>, <%s> are identical.", fn1, fn2)
	} else {
		log.Printf("ERROR??: images <%s>, <%s> differ in %v pixels (%.2f%% different), accumulative error=%v.",
			fn1, fn2,
			count, float32(count)/float32(maxx*maxy)*100, accumError)
	}
}

func testPngFromNarpi(s1 string, s2 string, t *testing.T) {
	defer timeTrack(time.Now(), "testPngFromNarpi")

	err := Png(s1, s2, true)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Converted:: [%s]->[%s]\n", filepath.Base(s1), filepath.Base(s2))
}

func testJpgFromNarpi(s1 string, s2 string, t *testing.T) {
	defer timeTrack(time.Now(), "testJpgFromNarpi")

	err := Jpg(s1, s2, true)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Converted:: [%s]->[%s]\n", filepath.Base(s1), filepath.Base(s2))
}

func testToNarpi(s1 string, s2 string, t *testing.T) {
	defer timeTrack(time.Now(), "testToNarpi")

	err := Narpi(s1, s2, true)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Converted:: [%s]->[%s]\n", filepath.Base(s1), filepath.Base(s2))

	sz1 := getFileSize(s1) / 1024
	sz2 := getFileSize(s2) / 1024
	gain := float64(sz1-sz2) / float64(sz1) * 100
	log.Printf("INFO::\t\tNarpi gain in size is:: %.2f%%\n\t\t::%s(%vkB)\n\t\t::%s(%vkB)", gain, filepath.Base(s1), sz1, filepath.Base(s2), sz2)
}

func TestGenerationOfLightBuffers(t *testing.T) {
	type args struct {
		pix        *[]uint8
		xs         int
		colorindex uint8
	}
	tests := []struct {
		name string
		args args
		want []uint8
	}{
		{"", struct {
			pix        *[]uint8
			xs         int
			colorindex uint8
		}{&[]uint8{2, 3, 4, 255, 2, 2, 3, 255, 4, 5, 6, 255, 8, 5, 6, 255}, 2, 0},
			[]uint8{0, 0, 0, 5, 2, 2, 129, 4, 8}},

		{"", struct {
			pix        *[]uint8
			xs         int
			colorindex uint8
		}{&[]uint8{2, 3, 4, 255, 2, 2, 3, 255, 4, 5, 6, 255, 8, 5, 6, 255}, 2, 1},
			[]uint8{0, 0, 0, 5, 129, 3, 2, 2, 5}},

		{"", struct {
			pix        *[]uint8
			xs         int
			colorindex uint8
		}{&[]uint8{2, 3, 1, 255, 2, 2, 4, 255, 4, 5, 4, 255, 8, 5, 4, 255}, 2, 2},
			[]uint8{0, 0, 0, 4, 128, 1, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createLightBuffer(tt.args.pix, tt.args.xs, tt.args.colorindex)
			if err != nil {
				t.Errorf("createLightBuffer() error = %v, wantErr %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createLightBuffer() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestDrawingOfLightBuffers(t *testing.T) {
	type args struct {
		buff *[]uint8
	}
	tests := []struct {
		name string
		args args
		want *[]uint8
	}{
		{"", struct {
			buff *[]uint8
		}{&[]uint8{0, 0, 0, 5, 2, 2, 129, 4, 8, 00, 0, 0, 5, 129, 3, 2, 2, 5, 0, 0, 0, 5, 129, 4, 3, 2, 6}},
			&[]uint8{2, 3, 4, 255, 2, 2, 3, 255, 4, 5, 6, 255, 8, 5, 6, 255}},

		{"", struct {
			buff *[]uint8
		}{&[]uint8{0, 0, 0, 5, 131, 1, 2, 3, 2, 0, 0, 0, 4, 131, 1, 2, 5, 1, 0, 0, 0, 4, 2, 1, 129, 6, 8}},
			&[]uint8{1, 1, 1, 255, 2, 2, 1, 255, 3, 5, 6, 255, 2, 1, 8, 255}},

		{"", struct {
			buff *[]uint8
		}{&[]uint8{0, 0, 0, 2, 4, 1, 0, 0, 0, 2, 4, 1, 0, 0, 0, 2, 4, 1}},
			&[]uint8{1, 1, 1, 255, 1, 1, 1, 255, 1, 1, 1, 255, 1, 1, 1, 255}},

		{"", struct {
			buff *[]uint8
		}{&[]uint8{0, 0, 0, 5, 131, 1, 2, 3, 4, 0, 0, 0, 5, 131, 1, 2, 3, 4, 0, 0, 0, 5, 131, 1, 2, 3, 4}},
			&[]uint8{1, 1, 1, 255, 2, 2, 2, 255, 3, 3, 3, 255, 4, 4, 4, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bI := bytes.NewBuffer(*tt.args.buff)
			pix := make([]uint8, len(*tt.want))

			drawLightBuffer(bI, &pix, 0)
			drawLightBuffer(bI, &pix, 1)
			drawLightBuffer(bI, &pix, 2)

			if !reflect.DeepEqual(*tt.want, pix) {
				t.Errorf("drawLightBuffer() = %v, want %v", pix, tt.want)
			}

		})
	}

}

func _TestImageFilesJpgToNARPIToPng(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "jpg")
	deleteExtFiles(t, "png")
	deleteExtFiles(t, "raw")
	deleteExtFiles(t, "narpi")
	for _, s := range fileNames {
		t.Run("ConstructFromJpgFile-Save-Load-DeconstructToPngFile::"+filepath.Base(s), func(t *testing.T) {
			testToNarpi(s+"jpg", s+"narpi", t)
			testPngFromNarpi(s+"narpi", s+"png", t)

			log.Println("___")
			compareJpgPngFiles(t, s)
			log.Println("---------------------------------------------------------------------------")
			log.Println()
		})

	}
}

func _TestImageFilesPngToNARPIToJpg(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "png")
	deleteExtFiles(t, "jpg")
	deleteExtFiles(t, "raw")
	deleteExtFiles(t, "narpi")
	for _, s := range fileNames {
		t.Run("ConstructFromJpgFile-Save-Load-DeconstructToPngFile::"+filepath.Base(s), func(t *testing.T) {
			testToNarpi(s+"png", s+"narpi", t)
			testJpgFromNarpi(s+"narpi", s+"jpg", t)

			log.Println("___")
			compareJpgPngFiles(t, s)
			log.Println("---------------------------------------------------------------------------")
			log.Println()
		})

	}
}
