package constant

import "runtime"

const (
	FileNameLen        = 10
	ContentLen         = 200000
	DefaultDepth int64 = 2
)

const (
	TxtExtension  = ".txt"
	DocxExtension = ".docx"
	PngExtension  = ".png"
	PdfExtension  = ".pdf"
	Mp4Extension  = ".mp4"
)

var MaxThreadCount = runtime.NumCPU() * 2
