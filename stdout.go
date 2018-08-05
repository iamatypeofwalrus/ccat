package main

import "os"

// stdout implements WriteAt but passes any invocation to os.Stdout's Write
// method.
//
// Technically, os.Stdout does implement the WriteAt interface but when it is called by the
// Download manager it errors out with the following error:
// 		write /dev/stdout: device not configured
// Since we've set the download manager to stream from S3 sequentially we can
// ignore the offset and just pass all WriteAt calls to WriteWrite
type stdout struct{}

// WriteAt passes every invocation to os.Stdout.Write
func (s stdout) WriteAt(p []byte, off int64) (n int, err error) {
	return os.Stdout.Write(p)
}
