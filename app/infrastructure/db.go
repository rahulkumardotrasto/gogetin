package infrastructure

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	mgo "gopkg.in/mgo.v2"
)

//DB  Infrastructure layer to incorporate all the database function
type DB struct{}

//MongoDB ...
type MongoDB struct {
	MongoDatabase   *mgo.Database
	MongoDBDatabase *mongo.Database
}

var _initMongoCtx sync.Once
var _mongoInstance *MongoDB
var _session *mgo.Session

var _mongoClient *mongo.Client

//Database configuration constants
const (
	Debug   = false
	LocalDB = "test1"
)

var _dbConfig = ConfigService{}

//GetMongo ...
func (ds *DB) GetMongo() *mgo.Database {
	_initMongoCtx.Do(func() {
		_logger.Info("Connecting to Mongo.....")
		var err error
		if !Debug {
			_config := _dbConfig.GetConfiguration()
			info := &mgo.DialInfo{
				Addrs:    []string{_config.MongoDb.HostName},
				Timeout:  60 * time.Second,
				Database: _config.MongoDb.Database,
				// Username: _config.MongoDb.UserName,
				// Password: _config.MongoDb.Password,
			}
			_session, err = mgo.DialWithInfo(info)
			_session.SetMode(mgo.Monotonic, true)
			_session.SetPoolLimit(10)
			if _mongoClient == nil {
				_mongoInstance = &MongoDB{MongoDatabase: _session.DB(_config.MongoDb.Database)}
			} else {
				_mongoInstance = &MongoDB{
					MongoDBDatabase: _mongoClient.Database(_config.MongoDb.Database),
					MongoDatabase:   _session.DB(_config.MongoDb.Database)}
			}
		} else {
			_session, err = mgo.Dial("mongodb://localhost:27017")
			_mongoInstance = &MongoDB{MongoDatabase: _session.DB(LocalDB)}
		}
		if err != nil {
			_logger.Error("Failed to connect to Mongo..." + " " + err.Error())
			return
		}
		_logger.Info("Connected to Mongo...")
	})
	if !Debug {
		_session.Refresh()
	}
	return _mongoInstance.MongoDatabase
}
