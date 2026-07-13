package domain

type DNSRecordType string

const (
	RecordTXT   DNSRecordType = "TXT"
	RecordMX    DNSRecordType = "MX"
	RecordCNAME DNSRecordType = "CNAME"
)

type DNSRecord struct {
	Type  DNSRecordType
	Host  string
	Value string
	TTL   int
}