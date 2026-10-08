package job

// Job represents a single digitization/preservation unit within a batch.
type Job struct {
	BaseName       string
	Batch          string
	InfoFile       string
	DataFolder     string
	MetadataFolder string
	OcflFile       string
	ReportFile     string
	UploadFile     string
	Signature      string
	Title          string
}

// Info represents the descriptive metadata decoded from an info.json file.
type Info struct {
	Signature string `json:"signature"`
	Title     string `json:"title"`
}

// UploadInfo represents the upload metadata stored in .upload.json.
type UploadInfo struct {
	Date string `json:"date"`
}
