package medusa

import (
    "github.com/gin-gonic/gin"
)

// Handle path /resolve
// - Get the url arguments
// - Make the DNS request
// - Send the result as a JSON object
func EndPointResolve(c *gin.Context) {
    queryName       := c.Query("name")
    queryType       := c.DefaultQuery("type", "1")
    queryCd         := c.DefaultQuery("cd", "false")
    queryEdns       := c.DefaultQuery("edns_client_subnet", "")

    m := NewDnsMessageOutput()
    m.dnsRequest(queryName, queryType, queryCd, queryEdns)

    c.JSON(200, m)
}
