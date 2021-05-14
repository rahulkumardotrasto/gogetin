package infrastructure

import (
	"../utils"
	"github.com/tkanos/gonfig"
)

//ConfigService : Utility layer to incorporate all the helper function
type ConfigService struct{}

var _logger = utils.GetLogger()

//Configuration : Database configuration
type Configuration struct {
	MongoDb struct {
		HostName string `json:"host_name"`
		Database string `json:"database"`
		// UserName    string `json:"user_name"`
		// Password    string `json:"password"`
		Collections struct {
			Channels string `json:"channels"`
			Videos   string `json:"videos"`
			Users    string `json:"users"`
		} `json:"collections"`
	} `json:"mongo_db"`
	Otp struct {
		Host      string `json:"host"`
		SendURL   string `json:"send_url"`
		VerifyURL string `json:"verify_url"`
		AuthKey   string `json:"auth_key"`
		Sender    string `json:"sender"`
		Length    string `json:"length"`
		Expire    string `json:"expire"`
	} `json:"otp"`
	LogLevel string `json:"log_level"`
}

//GetConfiguration : Returns configuration for the database
func (cs *ConfigService) GetConfiguration() Configuration {
	var configuration = Configuration{}
	var err = gonfig.GetConf("./config.json", &configuration)
	if err != nil {
		_logger.Error("Error access configuration file: " + err.Error())
	}
	return configuration
}
