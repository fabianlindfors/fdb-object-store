package main

import (
	"bytes"
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

var db fdb.Database

type File struct {
	Data        []byte
	ContentType string
}

func getFile(name string) *File {
	file, _ := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		contentType := tr.Get(tuple.Tuple{name, "content-type"}).MustGet()
		if contentType == nil {
			return nil, nil
		}

		// Retrieve the split data using a prefix query
		start, end := tuple.Tuple{name, "data"}.FDBRangeKeys()
		r := fdb.KeyRange{Begin: start, End: end}
		kvSlice := tr.GetRange(r, fdb.RangeOptions{}).GetSliceOrPanic()

		// Combine the retrieved file data into a buffer
		var b bytes.Buffer
		for _, kv := range kvSlice {
			b.Write(kv.Value)
		}

		return &File{Data: b.Bytes(), ContentType: string(contentType)}, nil
	})

	if file == nil {
		return nil
	}

	return file.(*File)
}

func saveFile(name string, contentType string, reader io.Reader) {
	db.Transact(func(tr fdb.Transaction) (ret interface{}, e error) {
		buffer := make([]byte, 10000)
		i := 0

		for {
			n, err := reader.Read(buffer)

			if err == io.EOF {
				break
			}

			tr.Set(tuple.Tuple{name, "data", i}, buffer[:n])
			i++
		}

		tr.Set(tuple.Tuple{name, "content-type"}, []byte(contentType))

		return
	})
}

func main() {
	fdb.MustAPIVersion(510)
	db = fdb.MustOpenDefault()

	router := gin.Default()

	router.GET("/object/*name", func(c *gin.Context) {
		name := c.Param("name")
		file := getFile(name)

		// File not found in FDB
		if file == nil {
			c.AbortWithStatus(404)
			return
		}

		// Split file path by slash to get file name
		splitName := strings.Split(name, "/")

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", "attachment; filename="+splitName[len(splitName)-1])
		c.Data(200, file.ContentType, file.Data)
	})

	router.POST("/object/*name", func(c *gin.Context) {
		name := c.Param("name")

		// Content type will be needed to enable downloads later
		contentType := c.PostForm("content_type")
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		reader, _ := file.Open()
		defer reader.Close() // Make sure to close the file handle

		saveFile(name, contentType, reader)

		c.String(200, "File saved")
	})

	router.Run()
}
