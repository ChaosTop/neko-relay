package relay

import (
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func (s *Relay) RunWssTunnelServer(tcp, udp bool) error {
	err := s.ListenTCP()
	if err != nil {
		return err
	}
	handler := http.NewServeMux()
	if tcp {
		handler.Handle("/wstcp/", websocket.Handler(s.WssTunnelServerTcpHandle))
	}
	if udp {
		handler.Handle("/wsudp/", websocket.Handler(s.WssTunnelServerUdpHandle))
	}
	handler.Handle("/", NewRP(Config.Fakeurl, Config.Fakehost))
	s.Svr = &http.Server{Handler: handler}
	go s.Svr.ServeTLS(s.TCPListen, Config.Certfile, Config.Keyfile)
	return nil
}
func (s *Relay) WssTunnelServerTcpHandle(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame
	defer ws.Close()

	tmp, err := net.DialTimeout("tcp", s.Raddr, time.Duration(s.TCPTimeout)*time.Second)
	if err != nil {
		return
	}
	rc := tmp.(*net.TCPConn)
	defer rc.Close()
	go Copy(rc, ws, s)
	Copy(ws, rc, s)
}

func (s *Relay) WssTunnelServerUdpHandle(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame
	defer ws.Close()

	rc, err := net.DialTimeout("udp", s.Raddr, time.Duration(s.UDPTimeout)*time.Second)
	if err != nil {
		return
	}
	defer rc.Close()

	go Copy(rc, ws, s)
	Copy(ws, rc, s)
}
