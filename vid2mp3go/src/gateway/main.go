package main

import (
	"encoding/json"
	"fmt"
	SendStatus "gateway/send_status"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)
type JsonStruct struct {
	Username	string		`json:"username"`
	Exp			float64		`json:"exp"`
	Admin		bool		`json:"admin"`
}

type RabbitMQMessage struct {
	VideoFid	string		`json:"video_fid"`
	Mp3Fid		string		`json:"mp3_fid"`
	Username	string		`json:"username"`
}

var servicePort string = "8080"
// var dbVideos, dbMp3s *mongo.Database
// var fsVideos, fsMp3s *gridfs.Bucket

// Returns the URL to the Authorization service's /login route
var GetAuthServiceUrl = func () (url string) {
	return "http://" + os.Getenv("AUTH_SVC_ADDRESS")
}

// Check if the HTTP request used the POST method. If POST was used, the function
// returns true. Otherwise MethodNotAllowed is sent and false is returned.
func IsPostRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		SendStatus.MethodNotAllowed(w)
		return false
	}
	return true
}

// Uses the received BasicAuth credentials to request authorization from the auth service.
// If everything is correct, this function returns the JWT token string given by the auth service.
// Otherwise it will write a StatusCode corresponding to what went wrong and return nil.
func AuthorizeUser(username string, password string, w http.ResponseWriter) (tokenString []byte) {
	url := GetAuthServiceUrl() + "/login"
	// Create a new POST request to the auth service
	reqToAuthService, err := http.NewRequest("POST", url, nil)
	if err != nil {
		SendStatus.InternalServerError(w)
		return nil
	}
	// Set basic auth credentials for the POST request
	reqToAuthService.SetBasicAuth(username, password)

	// Send the POST request
	resp, err := http.DefaultClient.Do(reqToAuthService)
	if err != nil {
		SendStatus.InternalServerError(w)
		return nil
	}
	defer resp.Body.Close()

	// If the request status is not 200, write the returned status code
	if resp.StatusCode != 200 {
		w.WriteHeader(resp.StatusCode)
		return nil
	}

	// Read the JWT tokenString from the Auth service's response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		SendStatus.InternalServerError(w)
		return nil
	}
	return body
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received")
	if !IsPostRequest(w, r) { return }

	log.Println("Extracting username and password")
	username, password, ok := r.BasicAuth()
	if !ok || username == "" || password == "" {
		SendStatus.InvalidCredentials(w)
		return
	}

	log.Println("Authorizing user")
	if tokenString := AuthorizeUser(username, password, w); tokenString != nil {
		log.Println("User authorized successfully")
		// Since AuthorizeUser handles setting statusCodes, we can just write the msg body here
		w.Write(tokenString)
	} else {
		log.Println("User authorization fialed")
	}
}

// This function expects to find a Username and Password for a new user
// in the POST request's headers. If found, this information is passed onto
// the authorization service and the status code the service returns is sent
// back to the user.
func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register request received")
	if !IsPostRequest(w, r) { return }

	username := r.Header.Get("Username")
	password := r.Header.Get("Password")
	if username == "" || password == "" {
		SendStatus.BadRequest(w)
		return
	}

	url := GetAuthServiceUrl() + "/register"
	reqToAuthService, err := http.NewRequest("POST", url, nil)
	if err != nil {
		SendStatus.InternalServerError(w)
		return
	}
	reqToAuthService.Header.Add("Username", username)
	reqToAuthService.Header.Add("Password", password)

	resp, err := http.DefaultClient.Do(reqToAuthService)
	if err != nil {
		SendStatus.InternalServerError(w)
		return
	}

	if resp.StatusCode != 200 {
		w.WriteHeader(resp.StatusCode)
		return
	}

	// Read the JWT tokenString from the Auth service's response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		SendStatus.InternalServerError(w)
		return
	}
	w.Write(body)
}

func ValidateToken(r *http.Request) (jwtObject []byte, statusCode int) {
	log.Println("Validating token")
	if r.Header.Get("Authorization") == "" {
		return nil, 401
	}

	url := GetAuthServiceUrl() + "/validate"
	reqToAuthService, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, 500
	}
	reqToAuthService.Header.Set("Authorization", r.Header.Get("Authorization"))

	resp, err := http.DefaultClient.Do(reqToAuthService)
	if err != nil {
		return nil, 500
	}
	if resp.StatusCode != 200 {
		return nil, resp.StatusCode
	}
	jwtObject, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500
	}
	return jwtObject, 200
}

func Upload(w http.ResponseWriter, r *http.Request) {
	log.Println("Upload request received")
	if !IsPostRequest(w, r) { return }

	jwtObject, statusCode := ValidateToken(r)
	switch statusCode {
	case 400:
		SendStatus.BadRequest(w)
		return
	case 401:
		SendStatus.InvalidCredentials(w)
		return
	case 403:
		SendStatus.Forbidden(w)
		return
	case 500:
		SendStatus.InternalServerError(w)
		return
	}

	log.Println("Converting jwtObject to JsonStruct")
	var token JsonStruct
	err := json.Unmarshal(jwtObject, &token)
	if err != nil {
		SendStatus.InternalServerError(w)
		log.Println(err.Error())
		return
	}

	if token.Admin {
		log.Println("Getting file from request")
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Println(err.Error())
			SendStatus.BadRequest(w)
			return
		}
		defer file.Close()
		fileName := strings.Split(header.Filename, ".")
		fmt.Println("Filename was ", fileName[0])

		log.Println("Connecting to MongoDB")
		uri, err := GetMongoUri()
		FailOnError(err, "MongoUri creation failed")
		client := ConnectToMongoDB(uri)
		dbVideos := client.Database("videos")

		log.Println("Creating GridFs buckets")
		fsVideos, err := gridfs.NewBucket(dbVideos, options.GridFSBucket())
		FailOnError(err, "fsVideo creation failed")

		log.Println("Uploading file to MongoDB")
		fid, err := fsVideos.UploadFromStream(fileName[0], file)
		FailOnError(err, "Video upload to MongoDB failed")

		log.Println("Connecting to RabbitMQ")
		connection := ConnectToRabbitMQ()
		defer connection.Close()
		channel := OpenChannel(connection)
		defer channel.Close()

		log.Println("Creating RabbitMQMessage JSON")
		rabbitMqMessage := RabbitMQMessage{
			VideoFid: fid.Hex(),
			Mp3Fid: "",
			Username: token.Username,
		}
		body, err := json.Marshal(rabbitMqMessage)
		FailOnError(err, "Creating RabbitMQMessage JSON failed")

		log.Println("Publishing JSON with FID to Mp3 queue")
		err = channel.Publish(
			"",						// Exchange
			os.Getenv("VIDEO_QUEUE"),// Routing key
			false,					// Mandatory
			false,					// Immediate
			amqp.Publishing{		// Msg
				ContentType: "application/json",
				Body: body,
			},
		)
		FailOnError(err, "RabbitMQ message publishing failed")

		log.Println("File uploaded with fid:", fid.Hex())
	}
}

func Download(w http.ResponseWriter, r *http.Request) {

}

func main() {
	log.Println("Gateway service starting...")

	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/upload", Upload)
	http.HandleFunc("/download", Download)

	log.Println("Gateway service running on port", servicePort)
	err := http.ListenAndServe(":"+servicePort, nil)
	if err != nil { log.Fatal(err.Error()) }
}