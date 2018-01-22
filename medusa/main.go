package main

import (
    "fmt"
    "os"
    "flag"
    "medusa"
    "github.com/gin-gonic/gin"
)

var (
    flagServer  = flag.String("s", "", "dns server address to resolve query")
    flagPort    = flag.String("p", "53", "dns server port to resolve query")
)

func usage() {
    fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", os.Args[0])
    flag.PrintDefaults()
}

func main() {

    flag.Usage = usage
    flag.Parse()

    if *flagServer == "" {
        fmt.Fprintf(os.Stderr, "server address mandatory")
        os.Exit(1)
    }

    medusa.DnsServerAddr = *flagServer
    medusa.DnsServerPort = *flagPort

    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
                "message": "pong",
        })
    })
    r.GET("/resolve", medusa.EndPointResolve)
    r.Run()
}
