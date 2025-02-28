package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

var port int

func main() {
	flag.IntVar(&port, "port", 8030, "The port server listens on")
	flag.Parse()

	r := gin.Default()
	r.PUT("/api/:db/:table/_stream_load", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Status": "Success",
		})
	})

	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
