package narpi

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func _TestcutBytesOfUint16(t *testing.T) {
	var bytesCutTests = []struct {
		in   uint16
		outB bool
		outL uint8
		outR uint8
	}{
		{60116, true, 234, 212},
		{14231, true, 111, 23},
		{16383, true, 255, 255},
		{255, false, 255, 255},
	}

	for _, bt := range bytesCutTests {
		if outB_, outL_, outR_ := cutBytesOfUint16(bt.in); bt.outB != outB_ &&
			bt.outL != outL_ && bt.outR != outR_ {
			t.Fatalf("cutBytesOfUint16(%v) => %v,%v,%v, want %v,%v,%v", bt.in, outB_, outL_, outR_,
				bt.outB, bt.outL, bt.outR)
		} else {
			log.Printf("cutBytesOfUint16(%v) => %v,%v,%v, successfully", bt.in, bt.outB, bt.outL, bt.outR)
		}
	}
}

func _TestputBytesToUint16(t *testing.T) {
	var bytesPutTests = []struct {
		in  []uint8
		out uint16
	}{
		{[]uint8{122}, 122},
		{[]uint8{234, 212}, 60116},
		{[]uint8{111, 23}, 28439},
		{[]uint8{255, 255}, 65535},
	}

	for _, bt := range bytesPutTests {
		if v := putBytesToUint16(bt.in); v != bt.out {
			t.Fatalf("putBytesToUint16(%v) => %v, want %v", bt.in, v, bt.out)
		} else {
			log.Printf("putBytesToUint16(%v) => %v, successfully", bt.in, bt.out)
		}
	}
}

func getTestImagesFilenames(t *testing.T, ext string) []string {
	defer timeTrack(time.Now(), "getTestImagesFilenames")
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
		t.Fatalf("couldn't delete files <*%v> during preparations to tests", ext)
	}
	log.Printf("Successfully deleted all <*%v> files during preparations to tests.", ext)
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

	if img1.Bounds() != img2.Bounds() {
		t.Fatalf("Images <%s>, <%s> have different sizes.", fn1, fn2)
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

func testConstructFromFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testConstructFromFile")

	_, fn := filepath.Split(s + "jpg")
	log.Printf("Constructing NARP image in memory from <%s>.\n", fn)
	err := narpimg.Load(s+"jpg", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Constructed NARP image in memory from <%s> ; bounds: %v, %v.\n",
		fn, narpimg.Size.X, narpimg.Size.Y)
	log.Printf("Number of\n\t\t pixels = %v,\n\t\t keys = %v.\nGain in reduction of pixel objects: %v%%.\n",
		int(narpimg.Size.X)*int(narpimg.Size.Y), len(narpimg.NARPixels),
		100-100*len(narpimg.NARPixels)/(int(narpimg.Size.X)*int(narpimg.Size.Y)))
}

func testSave(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "Save")

	_, fn := filepath.Split(s + "narp")
	err := narpimg.Save(s+"narp", true)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%s>.\n", fn)
}

func testLoad(s string, narpimgAfterLoading *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "Load")

	_, fn := filepath.Split(s + "narp")
	err := narpimgAfterLoading.Load(s+"narp", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Loaded NARP image from file <%s> to memory.\n", fn)
}

func testDeconstructToPngFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testDeconstructToPngFile")

	_, fn := filepath.Split(s + "png")
	err := narpimg.Png(s + "png")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%s>.\n", fn)
}

func testDeconstructToJpgFile(s string, narpimg *NARPImage, t *testing.T) {
	defer timeTrack(time.Now(), "testDeconstructToJpgFile")

	_, fn := filepath.Split(s + "jpg")
	err := narpimg.Jpg(s + "jpg")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Saved NARP image from memory to file <%sjpg>.\n", fn)
}

func TestImageFilesJpgToNARPIToPng(t *testing.T) {
	fileNames := getTestImagesFilenames(t, "jpg")
	deleteExtFiles(t, "png")
	for _, s := range fileNames {
		t.Run("ConstructFromJpgFile-Save-Load-DeconstructToPngFile::"+filepath.Base(s), func(t *testing.T) {
			narpimg := NARPImage{}
			narpimgAfterLoading := NARPImage{}

			testConstructFromFile(s, &narpimg, t)
			testSave(s, &narpimg, t)
			testLoad(s, &narpimgAfterLoading, t)
			if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
				log.Printf("Loaded NARP image is the same as the previous one in memory, as expected.\n")
			} else {
				t.Fatalf("Loaded NARP image is different from the previous one in memory.\n")
			}
			testDeconstructToPngFile(s, &narpimgAfterLoading, t)

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
	for _, s := range fileNames {
		t.Run("ConstructFromPngFile-Save-Load-DeconstructToJpgFile::"+filepath.Base(s), func(t *testing.T) {
			narpimg := NARPImage{}
			narpimgAfterLoading := NARPImage{}

			testConstructFromFile(s, &narpimg, t)
			testSave(s, &narpimg, t)
			testLoad(s, &narpimgAfterLoading, t)
			if reflect.DeepEqual(narpimg, narpimgAfterLoading) {
				log.Printf("Loaded NARP image is the same as the previous one in memory, as expected.\n")
			} else {
				t.Fatalf("Loaded NARP image is different from the previous one in memory.\n")
			}
			testDeconstructToJpgFile(s, &narpimgAfterLoading, t)

			log.Println("___")
			compareJpgPngFiles(t, s)
			log.Println("---------------------------------------------------------------------------")
			log.Println()
		})

	}
}
