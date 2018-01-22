package medusa

import (
    "strconv"
    "github.com/gin-gonic/gin"
)

func EndPointResolve(c *gin.Context) {
    queryName       := c.Query("name")
    queryType       := c.DefaultQuery("type", "1")
    queryCd         := c.DefaultQuery("cd", "false")
    queryEdns       := c.DefaultQuery("edns_client_subnet", "0.0.0.0/0")

    qtype, _ := strconv.Atoi(queryType) 
    m := NewDnsMessageOutput()
    m.dnsRequest(queryName, uint16(qtype), queryCd, queryEdns)

    c.JSON(200, m)
}
