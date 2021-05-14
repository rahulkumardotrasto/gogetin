package usecase

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"../domain"
	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
)

//UserService : Service layer to incorporate all the apis related to Users
type UserService struct{}

var (
	authKey = []byte("gogetin-key")
)

//Register ...
func (us *UserService) Register(reqData map[string]string) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	phoneNumber := reqData["phone_number"]
	var database = _db.GetMongo()
	if database != nil {
		userExists := false
		statusCode, success, message, userExists = us.DoesUserExists(reqData)
		if success {
			if !userExists {
				config := _config.GetConfiguration()
				collection := database.C(config.MongoDb.Collections.Users)
				statusCode, success, message = us.GenerateOtp(phoneNumber)
				if success {
					err := collection.Insert(bson.M{"phone_number": phoneNumber})
					if err == nil {
						statusCode = http.StatusOK
						success = true
						message = "User successfully registered."
					} else {
						_logger.Error("Error in saving user: " + err.Error())
					}
				}
			} else {
				_logger.Error("User already exists.")
				statusCode = http.StatusConflict
				message = "User already exists."
			}
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message
}

//DoesUserExists ...
func (us *UserService) DoesUserExists(reqData map[string]string) (int, bool, string, bool) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	userExists := false
	var response []map[string]interface{}
	var database = _db.GetMongo()
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Users)
		err := collection.Find(bson.M{"phone_number": reqData["phone_number"]}).All(&response)
		if err == nil {
			success = true
			if len(response) > 0 {
				statusCode = http.StatusOK
				message = "user exists."
				userExists = true
			} else {
				statusCode = http.StatusNoContent
				message = "user does not exist."
			}
		} else {
			_logger.Error("Error in fetching users: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message, userExists
}

// GenerateOtp ...
func (us *UserService) GenerateOtp(phoneNumber string) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var response map[string]interface{}
	config := _config.GetConfiguration()
	url := config.Otp.Host + config.Otp.SendURL + "&otp_length=" + config.Otp.Length + "&authkey=" + config.Otp.AuthKey + "&sender=" + config.Otp.Sender + "&otp_expiry=" + config.Otp.Expire + "&mobile=" + phoneNumber + "&message=" + "##OTP##"
	req, _ := http.NewRequest("POST", url, nil)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	_logger.Info("Response from otp: " + string(body))
	_ = json.Unmarshal(body, &response)
	if response["type"] == "success" {
		statusCode = http.StatusOK
		success = true
		message = "OTP Generated Successfully."
	}
	return statusCode, success, message
}

// VerifyOTP ...
func (us *UserService) VerifyOTP(reqData map[string]string) (int, bool, string, domain.AuthTokenResponse) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	authTokenResponse := domain.AuthTokenResponse{}
	var user domain.User
	var response map[string]interface{}
	config := _config.GetConfiguration()
	url := config.Otp.Host + config.Otp.VerifyURL + "&authkey=" + config.Otp.AuthKey + "&mobile=" + reqData["phone_number"] + "&otp=" + reqData["otp"]
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	_logger.Info("Response from verifying otp: " + string(body))
	_ = json.Unmarshal(body, &response)
	if response["type"] == "success" {
		statusCode, success, message, user = us.GetUser(reqData)
		if success {
			reqData["user_id"] = user.ID.Hex()
			statusCode, success, message, authTokenResponse = us.GenerateToken(reqData)
			if success {
				statusCode = http.StatusOK
				success = true
				message = "OTP verified Successfully."
			}
		}
	}
	return statusCode, success, message, authTokenResponse
}

//GetUser ...
func (us *UserService) GetUser(reqData map[string]string) (int, bool, string, domain.User) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var user domain.User
	var database = _db.GetMongo()
	param := bson.M{}
	if reqData["phone_number"] != "" {
		param["phone_number"] = reqData["phone_number"]
	} else if reqData["user_id"] != "" {
		param["_id"] = bson.ObjectIdHex(reqData["user_id"])
	}
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Users)
		err := collection.Find(param).One(&user)
		if err == nil {
			statusCode = http.StatusOK
			success = true
			message = "User data successfully retrieved."
		} else {
			_logger.Error("Error in fetching users: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message, user
}

//GenerateToken ...
func (us *UserService) GenerateToken(reqData map[string]string) (int, bool, string, domain.AuthTokenResponse) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	authTokenResponse := domain.AuthTokenResponse{}
	expiresAt := time.Now().Add(time.Hour * 2000000)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		Id:        reqData["user_id"],
		Issuer:    "gogetin",
		Audience:  "gogetin users",
		Subject:   "Generate gogetin Token",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expiresAt.Unix(),
	})
	tokenString, err := token.SignedString(authKey)
	if err == nil {
		statusCode = http.StatusOK
		success = true
		message = "Successfully generated token."
		authTokenResponse.Token = tokenString
		authTokenResponse.ExpiresAt = expiresAt
	} else {
		_logger.Error("Error in generating token: " + err.Error())
		statusCode = http.StatusUnauthorized
	}
	return statusCode, success, message, authTokenResponse
}

