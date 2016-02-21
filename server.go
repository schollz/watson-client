package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

type Watson struct {
	User        string
	Project     string
	Tags        []string
	DateTime    time.Duration
	AllTags     []string
	AllProjects []string
}

func getStatus() Watson {
	tags := getItem("watson", "tags")
	projects := getItem("watson", "projects")
	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	currentProject := "No current project."
	currentTags := []string{""}
	currentTime := time.Since(time.Now())
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("watson"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			t1, e := time.Parse(time.RFC3339, string(k))
			if e == nil {
				currentTime = time.Since(t1)
				if string(v) == ">>stop<<" {
					currentProject = "None"
					currentTags = []string{""}
				} else {
					vals := strings.Split(string(v), ",")
					currentProject = vals[0]
					if len(vals) > 1 {
						currentTags = vals[1:len(vals)]
					}
				}
			} else {
				fmt.Println(e)
			}
		}
		return nil
	})
	return Watson{"watson", currentProject, currentTags, currentTime, tags, projects}
}

func addItem(user string, name string, itemType string) {
	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		fmt.Errorf("create bucket: %s", err)
	}

	items := []string{}
	items = append(items, name)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		v := b.Get([]byte(itemType))
		if v != nil {
			items = append(items, strings.Split(string(v), ",")...)
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		err := b.Put([]byte(itemType), []byte(strings.Join(items, ",")))
		return err
	})
}

func startProject(user string, project string, tagString string) {

	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		fmt.Errorf("create bucket: %s", err)
	}

	project = strings.TrimSpace(project)
	tagString = strings.TrimSpace(tagString)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		err := b.Put([]byte(time.Now().Format(time.RFC3339)), []byte(project+","+tagString))
		return err
	})

}

func stopProject(user string) {

	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		fmt.Errorf("create bucket: %s", err)
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		err := b.Put([]byte(time.Now().Format(time.RFC3339)), []byte(">>stop<<"))
		return err
	})

}

func deleteItem(user string, name string, itemType string) {
	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		fmt.Errorf("create bucket: %s", err)
	}

	items := []string{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		v := b.Get([]byte(itemType))
		if v != nil {
			items = strings.Split(string(v), ",")
		}
		return nil
	})

	j := 0
	for i := range items {
		j = i
		if items[i] == name {
			break
		}
	}
	items = append(items[:j], items[j+1:]...)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		err := b.Put([]byte(itemType), []byte(strings.Join(items, ",")))
		return err
	})
}

func getItem(user string, itemType string) []string {
	projects := []string{}
	db, err := bolt.Open("watson.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(user))
		return err
	})
	if err != nil {
		fmt.Errorf("create bucket: %s", err)
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(user))
		v := b.Get([]byte(itemType))
		if v != nil {
			projects = strings.Split(string(v), ",")
		}
		return nil
	})
	db.Close()
	return projects
}

func main() {
	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		d := getStatus()
		fmt.Println(d)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"User":          d.User,
			"StatusMessage": "Currently: " + d.Project + " [" + strings.Join(d.Tags, ",") + "] (" + d.DateTime.String() + ")",
			"ProjectString": d.Project,
			"TagString":     template.JS(strings.Join(d.AllTags, "','")),
			"Projects":      d.AllProjects,
			"Tags":          d.AllTags,
		})
	})
	r.POST("/start", func(c *gin.Context) {
		user := strings.TrimSpace(strings.ToLower(c.PostForm("user")))
		currentProject := strings.TrimSpace(strings.ToLower(c.PostForm("currentProject")))
		tagString := strings.TrimSpace(strings.ToLower(c.PostForm("tagString")))
		startProject(user, currentProject, tagString)
		c.Redirect(302, "/")
	})
	r.POST("/stop", func(c *gin.Context) {
		user := strings.TrimSpace(strings.ToLower(c.PostForm("user")))
		stopProject(user)
		c.Redirect(302, "/")
	})
	r.POST("/add", func(c *gin.Context) {
		itemName := strings.TrimSpace(strings.ToLower(c.PostForm("itemName")))
		user := strings.TrimSpace(strings.ToLower(c.PostForm("user")))
		itemType := strings.TrimSpace(strings.ToLower(c.PostForm("itemType")))
		fmt.Println(user, itemName, itemType)
		addItem(user, itemName, itemType)
		c.Redirect(302, "/")
	})
	r.POST("/delete", func(c *gin.Context) {
		itemName := strings.TrimSpace(strings.ToLower(c.PostForm("itemName")))
		user := strings.TrimSpace(strings.ToLower(c.PostForm("user")))
		itemType := strings.TrimSpace(strings.ToLower(c.PostForm("itemType")))
		fmt.Println(user, itemName, itemType)
		deleteItem(user, itemName, itemType)
		c.Redirect(302, "/")
	})
	r.Run(":8002") // listen and server on 0.0.0.0:8080
}
