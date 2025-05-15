package static

import (
	"time"
)

// File представление файла в кеше.
type File struct {
	Body        []byte
	ContentType string
	ExpireAt    time.Time
}

func NewFile(body []byte, contentType string, expireAt time.Time) *File {
	return &File{
		Body:        body,
		ContentType: contentType,
		ExpireAt:    expireAt,
	}
}

func (f *File) IsExpired() bool {
	return time.Now().After(f.ExpireAt)
}
