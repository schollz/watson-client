package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func getStatus() (watsonStatus string, watsonTagString string, watsonProjectString string, watsonProjects []string, watsonTags []string) {
	out, err := exec.Command("watson", "status").Output()
	if err != nil {
		log.Fatal(err)
	}
	watsonStatus = string(out)
	if strings.Contains(watsonStatus, "No project started") {
		watsonTagString = ""
		watsonProjectString = ""
	} else {
		r, _ := regexp.Compile("\\[(.*?)\\]")
		watsonTagString = r.FindString(watsonStatus)
		if len(watsonTagString) > 3 {
			watsonTagString = watsonTagString[1 : len(watsonTagString)-1]
			foo := "\""
			for _, s := range strings.Split(watsonTagString, ",") {
				foo = foo + strings.ToLower(strings.TrimSpace(s)) + "\",\""
			}
			if len(foo) < 3 {
				watsonTagString = foo
			} else {
				watsonTagString = foo[0 : len(foo)-2]
			}
		} else {
			watsonTagString = ""
		}
		watsonProjectString = strings.Split(watsonStatus, " ")[1]
	}

	out, err = exec.Command("watson", "projects").Output()
	if err != nil {
		log.Fatal(err)
	}
	watsonProjects = strings.Split(string(out), "\n")
	watsonProjects = watsonProjects[0 : len(watsonProjects)-1]

	out, err = exec.Command("watson", "tags").Output()
	if err != nil {
		log.Fatal(err)
	}
	watsonTags = strings.Split(string(out), "\n")
	watsonTags = watsonTags[0 : len(watsonTags)-1]

	return

}

func main() {
	badChars := "./;'[]~!@#$%^&*()_+<>:{}?'"

	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
		fmt.Println(watsonTagString, watsonProjectString)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"StatusMessage": watsonStatus,
			"ProjectString": watsonProjectString,
			"TagString":     template.JS(watsonTagString),
			"Projects":      watsonProjects,
			"Tags":          watsonTags,
		})
	})
	r.POST("/", func(c *gin.Context) {
		switchVal := c.PostForm("switchVal")
		if switchVal == "stop" {
			out, err := exec.Command("watson", "stop").Output()
			if err != nil {
				log.Fatal(err)
			}
			watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"SuccessMessage": string(out),
				"StatusMessage":  watsonStatus,
				"ProjectString":  watsonProjectString,
				"TagString":      template.JS(watsonTagString),
				"Projects":       watsonProjects,
				"Tags":           watsonTags,
			})
		} else if switchVal == "start" {
			currentProject := c.PostForm("currentProject")
			fmt.Println(currentProject)
			tagString := c.PostForm("tagString")
			fmt.Println(tagString)
			if strings.ContainsAny(currentProject, badChars) == true && strings.ContainsAny(tagString, badChars) == true {
				watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
				c.HTML(http.StatusOK, "index.tmpl", gin.H{
					"ErrorMessage":  "Can not process project:'" + currentProject + "' and tags: '" + tagString + "'",
					"StatusMessage": watsonStatus,
					"ProjectString": watsonProjectString,
					"TagString":     template.JS(watsonTagString),
					"Projects":      watsonProjects,
					"Tags":          watsonTags,
				})
			} else if strings.ContainsAny(currentProject, badChars) == true {
				watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
				c.HTML(http.StatusOK, "index.tmpl", gin.H{
					"ErrorMessage":  "Can not process project:'" + currentProject + "'",
					"StatusMessage": watsonStatus,
					"ProjectString": watsonProjectString,
					"TagString":     template.JS(watsonTagString),
					"Projects":      watsonProjects,
					"Tags":          watsonTags,
				})
			} else if strings.ContainsAny(tagString, badChars) == true {
				watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
				c.HTML(http.StatusOK, "index.tmpl", gin.H{
					"ErrorMessage":  "Can not process tags:'" + tagString + "'",
					"StatusMessage": watsonStatus,
					"ProjectString": watsonProjectString,
					"TagString":     template.JS(watsonTagString),
					"Projects":      watsonProjects,
					"Tags":          watsonTags,
				})
			} else {
				watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags := getStatus()
				fmt.Println(watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags)
				if watsonProjectString != strings.TrimSpace(strings.ToLower(currentProject)) {
					fmt.Println("Trying to stop")
					out, err := exec.Command("watson", "stop").Output()
					fmt.Println(string(out))

					cmd := []string{"start", strings.TrimSpace(strings.ToLower(currentProject))}
					for _, s := range strings.Split(tagString, ",") {
						sNew := strings.TrimSpace(strings.ToLower(s))
						if len(sNew) > 1 {
							cmd = append(cmd, "+"+sNew)
						}
					}
					fmt.Println("[", cmd, "]")
					out, err = exec.Command("watson", cmd...).Output()
					if err != nil {
						log.Fatal(err)
					}
					watsonStatus, watsonTagString, watsonProjectString, watsonProjects, watsonTags = getStatus()
					c.HTML(http.StatusOK, "index.tmpl", gin.H{
						"SuccessMessage": string(out),
						"StatusMessage":  watsonStatus,
						"ProjectString":  watsonProjectString,
						"TagString":      template.JS(watsonTagString),
						"Projects":       watsonProjects,
						"Tags":           watsonTags,
					})
				} else {
					c.HTML(http.StatusOK, "index.tmpl", gin.H{
						"StatusMessage": watsonStatus,
						"ProjectString": watsonProjectString,
						"TagString":     template.JS(watsonTagString),
						"Projects":      watsonProjects,
						"Tags":          watsonTags,
					})

				}
			}

		}
	})
	r.Run(":8002") // listen and server on 0.0.0.0:8080
}
