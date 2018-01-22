package medusa

import (
    "fmt"
    "strings"
    "strconv"
    "net"
    //"reflect"
    "github.com/miekg/dns"
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

func (m *DnsMessageOutput) dnsRequest(domain string, queryType uint16, queryCd, queryEdns string) (err error) {
    request := new(dns.Msg)
    request.Id = dns.Id()
    request.RecursionDesired = true
    fqdn := dns.Fqdn(domain)
    request.Question = make([]dns.Question, 1)
    request.Question[0] = dns.Question{fqdn, queryType, dns.ClassINET}

    // edns_subnet handling
    eDnsSubnet := strings.Split(queryEdns, "/")
    sourceNetmask, _ := strconv.Atoi(eDnsSubnet[1])
    o := &dns.OPT{
        Hdr: dns.RR_Header{
            Name:   ".",
            Rrtype: dns.TypeOPT,
        },
    }
    e := &dns.EDNS0_SUBNET{
        Code:          dns.EDNS0SUBNET,
        Address:       net.ParseIP(eDnsSubnet[0]),
        Family:        1, // IP4
        SourceNetmask: uint8(sourceNetmask),
    }
    o.SetUDPSize(dns.DefaultMsgSize)
    o.Option = append(o.Option, e)
    request.Extra = append(request.Extra, o)

    c := new(dns.Client)
    s := []string{DnsServerAddr, DnsServerPort}
    r, _, err := c.Exchange(request, strings.Join(s, ":"))

    if err != nil {
        fmt.Printf("medusa.dns: error %s\n", err)
        return
    }

    //ttype, _ := dns.TypeToString[queryType]

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

            sla := strings.Split(strings.Replace(answer.String(), "  ", " ", -1), "\t")
            m.Answer[i].Data    = sla[len(sla)-1] 
        }
    }

    return err
}
