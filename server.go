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
	prefix, ok := c.GetQuery("prefix")
	var links []link

	data.pendingWriters.Wait()
	data.treeLock.RLock()
	defer data.treeLock.RUnlock()

	if !ok || prefix == "" {
		for _, pair := range data.links.InOrder() {
			links = append(links, link{Alias: pair[0].(string), URL: pair[1].(string)})
		}
	} else {
		nextPrefix := prefix[:len(prefix)-1] + string(prefix[len(prefix)-1]+1)
		for _, pair := range data.links.Range(prefix, nextPrefix) {
			links = append(links, link{Alias: pair[0].(string), URL: pair[1].(string)})
		}
	}
	
	c.IndentedJSON(http.StatusOK, links)
}

func getLinkByAlias(c *gin.Context) {
	alias := c.Param("alias")

	data.pendingWriters.Wait()
	data.treeLock.RLock()
	defer data.treeLock.RUnlock()

	url, err := data.links.At(alias)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Alias not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, link{Alias: alias, URL: url})
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

func deleteLink(c *gin.Context) {
	alias := c.Param("alias")

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

	err := data.links.Delete(alias)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Alias not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func modifyURL(c *gin.Context) {
	alias := c.Param("alias")
	var newLink link
	err := c.BindJSON(&newLink)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if newLink.Alias == "" {
		newLink.Alias = alias
	}
	if newLink.Alias != alias {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Alias cannot be changed"})
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
	err = data.links.Assign(alias, newLink.URL)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Alias not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, newLink)
}

func main() {
	data.links.Insert("vids", "https://www.youtube.com")
	data.links.Insert("docs", "https://gobyexample.com")
	router := gin.Default()
	router.GET("/links", getLinks)
	router.GET("links/:alias", getLinkByAlias)
	router.POST("/links", addLink)
	router.PATCH("/links/:alias", modifyURL)
	router.DELETE("/links/:alias", deleteLink)
	router.Run("localhost:8090")
}
