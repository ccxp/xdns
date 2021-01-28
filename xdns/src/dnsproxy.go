package main

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

var useVpn = false
var dnsListenAddr string
var dnsForward string
var dnsForwardVpn string
var dnsSrv *DNSService
var vpnIfname string

// dns

// DNSService is an implementation of DNSServer interface.
type DNSService struct {
	localConn     *net.UDPConn
	remoteConn    *net.UDPConn
	memo          dnsRequestPool
	book          dnsStore
	vpnBook       dnsStore
	forwarders    []*net.UDPAddr
	forwardersVpn []*net.UDPAddr
}

func resolveDnsAddr(s string) (*net.UDPAddr, error) {
	if strings.Index(s, ":") < 0 {
		s += ":53"
	}

	return net.ResolveUDPAddr("udp", s)
}

// Listen starts a DNS server on port 53
func (s *DNSService) Listen() {

	s.memo.data = make(map[uint16]*net.UDPAddr)
	s.book.data = make(map[string]dnsEntry)
	s.vpnBook.data = make(map[string]dnsEntry)

	for _, addr := range strings.Split(dnsForward, ",") {
		a, err := resolveDnsAddr(addr)
		if err != nil {
			log.Fatalf("resolve %v fail: %v", addr, err)
		}
		s.forwarders = append(s.forwarders, a)
	}
	for _, addr := range strings.Split(dnsForwardVpn, ",") {
		a, err := resolveDnsAddr(addr)
		if err != nil {
			log.Fatalf("resolve %v fail: %v", addr, err)
		}
		s.forwardersVpn = append(s.forwardersVpn, a)
	}

	log.Printf("dns listen on %v", dnsListenAddr)

	a, err := resolveDnsAddr(dnsListenAddr)
	if err != nil {
		log.Fatalf("resolve %v fail: %v", dnsListenAddr, err)
	}
	s.localConn, err = net.ListenUDP("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	//defer s.localConn.Close()

	s.remoteConn, err = net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatal(err)
	}
	//defer s.remoteConn.Close()

	go s.handleConn(s.remoteConn)
	s.handleConn(s.localConn)

}

func (s *DNSService) handleConn(c *net.UDPConn) {

	for {
		buf := make([]byte, 512)
		n, addr, err := c.ReadFromUDP(buf)
		log.Printf("read dns message from %v", addr)
		if err != nil {
			log.Println(err)
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				c.Close()
				return
			}
			continue
		}

		go s.handle(n, buf, addr)
		buf = nil
	}
}

func (s *DNSService) handle(n int, buf []byte, addr *net.UDPAddr) {

	m := &dnsmessage.Message{}
	err := m.Unpack(buf[0:n])
	if err != nil {
		log.Println(err)
		return
	}
	if len(m.Questions) == 0 {
		return
	}
	s.Query(addr, m, buf[0:n])
}

// Query lookup answers for DNS message.
func (s *DNSService) Query(addr *net.UDPAddr, m *dnsmessage.Message, buf []byte) {
	// got response from forwarder, send it back to client

	q := m.Questions[0]
	key := qString(q)

	var book *dnsStore = &s.book
	var forwarders []*net.UDPAddr = s.forwarders
	if useVpn {
		book = &s.vpnBook
		forwarders = s.forwardersVpn
	}

	if m.Header.Response {
		if clientAddr, ok := s.memo.get(m.ID); ok {
			log.Printf("result for %v from %v to %v", key, addr, clientAddr)
			s.sendPacket(s.localConn, buf, clientAddr)
		}

		book.set(key, m.Answers)
		return
	}

	// was checked before entering this routine
	val, ok := book.get(key)
	if ok {
		m.Answers = append(m.Answers, val...)
		m.Header.Response = true

		packed, err := m.Pack()
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("result for %v from cache to %v", key, addr)
		s.sendPacket(s.localConn, packed, addr)
		return
	}

	// forwarding
	s.memo.set(m.ID, addr)

	for _, forward := range forwarders {
		log.Printf("searching %v from %v to %v", key, addr, forward)
		s.sendPacket(s.remoteConn, buf, forward)
	}
}

func (s *DNSService) sendPacket(conn *net.UDPConn, buf []byte, addr *net.UDPAddr) {

	_, err := conn.WriteToUDP(buf, addr)
	if err != nil {
		log.Println(err)
	}

}

type dnsRequestPool struct {
	sync.Mutex
	data map[uint16]*net.UDPAddr
}

func (b *dnsRequestPool) get(key uint16) (*net.UDPAddr, bool) {
	b.Lock()
	addr, ok := b.data[key]
	if ok {
		delete(b.data, key)
	}
	b.Unlock()
	return addr, ok
}

func (b *dnsRequestPool) set(key uint16, addr *net.UDPAddr) {
	b.Lock()
	b.data[key] = addr
	b.Unlock()
}

// question to string
func qString(q dnsmessage.Question) string {
	b := make([]byte, q.Name.Length+2)
	for i := 0; i < int(q.Name.Length); i++ {
		b[i] = q.Name.Data[i]
	}
	b[q.Name.Length] = uint8(q.Type >> 8)
	b[q.Name.Length+1] = uint8(q.Type)

	return string(b)
}

// resource to string
func rString(r dnsmessage.Resource) string {
	var sb strings.Builder
	sb.Write(r.Header.Name.Data[:])
	sb.WriteString(r.Header.Type.String())
	sb.WriteString(r.Body.GoString())

	return sb.String()
}

type dnsStore struct {
	sync.RWMutex
	data map[string]dnsEntry
}

type dnsEntry struct {
	Resources []dnsmessage.Resource
	TTL       uint32
	Created   int64
}

func (s *dnsStore) get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	now := time.Now().Unix()
	if e.TTL > 1 && (e.Created+int64(e.TTL) < now) {
		s.remove(key)
		return nil, false
	}
	return e.Resources, ok
}

func (s *dnsStore) set(key string, resources []dnsmessage.Resource) {

	e := dnsEntry{
		Resources: resources,
		Created:   time.Now().Unix(),
	}
	if len(resources) > 0 {
		e.TTL = resources[0].Header.TTL
	}

	s.Lock()
	s.data[key] = e
	s.Unlock()
}

func (s *dnsStore) remove(key string) {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
}

func (s *dnsStore) clear() {
	s.Lock()
	s.data = make(map[string]dnsEntry)
	s.Unlock()
}
