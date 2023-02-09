package libgen

import (
	"fmt"
	"generator/constant"
	"generator/util"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/abema/go-mp4"
	"github.com/gingfrederik/docx"
	"github.com/jung-kurt/gofpdf"
)

func GenerateFiles(rootpath string, limit, depth int64) error {
	if depth == 0 {
		return nil
	}

	var active int
	err := make(chan error)

	for index := int64(0); index < limit; index++ {
		if active > constant.MaxThreadCount {
			if <-err != nil {
				return <-err
			}
			active--
		}

		pathfmt := fmt.Sprintf("%s_%d", util.GetRandomString(constant.FileNameLen), index)
		path := filepath.Join(rootpath, pathfmt)

		go populateFolder(path, limit, err)
		active++

		GenerateFiles(path, limit, depth-1)
	}

	for active > 0 {
		if <-err != nil {
			return <-err
		}
		active--
	}

	return nil
}

func populateFolder(path string, limit int64, channeledErr chan error) {
	ioerr := os.MkdirAll(path, os.ModePerm)
	if ioerr != nil {
		channeledErr <- ioerr
	}

	var active int
	generr := make(chan error)

	for index := int64(0); index < limit; index++ {
		if active > constant.MaxThreadCount {
			if <-generr != nil {
				channeledErr <- <-generr
			}
			active--
		}

		go generateTxt(path, generr)
		active++

		go generateDocx(path, generr)
		active++

		go generatePng(path, generr)
		active++

		go generatePdf(path, generr)
		active++

		go generateMp4(path, generr)
		active++
	}

	for active > 0 {
		if <-generr != nil {
			channeledErr <- <-generr
		}
		active--
	}

	channeledErr <- nil
}

func generateTxt(basedir string, channeledErr chan error) {
	txtpath := util.GetFilePath(basedir, constant.TxtExtension)
	randomData := []byte(util.GetRandomString(constant.ContentLen))
	channeledErr <- os.WriteFile(txtpath, randomData, os.ModePerm)
}

func generateDocx(basedir string, channeledErr chan error) {
	docfile := docx.NewFile()
	para := docfile.AddParagraph()
	randomData := util.GetRandomString(constant.ContentLen)
	para.AddText(randomData).Size(12)
	docxpath := util.GetFilePath(basedir, constant.DocxExtension)
	channeledErr <- docfile.Save(docxpath)
}

func generatePng(basedir string, channeledErr chan error) {
	width := 500
	height := 500
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	cyan := color.RGBA{100, 200, 200, 0xff}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2:
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2:
				img.Set(x, y, color.White)
			default:
			}
		}
	}

	pngpath := util.GetFilePath(basedir, constant.PngExtension)
	pngfile, pngerr := os.Create(pngpath)
	if pngerr != nil {
		channeledErr <- pngerr
	}
	encodeErr := png.Encode(pngfile, img)
	if encodeErr != nil {
		channeledErr <- encodeErr
	}
	channeledErr <- pngfile.Close()
}

func generatePdf(basedir string, channeledErr chan error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	for i := 0; i < 25; i++ {
		randomData := util.GetRandomString(20)
		pdf.Text(10, 10*float64(i), randomData)
	}
	pdfpath := util.GetFilePath(basedir, constant.PdfExtension)
	channeledErr <- pdf.OutputFileAndClose(pdfpath)
}

func generateMp4(basedir string, channeledErr chan error) {
	mp4path := util.GetFilePath(basedir, constant.Mp4Extension)
	mp4file, mp4err := os.Create(mp4path)
	if mp4err != nil {
		channeledErr <- mp4err
	}
	mp4writer := mp4.NewWriter(mp4file)
	randomData := []byte(util.GetRandomString(constant.ContentLen))
	_, writeErr := mp4writer.Write(randomData)
	if writeErr != nil {
		channeledErr <- writeErr
	}
	channeledErr <- mp4file.Close()
}
