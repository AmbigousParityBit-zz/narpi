package libnarpi

import (
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

func TestImageFilesJpgToNARPIToPng(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "jpg")
	deleteExtFiles(t, "png")
	deleteExtFiles(t, "raw")
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
	for _, s := range fileNames {
		t.Run("ConstructFromJpgFile-Save-Load-DeconstructToPngFile::"+filepath.Base(s), func(t *testing.T) {
			testToNarpi(s+"png", s+"narpi", t)
			testPngFromNarpi(s+"narpi", s+"jpg", t)

			log.Println("___")
			compareJpgPngFiles(t, s)
			log.Println("---------------------------------------------------------------------------")
			log.Println()
		})

	}
}
