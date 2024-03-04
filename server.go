package main

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/Rbd3178/redBlackTree/tree"
	"github.com/gin-gonic/gin"
)

func shortestNext(s string) string {
	chars := []byte(s)

	for i := len(chars) - 1; i >= 0; i-- {
		if chars[i] < 'z' {
			chars[i]++
			return string(chars[:i+1])
		}
	}

	return s + "a"
}

type link struct {
	Alias string `json:"alias"`
	URL   string `json:"url"`
}

type concurrentTree struct {
	links          tree.Tree[string, string]
	treeLock       sync.RWMutex
	pendingWriters sync.WaitGroup
	WGLock         sync.Mutex
}

var data concurrentTree

func getLinks(c *gin.Context) {
	data.pendingWriters.Wait()
	data.treeLock.RLock()
	defer data.treeLock.RUnlock()

	var links []link
	for _, pair := range data.links.InOrder() {
		links = append(links, link{Alias: pair[0].(string), URL: pair[1].(string)})
	}

	c.IndentedJSON(http.StatusOK, links)
}

func addLink(c *gin.Context) {
	var newLink link

	err := c.BindJSON(&newLink)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newLink.URL == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	parsed, err := url.Parse(newLink.URL)
	if err != nil || parsed.Scheme != "http" && parsed.Scheme != "https" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	data.WGLock.Lock()
	data.pendingWriters.Add(1)
	data.WGLock.Unlock()
	defer func() {
		data.WGLock.Lock()
		data.pendingWriters.Done()
		data.WGLock.Unlock()
	}()

	data.treeLock.Lock()
	defer data.treeLock.Unlock()

	if newLink.Alias == "" {
		maxAlias, _, _ := data.links.Max()
		newLink.Alias = shortestNext(maxAlias)
	}

	err = data.links.Insert(newLink.Alias, newLink.URL)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Alias is already taken"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newLink)
}

func main() {
	data.links.Insert("vids", "https://www.youtube.com")
	data.links.Insert("docs", "https://gobyexample.com")
	router := gin.Default()
	router.GET("/links", getLinks)
	router.POST("/links", addLink)
	router.Run("localhost:8090")
}
