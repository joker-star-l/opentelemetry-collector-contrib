package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8030, "The port server listens on")
}

func main() {
	r := gin.Default()
	r.PUT("/api/:db/:table/_stream_load", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Status": "Success",
		})
	})

	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
