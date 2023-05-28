package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type ClientData struct {
	ClientId     string
	DataSize     int64
	DateUploaded time.Time
}

var clientDataList = make(map[string]*ClientData)

func main() {
	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("templates/*")

	// Serve the files under /public/ on route /client/
	r.Static("/download", "./public")

	// Load existing client files on server startup
	dirs, err := ioutil.ReadDir("./public")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		if d.IsDir() {
			files, err := ioutil.ReadDir("./public/" + d.Name())
			if err != nil {
				log.Fatal(err)
			}

			for _, f := range files {
				clientDataList[d.Name()] = &ClientData{
					ClientId:     d.Name(),
					DataSize:     f.Size(),
					DateUploaded: f.ModTime(),
				}
			}
		}
	}

	// Routes
	r.GET("/", handleGetClientData)
	r.POST("/", handlePostFile)
	// Authentication
	r.GET("/login", showLoginPage)
	r.POST("/login", handleLogin)
	r.GET("/logout", handleLogout)

	// View Files for Client
	r.GET("/client/:clientId", handleGetHostData)

	r.Run() // listen and serve on 0.0.0.0:8080
}

func showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// TODO: Validate the username and password against your database
	if username == "admin" && password == "admin" {
		session := sessions.Default(c)
		session.Set("user", username)
		session.Save()
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

func handleLogout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		// User is already logged out
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	// Clear session
	session.Clear()
	session.Save()
	c.Redirect(http.StatusSeeOther, "/login")
}

func handleGetHostData(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")

	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	clientId := c.Param("clientId")

	dirs, err := ioutil.ReadDir("public/" + clientId)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to read client directory")
		return
	}

	var subdirs []string
	for _, dir := range dirs {
		if dir.IsDir() {
			subdirs = append(subdirs, dir.Name())
		}
	}

	c.HTML(http.StatusOK, "client.html", gin.H{
		"ClientId": clientId,
		"subdirs":  subdirs,
	})
}

func handleGetClientData(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")

	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	var dataList []ClientData
	dirs, err := ioutil.ReadDir("./public")
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to read directory")
		return
	}
	for _, d := range dirs {
		if d.IsDir() {
			files, err := ioutil.ReadDir("./public/" + d.Name())
			if err != nil {
				c.String(http.StatusInternalServerError, "Unable to read client directory")
				return
			}
			for _, f := range files {
				if v, ok := clientDataList[d.Name()]; ok && v.DataSize == f.Size() {
					dataList = append(dataList, *v)
				}
			}
		}
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"dataList": dataList,
	})
}

func handlePostFile(c *gin.Context) {
	clientId := c.GetHeader("ClientId")
	if clientId == "" {
		c.String(http.StatusBadRequest, "ClientId header missing")
		return
	}

	hostId := c.GetHeader("HostId")
	if hostId == "" {
		c.String(http.StatusBadRequest, "HostId header missing")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving file")
		return
	}

	os.MkdirAll("public/"+clientId, os.ModePerm)

	out, err := os.Create(fmt.Sprintf("%s/%s", "public/"+clientId, file.Filename))
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to create file")
		return
	}
	defer out.Close()

	fh, _ := file.Open()
	defer fh.Close()

	size, _ := io.Copy(out, fh)

	clientDataList[clientId] = &ClientData{
		ClientId:     clientId,
		DataSize:     size,
		DateUploaded: time.Now(),
	}

	c.String(http.StatusOK, "File uploaded successfully")
}
