package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"./domain"
	"./infrastructure"
	"./providers"
	"./usecase"
	"./utils"
	gzip "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//App has router and db instances
type App struct{}

var auth = providers.Auth{}
var channelService = usecase.ChannelService{}
var configService = infrastructure.ConfigService{}
var _logger = utils.GetLogger()
var videoService = usecase.VideoService{}
var userService = usecase.UserService{}
var homeService = usecase.HomeService{}

//Init initializes the app with predefined configuration
func (app *App) Init() {
	rand.Seed(time.Now().UnixNano())
	gin.SetMode(gin.ReleaseMode)
	routes := gin.Default()
	routes.Use(gzip.Gzip(gzip.DefaultCompression))
	routes.Use(gin.ErrorLogger())
	routes.Use(gin.Recovery())

	channels := routes.Group("categories")
	{
		channels.GET("/", GetCategories)                   //channel list page
		channels.GET("/channels", GetChannels)             //channel list page
		channels.GET("/channels/videos", GetChannelVideos) // detail page
	}

	home := routes.Group("home")
	{
		home.GET("/videos", HomePageVideos) // home page
	}

	privateVideos := routes.Group("videos").Use(auth.Authenticate())
	{
		privateVideos.POST("/", UploadVideo)
	}

	privateUsers := routes.Group("users").Use(auth.Authenticate())
	{
		privateUsers.GET("/", GetUser)                 //settings page
		privateUsers.GET("/channels", GetUserChannels) //subscription page

		privateUsers.PUT("/", UpdateUser)                        //settings page
		privateUsers.PUT("/channels", AddChannelsToUser)         //channel list page
		privateUsers.DELETE("/channels", DeleteChannelsFromUser) //subscription page
	}

	users := routes.Group("users")
	{
		users.POST("/", Register)
		users.POST("/verify_otp", VerifyOTP)

		users.OPTIONS("/", VastOptions)

	}

	routes.StaticFile("/favicon.ico", "")
	routes.Run(":9000")
}

//InitLogger ... Inits the logger with the provided log level
func (app App) InitLogger(f *os.File) {
	cfg := configService.GetConfiguration()
	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		_logger.Error("Couldn't read log level. Log level set to INFO as default.")
	} else {
		_logger.SetLevel(lvl)
	}
	// _logger.SetReportCaller(true)
	_logger.SetOutput(io.MultiWriter(os.Stdout, f))
}

//VastOptions ...
func VastOptions(c *gin.Context) {
	setOriginHeaders(c)

	c.String(http.StatusOK, "")
}

//setOriginHeaders ...
func setOriginHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin != "" {
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	}
}

//GetCategories ...
func GetCategories(c *gin.Context) {
	statusCode, success, message, data := channelService.GetCategories()
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//GetChannels ...
func GetChannels(c *gin.Context) {
	reqParams := map[string]string{}
	reqParams["channel_id"] = c.Query("channel_id")
	statusCode, success, message, data := channelService.GetChannels(reqParams)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//GetChannelVideos ...
func GetChannelVideos(c *gin.Context) {
	reqParams := map[string]string{}
	reqParams["channel_id"] = c.Query("channel_id")
	statusCode, success, message, data := channelService.GetChannelVideos(reqParams)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//HomePageVideos ...
func HomePageVideos(c *gin.Context) {
	statusCode, success, message, data := homeService.HomePageVideos()
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//UploadVideo ...
func UploadVideo(c *gin.Context) {
	reqBody := map[string]string{}
	reqBody["user_id"] = c.MustGet("user_id").(string)
	reqBody["channel_id"] = c.PostForm("channel_id")
	video, err := c.FormFile("video")
	fmt.Println("video")
	fmt.Println(video)
	c.Header("Access-Control-Allow-Origin", "*")
	if err != nil {
		c.SecureJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error while uploading: " + err.Error(),
		})
	} else {
		statusCode, success, message := videoService.UploadVideo(reqBody, video)
		c.SecureJSON(statusCode, gin.H{
			"success": success,
			"message": message,
		})
	}
}

//GetUser ...
func GetUser(c *gin.Context) {
	reqParams := map[string]string{}
	reqParams["user_id"] = c.MustGet("user_id").(string)
	statusCode, success, message, data := userService.GetUser(reqParams)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//GetUserChannels ...
func GetUserChannels(c *gin.Context) {
	reqParams := map[string]string{}
	reqParams["user_id"] = c.MustGet("user_id").(string)
	statusCode, success, message, data := userService.GetUserChannels(reqParams)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

//UpdateUser ...
func UpdateUser(c *gin.Context) {
	reqBody := map[string]string{}
	reqBody["user_id"] = c.MustGet("user_id").(string)
	reqBody["name"] = c.PostForm("name")
	statusCode, success, message := userService.UpdateUser(reqBody)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
	})
}

//AddChannelsToUser ...
func AddChannelsToUser(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	addChannelsToUserReqBody := domain.AddChannelsToUserReqBody{}
	channels := []domain.ChannelReqData{}
	fmt.Println(channels)
	addChannelsToUserReqBody.UserID = c.MustGet("user_id").(string)
	if c.BindJSON(&channels) == nil {
		addChannelsToUserReqBody.Channels = channels
		statusCode, success, message := userService.AddChannelsToUser(addChannelsToUserReqBody)
		c.Header("Access-Control-Allow-Origin", "*")
		c.SecureJSON(statusCode, gin.H{
			"success": success,
			"message": message,
		})
	} else {
		c.SecureJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Please provide required input.",
		})
	}
}

//DeleteChannelsFromUser ...
func DeleteChannelsFromUser(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	deleteChannelsFromUserReqBody := domain.DeleteChannelsFromUserReqBody{}
	channels := []domain.ChannelReqData{}
	deleteChannelsFromUserReqBody.UserID = c.MustGet("user_id").(string)
	if c.BindJSON(&channels) == nil {
		deleteChannelsFromUserReqBody.Channels = channels
		statusCode, success, message := userService.DeleteChannelsFromUser(deleteChannelsFromUserReqBody)
		c.SecureJSON(statusCode, gin.H{
			"success": success,
			"message": message,
		})
	} else {
		c.SecureJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Please provide required input.",
		})
	}
}

//Register ...
func Register(c *gin.Context) {
	setOriginHeaders(c)
	reqBody := map[string]string{}
	reqBody["phone_number"] = c.PostForm("phone_number")
	statusCode, success, message := userService.Register(reqBody)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
	})
}

//VerifyOTP ...
func VerifyOTP(c *gin.Context) {
	reqBody := map[string]string{}
	reqBody["otp"] = c.PostForm("otp")
	reqBody["phone_number"] = c.PostForm("phone_number")
	statusCode, success, message, data := userService.VerifyOTP(reqBody)
	c.Header("Access-Control-Allow-Origin", "*")
	c.SecureJSON(statusCode, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}
