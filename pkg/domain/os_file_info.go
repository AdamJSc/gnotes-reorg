package domain

import "os"

// OsFileInfo implements FileInfo for the local file system
type OsFileInfo struct {
	fi os.FileInfo
}

// IsDir implements FileInfo.IsDir()
func (o *OsFileInfo) IsDir() bool {
	return o.fi.IsDir()
}

// IsDir implements FileInfo.Name()
func (o *OsFileInfo) Name() string {
	return o.fi.Name()
}
