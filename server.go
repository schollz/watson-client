package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	badChars := "./;'[]~!@#$%^&*()_+<>:{}?'"

	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		out, err := exec.Command("watson status").Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The status is %s\n", out)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"StatusMessage": "something...",
		})
	})
	r.POST("/start", func(c *gin.Context) {
		currentProject := c.PostForm("currentProject")
		fmt.Println(currentProject)
		tagString := c.PostForm("tagString")
		fmt.Println(tagString)
		if strings.ContainsAny(currentProject, badChars) == true && strings.ContainsAny(tagString, badChars) == true {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"ErrorMessage": "Can not process project:'" + currentProject + "' and tags: '" + tagString + "'",
			})
		} else if strings.ContainsAny(currentProject, badChars) == true {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"ErrorMessage": "Can not process project:'" + currentProject + "'",
			})
		} else if strings.ContainsAny(tagString, badChars) == true {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"ErrorMessage": "Can not process tags:'" + tagString + "'",
			})
		} else {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"SuccessMessage": "Started project:'" + currentProject + "' with tags: '" + tagString + "'",
				"StatusMessage":  "something...",
			})
		}
	})
	r.Run(":8002") // listen and server on 0.0.0.0:8080
}
