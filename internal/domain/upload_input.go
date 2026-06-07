package domain

import "io"

type UploadInput struct {
	Object      io.Reader
	ObjectName  string
	ObjectSize  int64
	ContentType string
}
