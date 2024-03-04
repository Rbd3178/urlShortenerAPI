package main

import (
	"net/http"
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

func addLink(c *gin.Context)

func main() {
	data.links.Insert("aboba", "https://www.youtube.com")
	data.links.Insert("bebra", "https://gobyexample.com")
	router := gin.Default()
	router.GET("/links", getLinks)
	router.Run("localhost:8090")
}
