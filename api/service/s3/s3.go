package s3

import (
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go"

	"etix-tv-manager/api/model"
)

const (
	s3location = "us-east-1"
	bucketName = "media"
)

var client *minio.Client

// Init is used to initialize the minio client.
// It will connect the app to the minio server or return an error.
func Init(endpoint, accessKeyID, secretAccessKey string, ssl bool) error {
	var err error
	var found bool

	// Create s3 client.
	client, err = minio.New(endpoint, accessKeyID, secretAccessKey, ssl)
	if err != nil {
		return err
	}

	// Create app bucket.
	if found, err = client.BucketExists(bucketName); !found && err == nil {
		if err = client.MakeBucket(bucketName, s3location); err != nil {
			log.Fatalln("S3 initialization: can't create bucket: ", err)
		}
	} else if err != nil {
		log.Fatalln("S3 initialization: can't find bucket: ", err)
	}
	return nil
}

// ListObjects return all objects in a folder.
func ListObjects(folder string) ([]model.Media, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	mediaList := []model.Media{}

	objectCh := client.ListObjectsV2(bucketName, folder, true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return nil, object.Err
		}
		mediaList = append(mediaList, model.Media{Name: object.Key, Size: object.Size, LastModified: object.LastModified})
	}
	return mediaList, nil
}

// NOTE: For now GetObject return a minio object.
// A DTO custom type will be created to abstract the minio one.

// GetObject return an object described by its name.
func GetObject(name string) (*minio.Object, error) {
	object, err := client.GetObject(bucketName, name)
	if err != nil {
		log.Fatalln("S3 operation: can't get object [", name, "]: ", err)
		return nil, err
	}
	return object, nil
}

// Upload is used to upload a file.
func Upload(name string, file io.Reader) error {
	_, err := client.PutObject(bucketName, name, file, "application/octet-stream")
	if err != nil {
		log.Println("S3 operation: can't upload file [", name, "]: ", err)
		return err
	}
	return nil
}

// Remove is used remove a file.
func Remove(name string) error {
	err := client.RemoveObject(bucketName, name)
	if err != nil {
		log.Println("S3 operation: can't remove file [", name, "]: ", err)
		return err
	}
	return nil
}

// Rename is used to rename a file.
func Rename(name, newName string) error {
	src := minio.NewSourceInfo(bucketName, name, nil)
	dst, err := minio.NewDestinationInfo(bucketName, newName, nil, nil)
	if err != nil {
		log.Println("S3 operation: can't rename file [", name, "]: ", err)
		return err
	}

	err = client.CopyObject(dst, src)
	if err != nil {
		log.Println("S3 operation: can't rename file [", name, "]: ", err)
		return err
	}

	return Remove(name)
}
