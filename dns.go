package medusa

import (
    "fmt"
    "github.com/miekg/dns"
    "net"
    "strconv"
    "strings"
)

var (
    DnsServerAddr string
    DnsServerPort string
)

type QuestionOutput struct {
    Name    string  `json:"name"`
    Type    uint16  `json:"type"`
}

type AnswerOutput struct {
    Name    string  `json:"name"`
    Type    uint16  `json:"type"`
    Ttl     uint32  `json:"TTL"`
    Data    string  `json:"data"`
}

type DnsMessageOutput struct {
    Status      int
    TC          bool
    RD          bool
    RA          bool
    AD          bool
    CD          bool
    Question    []QuestionOutput
    Answer      []AnswerOutput `json:",omitempty"`
    Comment     string  `json:",omitempty"`
    Additional  string  `json:",omitempty"`
    EDnsSubnet  string  `json:"edns_client_subnet,omitempty"`
}

func NewDnsMessageOutput() (m *DnsMessageOutput) {
    m = new(DnsMessageOutput)
    return
}

func (m *DnsMessageOutput) dnsRequest(domain, queryType, queryCd, queryEdns string) (err error) {
    request := new(dns.Msg)
    request.Id = dns.Id()
    request.RecursionDesired = true
    fqdn := dns.Fqdn(domain)

    // convert type to int16 from int or string
    // depending of the user input
    var queryType16 uint16
    queryTypeInt, err := strconv.Atoi(queryType)
    if err != nil {
        queryType16 = dns.StringToType[strings.ToUpper(queryType)]
    } else {
        queryType16 = uint16(queryTypeInt)
    }

    request.Question = make([]dns.Question, 1)
    request.Question[0] = dns.Question{fqdn, queryType16, dns.ClassINET}

    // edns_subnet handling
    if queryEdns != "" {
        eDnsSubnet := strings.Split(queryEdns, "/")
        sourceNetmask, _ := strconv.Atoi(eDnsSubnet[1])
        o := &dns.OPT{
            Hdr: dns.RR_Header{
                Name:   ".",
                Rrtype: dns.TypeOPT,
            },
        }
        addr := net.ParseIP(eDnsSubnet[0])
        // family set default to ipv4
        family := uint16(1)
        // check if it is an ipv4 or v6
        // set to v6 if not v4
        if addr.To4() == nil {
            family = uint16(2)
        }
        e := &dns.EDNS0_SUBNET{
            Code:          dns.EDNS0SUBNET,
            Address:       addr,
            Family:        family, 
            SourceNetmask: uint8(sourceNetmask),
        }
        o.SetUDPSize(dns.DefaultMsgSize)
        o.Option = append(o.Option, e)
        request.Extra = append(request.Extra, o)
    }

    c := new(dns.Client)
    s := []string{DnsServerAddr, DnsServerPort}
    r, _, err := c.Exchange(request, strings.Join(s, ":"))

    if err != nil {
        fmt.Printf("medusa.dns: error %s\n", err)
        return
    }

    m.Question          = make([]QuestionOutput, 1)
    m.Question[0].Name  = r.Question[0].Name
    m.Question[0].Type  = r.Question[0].Qtype

    m.Status    = r.Rcode
    m.TC        = r.Truncated
    m.RD        = r.RecursionDesired
    m.RA        = r.RecursionAvailable
    m.AD        = r.AuthenticatedData
    m.CD        = r.CheckingDisabled

    if r.Rcode != dns.RcodeSuccess {
        m.Comment       = "error"
    } else {
        m.Additional    = ""
        m.EDnsSubnet    = queryEdns
        m.Answer    = make([]AnswerOutput, len(r.Answer))
        for i, answer := range r.Answer {
            h := answer.Header()
            m.Answer[i].Name    = h.Name
            m.Answer[i].Type    = h.Rrtype
            m.Answer[i].Ttl     = h.Ttl

            data := strings.Split(strings.Replace(answer.String(), "  ", " ", -1), "\t")
            m.Answer[i].Data    = data[len(data)-1]
        }
    }

    return err
}
