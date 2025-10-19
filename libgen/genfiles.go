package libgen

import (
	"archive/zip"
	"fmt"
	"generator/constant"
	"generator/util"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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

		go generateCsv(path, generr)
		active++

		go generateJson(path, generr)
		active++

		go generateXml(path, generr)
		active++

		go generateHtml(path, generr)
		active++

		go generateLog(path, generr)
		active++

		go generateReg(path, generr)
		active++

		go generateZip(path, generr)
		active++

		go generateExe(path, generr)
		active++

		go generateJsonl(path, generr)
		active++

		go generateSyslog(path, generr)
		active++

		go generateMarkdown(path, generr)
		active++

		go generateEml(path, generr)
		active++

		go generateMbox(path, generr)
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

func generateCsv(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.CsvExtension)
	rows := []string{"id,name,value"}
	for i := 0; i < 10; i++ {
		rows = append(rows, fmt.Sprintf("%d,%s,%s", i+1, util.GetRandomString(8), util.GetRandomString(12)))
	}
	channeledErr <- os.WriteFile(p, []byte(strings.Join(rows, "\n")+"\n"), os.ModePerm)
}

func generateJson(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.JsonExtension)
	content := fmt.Sprintf(`{"id":%d,"name":"%s","timestamp":"%s"}`,
		1, util.GetRandomString(8), time.Now().UTC().Format(time.RFC3339))
	channeledErr <- os.WriteFile(p, []byte(content), os.ModePerm)
}

func generateXml(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.XmlExtension)
	content := fmt.Sprintf(`<root><id>%d</id><name>%s</name></root>`, 1, util.GetRandomString(8))
	channeledErr <- os.WriteFile(p, []byte(content), os.ModePerm)
}

func generateHtml(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.HtmlExtension)
	body := util.GetRandomString(64)
	html := fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><title>%s</title></head><body><p>%s</p></body></html>`, body[:8], body)
	channeledErr <- os.WriteFile(p, []byte(html), os.ModePerm)
}

func generateLog(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.LogExtension)
	lines := make([]string, 0, 50)
	t := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 50; i++ {
		msg := util.GetRandomString(24)
		lines = append(lines, fmt.Sprintf("%s INFO user=%s event=%s", t.Format(time.RFC3339), util.GetRandomString(6), msg))
		t = t.Add(1 * time.Minute)
	}
	channeledErr <- os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), os.ModePerm)
}

func generateReg(basedir string, channeledErr chan error) {
	// Windows .reg format (REGEDIT4 or Windows Registry Editor Version 5.00)
	p := util.GetFilePath(basedir, constant.RegExtension)
	content := "Windows Registry Editor Version 5.00\r\n\r\n" +
		fmt.Sprintf("[HKEY_CURRENT_USER\\Software\\Fsagen\\%s]\r\n\"Value\"=\"%s\"\r\n", util.GetRandomString(6), util.GetRandomString(12))
	channeledErr <- os.WriteFile(p, []byte(content), os.ModePerm)
}

func generateZip(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.ZipExtension)
	f, err := os.Create(p)
	if err != nil {
		channeledErr <- err
		return
	}
	zw := zip.NewWriter(f)
	// add a few deterministic files
	for i := 0; i < 3; i++ {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:     fmt.Sprintf("file%d.txt", i+1),
			Method:   zip.Store,
			Modified: time.Date(2021, 1, 1, i, 0, 0, 0, time.UTC),
		})
		if err != nil {
			_ = zw.Close()
			_ = f.Close()
			channeledErr <- err
			return
		}
		_, _ = io.WriteString(w, util.GetRandomString(32))
	}
	if err := zw.Close(); err != nil {
		_ = f.Close()
		channeledErr <- err
		return
	}
	channeledErr <- f.Close()
}

// generateExe writes a minimal DOS MZ stub .exe intended for filesystem artifact testing.
// It is not a functional Windows PE binary but includes the 'MZ' signature and DOS-mode message.
func generateExe(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.ExeExtension)
	// Minimal MZ header with DOS stub message; simplifies to a recognizable EXE artifact.
	buf := make([]byte, 256)
	buf[0] = 'M'
	buf[1] = 'Z'
	copy(buf[0x40:], []byte("This program cannot be run in DOS mode.\r\r\n$"))
	channeledErr <- os.WriteFile(p, buf, os.ModePerm)
}

func generateJsonl(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.JsonlExtension)
	t := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	var b strings.Builder
	for i := 0; i < 50; i++ {
		line := fmt.Sprintf(`{"ts":"%s","level":"info","user":"%s","msg":"%s"}`,
			t.Format(time.RFC3339), util.GetRandomString(6), util.GetRandomString(18))
		b.WriteString(line)
		b.WriteByte('\n')
		t = t.Add(30 * time.Second)
	}
	channeledErr <- os.WriteFile(p, []byte(b.String()), os.ModePerm)
}

func generateSyslog(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.SyslogExtension)
	t := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	host := "host01"
	app := "fsagen"
	lines := make([]string, 0, 50)
	for i := 0; i < 50; i++ {
		msg := util.GetRandomString(20)
		lines = append(lines, fmt.Sprintf("%s %s %s[%d]: %s", t.Format(time.RFC3339), host, app, 1000+i, msg))
		t = t.Add(45 * time.Second)
	}
	channeledErr <- os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), os.ModePerm)
}

func generateMarkdown(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.MdExtension)
	title := util.GetRandomString(12)
	body := util.GetRandomString(80)
	md := fmt.Sprintf("# %s\n\n%s\n", title, body)
	channeledErr <- os.WriteFile(p, []byte(md), os.ModePerm)
}

func generateEml(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.EmlExtension)
	date := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC1123Z)
	from := "alice@example.com"
	to := "bob@example.com"
	subj := "Test message " + util.GetRandomString(6)
	body := util.GetRandomString(120)
	eml := fmt.Sprintf("Date: %s\r\nFrom: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", date, from, to, subj, body)
	channeledErr <- os.WriteFile(p, []byte(eml), os.ModePerm)
}

func generateMbox(basedir string, channeledErr chan error) {
	p := util.GetFilePath(basedir, constant.MboxExtension)
	var b strings.Builder
	base := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		ts := base.Add(time.Duration(i) * time.Hour).Format(time.RFC1123Z)
		subj := fmt.Sprintf("Message %d %s", i+1, util.GetRandomString(6))
		body := util.GetRandomString(100)
		b.WriteString(fmt.Sprintf("From alice@example.com %s\n", ts))
		b.WriteString(fmt.Sprintf("Date: %s\nFrom: alice@example.com\nTo: bob@example.com\nSubject: %s\n\n%s\n\n", ts, subj, body))
	}
	channeledErr <- os.WriteFile(p, []byte(b.String()), os.ModePerm)
}
