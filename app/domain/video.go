package domain

import "gopkg.in/mgo.v2/bson"

//Video ...
type Video struct {
	VideoURL     string        `json:"video_url" bson:"video_url"`
	ThumbnailURL string        `json:"thumbnail_url" bson:"thumbnail_url"`
	UserID       bson.ObjectId `json:"user_id" bson:"user_id"`
	ChannelIDS   []ChannelID   `json:"channel_ids" bson:"channel_ids"`
}

//ChannelID ...
type ChannelID struct {
	ID bson.ObjectId `json:"_id" bson:"_id"`
}