//UpdateUser ...
func (us *UserService) UpdateUser(reqData map[string]string) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var database = _db.GetMongo()
	param := bson.M{}
	param["_id"] = bson.ObjectIdHex(reqData["user_id"])
	update := bson.M{}
	shouldUpdate := true
	if reqData["name"] != "" {
		update["$set"] = bson.M{"name": reqData["name"]}
	} else {
		shouldUpdate = false
	}
	if shouldUpdate {
		if database != nil {
			config := _config.GetConfiguration()
			collection := database.C(config.MongoDb.Collections.Users)
			err := collection.Update(param, update)
			if err == nil {
				statusCode = http.StatusOK
				success = true
				message = "User data successfully updated."
			} else {
				_logger.Error("Error in updating users: " + err.Error())
			}
		} else {
			_logger.Error("Error connecting to db.")
		}
	} else {
		statusCode = http.StatusBadRequest
		success = false
		message = "No data to update."
	}
	return statusCode, success, message
}

//AddChannelsToUser ...
func (us *UserService) AddChannelsToUser(reqData domain.AddChannelsToUserReqBody) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var database = _db.GetMongo()
	var user domain.User
	param := bson.M{}
	param["_id"] = bson.ObjectIdHex(reqData.UserID)
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Users)
		err := collection.Find(param).One(&user)
		if err == nil {
			userChannels := user.UserChannels
			for _, channelReqData := range reqData.Channels {
				channelSubscribedByUser := false
				for j, userChannel := range userChannels {
					if channelReqData.ChannelID == userChannel.ChannelID.Hex() {
						userChannels[j].Deleted = false
						channelSubscribedByUser = true
						break
					}
				}
				if !channelSubscribedByUser {
					userChannel := domain.UserChannel{}
					userChannel.ChannelID = bson.ObjectIdHex(channelReqData.ChannelID)
					userChannel.Name = channelReqData.Name
					userChannel.ThumbnailURL = channelReqData.ThumbnailURL
					userChannel.Deleted = false
					userChannels = append(userChannels, userChannel)
				}
			}
			update := bson.M{}
			update["$set"] = bson.M{"channels": userChannels}
			err := collection.Update(param, update)
			if err == nil {
				statusCode = http.StatusOK
				success = true
				message = "User's channel data successfully added."
			} else {
				_logger.Error("Error in updating user's channel: " + err.Error())
			}
		} else {
			_logger.Error("Error in finding user: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message
}

// DeleteChannelsFromUser ...
func (us *UserService) DeleteChannelsFromUser(reqData domain.DeleteChannelsFromUserReqBody) (int, bool, string) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var database = _db.GetMongo()
	var user domain.User
	param := bson.M{}
	param["_id"] = bson.ObjectIdHex(reqData.UserID)
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Users)
		err := collection.Find(param).One(&user)
		if err == nil {
			userChannels := user.UserChannels
			for _, channelReqData := range reqData.Channels {
				for j, userChannel := range userChannels {
					if channelReqData.ChannelID == userChannel.ChannelID.Hex() {
						userChannels[j].Deleted = true
						break
					}
				}
			}
			update := bson.M{}
			update["$set"] = bson.M{"channels": userChannels}
			err := collection.Update(param, update)
			if err == nil {
				statusCode = http.StatusOK
				success = true
				message = "User's channel successfully deleted."
			} else {
				_logger.Error("Error in deleting user's channel: " + err.Error())
			}
		} else {
			_logger.Error("Error in finding user: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message
}

//GetUserChannels ...
func (us *UserService) GetUserChannels(reqData map[string]string) (int, bool, string, []domain.User) {
	statusCode := http.StatusInternalServerError
	success := false
	message := "Something went wrong."
	var user []domain.User
	var database = _db.GetMongo()
	param := bson.M{}
	param["_id"] = bson.ObjectIdHex(reqData["user_id"])
	fmt.Println(param["_id"])
	if database != nil {
		config := _config.GetConfiguration()
		collection := database.C(config.MongoDb.Collections.Users)
		pipeline := []bson.M{
			bson.M{"$match": bson.M{"_id": bson.ObjectIdHex(reqData["user_id"])}},
			bson.M{"$unwind": "$channels"},
			bson.M{"$match": bson.M{"channels.deleted": false}},
			bson.M{"$group": bson.M{"_id": bson.M{"_id": "_id"}, "channels": bson.M{"$push": "$$ROOT.channels"}}},
		}
		pipe := collection.Pipe(pipeline)
		err := pipe.All(&user)
		if err == nil {
			statusCode = http.StatusOK
			success = true
			message = "User data successfully retrieved."
		} else {
			_logger.Error("Error in fetching users: " + err.Error())
		}
	} else {
		_logger.Error("Error connecting to db.")
	}
	return statusCode, success, message, user
}
