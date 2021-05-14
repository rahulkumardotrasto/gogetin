package usecase

import (
	"net/http"

	"../infrastructure"
	"../utils"
	"gopkg.in/mgo.v2/bson"
)

//ChannelService ...
type ChannelService struct{}

var _db = infrastructure.DB{}
var _config = infrastructure.ConfigService{}
var _logger = utils.GetLogger()

//GetCategories ...
func (cs *ChannelService) GetCategories() (int, bool, string, []map[string]interface{}) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var response []map[string]interface{}
	var database = _db.GetMongo()
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Channels)
		err := collection.Find(bson.M{"is_parent": true}).All(&response)
		if err == nil {
			success = true
			if len(response) == 0 {
				_logger.Error("No categories present.")
				statusCode = http.StatusNotFound
				message = "No data present."
			} else {
				statusCode = http.StatusOK
				message = "successfully retrieved categories."
			}
		} else {
			_logger.Error("Error in fetching categories: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
		statusCode = http.StatusInternalServerError
	}
	return statusCode, success, message, response
}

//GetChannels ...
func (cs *ChannelService) GetChannels(reqData map[string]string) (int, bool, string, []map[string]interface{}) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var response []map[string]interface{}
	var database = _db.GetMongo()
	channelID := reqData["channel_id"]
	if bson.IsObjectIdHex(channelID) {
		if database != nil {
			config := _config.GetConfiguration()
			collection := database.C(config.MongoDb.Collections.Channels)
			err := collection.Find(bson.M{"_id": bson.ObjectIdHex(channelID)}).All(&response)
			if err == nil {
				success = true
				if len(response) == 0 {
					_logger.Error("No channels present.")
					statusCode = http.StatusNotFound
					message = "No data present."
				} else {
					statusCode = http.StatusOK
					message = "successfully retrieved channels."
				}
			} else {
				_logger.Error("Error in fetching channels: " + err.Error())
			}
		} else {
			_logger.Error("Error connecting to db.")
		}
	} else {
		_logger.Error("Invalid channel id.")
		statusCode = http.StatusBadRequest
		message = "bad id."
	}

	return statusCode, success, message, response
}

//GetChannelVideos ...
func (cs *ChannelService) GetChannelVideos(reqData map[string]string) (int, bool, string, []map[string]interface{}) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var response []map[string]interface{}
	var database = _db.GetMongo()
	channelID := reqData["channel_id"]
	if bson.IsObjectIdHex(channelID) {
		if database != nil {
			config := _config.GetConfiguration()
			collection := database.C(config.MongoDb.Collections.Videos)
			err := collection.Find(bson.M{"channel_ids._id": bson.M{"$in": []interface{}{bson.ObjectIdHex(channelID)}}}).All(&response)
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
	} else {
		_logger.Error("Invalid channel id.")
		statusCode = http.StatusBadRequest
		message = "bad id."
	}
	return statusCode, success, message, response
}
