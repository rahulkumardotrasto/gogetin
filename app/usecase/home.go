package usecase

import (
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

//HomeService ...
type HomeService struct{}

//HomePageVideos ...
func (hs *HomeService) HomePageVideos() (int, bool, string, []map[string]interface{}) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var response []map[string]interface{}
	var database = _db.GetMongo()
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Videos)
		err := collection.Find(bson.M{}).All(&response)
		if err == nil {
			success = true
			if len(response) == 0 {
				_logger.Error("No videos present.")
				statusCode = http.StatusNotFound
				message = "No data present."
			} else {
				statusCode = http.StatusOK
				message = "successfully retrieved videos."
			}
		} else {
			_logger.Error("Error in fetching videos: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message, response
}
