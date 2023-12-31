package object

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/casbin/casibase/storage"
)

func UpdateFile(storeId string, key string, file *File) bool {
	return true
}

func AddFile(storeId string, key string, isLeaf bool, filename string, file multipart.File) (bool, []byte) {
	store := GetStore(storeId)
	if store == nil {
		return false, nil
	}

	var objectKey string
	var fileBuffer *bytes.Buffer
	if isLeaf {
		objectKey = fmt.Sprintf("%s/%s", key, filename)
		objectKey = strings.TrimLeft(objectKey, "/")
		fileBuffer = bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, file); err != nil {
			panic(err)
		}
	} else {
		objectKey = fmt.Sprintf("%s/%s/_hidden.ini", key, filename)
		objectKey = strings.TrimLeft(objectKey, "/")
		fileBuffer = bytes.NewBuffer(nil)
	}

	bs := fileBuffer.Bytes()
	storage.PutObject(store.Bucket, objectKey, fileBuffer)
	return true, bs
}

func DeleteFile(storeId string, key string, isLeaf bool) bool {
	store := GetStore(storeId)
	if store == nil {
		return false
	}

	if isLeaf {
		storage.DeleteObject(store.Bucket, key)
	} else {
		objects := storage.ListObjects(store.Bucket, key)
		for _, object := range objects {
			storage.DeleteObject(store.Bucket, object.Key)
		}
	}
	return true
}
