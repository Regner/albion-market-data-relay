package client

type Config struct {
	IngestUrl     string
	DisableUpload bool
	SaveLocally   bool
	OfflinePath   string
	Offline       bool
	DumpUnknown		bool
}

var GlobalConfiguration = &Config{}
