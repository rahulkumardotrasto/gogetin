package domain

import (
	"gopkg.in/mgo.v2/bson"
)

//User Model for the user data
type User struct {
	ID           bson.ObjectId `json:"_id" bson:"_id"`
	PhoneNumber  string        `json:"phone_number" bson:"phone_number"`
	UserChannels []UserChannel `json:"channels" bson:"channels"`
}

// UserChannel ...
type UserChannel struct {
	ChannelID    bson.ObjectId `json:"channel_id" bson:"channel_id"`
	Name         string        `json:"name" bson:"name"`
	ThumbnailURL string        `json:"thumbnail_url" bson:"thumbnail_url"`
	Deleted      bool          `json:"deleted" bson:"deleted"`
}

//DeleteChannelsFromUserReqBody ...
type DeleteChannelsFromUserReqBody struct {
	UserID   string           `json:"user_id" bson:"user_id"`
	Channels []ChannelReqData `json:"channels" bson:"channels"`
}

//AddChannelsToUserReqBody ...
type AddChannelsToUserReqBody struct {
	UserID   string           `json:"user_id" bson:"user_id"`
	Channels []ChannelReqData `json:"channels" bson:"channels"`
}

// ChannelReqData ...
type ChannelReqData struct {
	ChannelID    string `json:"channel_id" bson:"channel_id"`
	Name         string `json:"name" bson:"name"`
	ThumbnailURL string `json:"thumbnail_url" bson:"thumbnail_url"`
}
