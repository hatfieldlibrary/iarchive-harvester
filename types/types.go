package types

type ApiKey struct {
	Comment string
	Key string
}
type Record struct {
	Title string
	IarchiveID string
	Oclc string
}

type InputData struct {
	Data[]Record
}

type DataSource struct {
	File   string
	OclcNumber string
	Source string
	BaseUrl string
}

type IArchiveFileFormat struct {
	Name string
	Source string
	Mtime string
	Size string
	Md5 string
	Crc32 string
	Sha1 string
	Format string
}

type Audit struct {
	Title string
	Author string
	Date string
	Description string
	IArchiveID string
	OCLCNumber string
	OutputDirectory string
}