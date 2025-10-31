package constant

import "runtime"

// Supported file types and behaviors are documented via these extension constants.
// Bulk generator and manifest/playbook modes can create files of these types:
// - Text and docs: .txt, .md, .docx, .pdf
// - Images/media: .png, .mp4, .jpg (via content),
// - Data/markup: .csv, .json, .jsonl, .xml, .html
// - Logs: .log, .syslog (also JSONL logs)
// - Archives: .zip
// - Email: .eml, .mbox
// - Windows artifacts: .reg, .exe (EXE is generated as a minimal PE stub)
// Note: Some formats are simplistic or synthetic, focused on filesystem artifact generation.
const (
	FileNameLen        = 10
	ContentLen         = 200000
	DefaultDepth int64 = 2
)

const (
	TxtExtension    = ".txt"
	DocxExtension   = ".docx"
	PngExtension    = ".png"
	PdfExtension    = ".pdf"
	Mp4Extension    = ".mp4"
	CsvExtension    = ".csv"
	JsonExtension   = ".json"
	XmlExtension    = ".xml"
	HtmlExtension   = ".html"
	LogExtension    = ".log"
	RegExtension    = ".reg"
	ZipExtension    = ".zip"
	JsonlExtension  = ".jsonl"
	EmlExtension    = ".eml"
	MboxExtension   = ".mbox"
	MdExtension     = ".md"
	SyslogExtension = ".syslog"
	ExeExtension    = ".exe"
	DbExtension	 	= ".db"
	SqLiteExtension = ".sqlite"
)

var MaxThreadCount = runtime.NumCPU() * 2

// Number of file types generated per iteration in libgen
const NumFileTypes int64 = 19
