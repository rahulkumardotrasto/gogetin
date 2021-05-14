package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"../domain"
	"../utils"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2/bson"
)

//VideoService ...
type VideoService struct{}

//UploadVideo ...
func (vs *VideoService) UploadVideo(reqBody map[string]string, video *multipart.FileHeader) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var database = _db.GetMongo()
	fileExtension := filepath.Ext(video.Filename)
	fmt.Println("fileExtension")
	fmt.Println(fileExtension)
	if fileExtension == "" {
		fileExtension = ".mp4"
	}
	videoFileName := uuid.New().String() + fileExtension
	file, err := video.Open()
	defer file.Close()
	if bson.IsObjectIdHex(reqBody["channel_id"]) {
		if database != nil {
			if err == nil {
				videoBytes, err := ioutil.ReadAll(file)
				if err == nil {
					statusCode, success, message = vs.UploadToStorage(videoBytes, videoFileName, fileExtension, reqBody["Content-Type"], utils.VIDEO)
					if success {
						config := _config.GetConfiguration()
						collection := database.C(config.MongoDb.Collections.Videos)
						var video domain.Video
						video.VideoURL = "https://storage.googleapis.com/gogetin/" + videoFileName
						video.UserID = bson.ObjectIdHex(reqBody["user_id"])

						channelIDS := []domain.ChannelID{}
						channelID := domain.ChannelID{}
						channelID.ID = bson.ObjectIdHex(reqBody["channel_id"])
						channelIDS = append(channelIDS, channelID)
						video.ChannelIDS = channelIDS
						err := collection.Insert(video)
						if err == nil {
							statusCode = http.StatusOK
							success = true
							message = "Successfully uploaded and saved video."
						} else {
							_logger.Error("Error in inserting video: " + err.Error())
						}
					} else {
						_logger.Error("Error in uploading video: " + err.Error())
					}
				} else {
					_logger.Error("Error in uploading video: " + err.Error())
				}
			} else {
				_logger.Error("Error in uploading video: " + err.Error())
			}
		} else {
			_logger.Error("Error connecting to db.")
		}
	} else {
		_logger.Error("Invalid channel id.")
		statusCode = http.StatusBadRequest
		message = "Invalid id."
	}
	return statusCode, success, message
}

//UploadToStorage ...
func (vs *VideoService) UploadToStorage(fileBytes []byte, fileName, ext, contentType, fileType string) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	StorageBucketName := "gogetin"
	StorageBucket, err := vs.configureStorage(StorageBucketName)
	if err == nil {
		ctx := context.Background()
		w := StorageBucket.Object(fileName).NewWriter(ctx)
		w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
		// Entries are immutable, be aggressive about caching (1 day).
		w.CacheControl = "public, max-age=86400"
		r := bytes.NewReader(fileBytes)
		_, err := io.Copy(w, r)
		err = w.Close()
		if err == nil {
			statusCode = http.StatusOK
			success = true
			message = "Succcessfully uploaded."
		} else {
			_logger.Error("Error in uploading video: " + err.Error())
		}
	} else {
		_logger.Error("Error in uploading video: " + err.Error())
	}
	return statusCode, success, message
}

func (vs *VideoService) configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}
