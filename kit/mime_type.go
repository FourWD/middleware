package kit

type MimeType string

func (m MimeType) String() string {
	return string(m)
}

const (
	TextCSV   = MimeType("text/csv")
	TextHTML  = MimeType("text/html")
	TextPlain = MimeType("text/plain")

	ApplicationXML         = MimeType("application/xml")
	ApplicationPDF         = MimeType("application/pdf")
	ApplicationZip         = MimeType("application/zip")
	ApplicationOctetStream = MimeType("application/octet-stream")
)
