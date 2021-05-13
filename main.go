package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/olahol/melody.v1"
)

type FileDetails struct {
	Name string `json:"name"`
	File string `json:"file"`
	Path string `json:"path"`
}

func main() {

	folder := "./public/upload/"

	r := gin.Default()
	m := melody.New()

	r.Static("/messenger", "./public")
	r.Static("/images/icons", "./public/images/icons")
	r.Static("/public/upload", "./public/upload")

	r.MaxMultipartMemory = 8 << 20
	r.POST("/upload", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["file[]"]

		for _, file := range files {
			log.Println(file.Filename)

			// extension := filepath.Ext(file.Filename)
			// newFileName := uuid.New().String() + extension

			newFileName := strings.ToLower(file.Filename)

			if err := c.SaveUploadedFile(file, folder+newFileName); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%d files uploaded!", len(files)),
		})
	})

	r.GET("/files", func(c *gin.Context) {

		var files []string
		var fileList []FileDetails

		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}
		for _, file := range files {

			_, filename := filepath.Split(file)

			var fileDetail FileDetails
			fileDetail.Name = filename
			fileDetail.File = strings.Replace(filepath.Ext(file), ".", "", -1)
			fileDetail.Path = file

			if len(filename) > 0 {
				fileList = append(fileList, fileDetail)
			}
		}

		c.JSON(http.StatusOK, fileList)
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, response []byte) {
		m.Broadcast(response)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// r.Run()

	o := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("messenger-golang.herokuapp.com"),
	}

	log.Fatal(autotls.RunWithManager(r, &o))
}
