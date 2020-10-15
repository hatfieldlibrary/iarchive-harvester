package types

type Thesis struct {
	Title string
	IarchiveID string
	Oclc string
}

type InputData struct {
	Data[]Thesis
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