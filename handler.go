package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"net/url"
)

type GinHandler struct {
	httpClient     *http.Client
	svcConf        *ServiceConfigModel
	srcIPGetter    func() *net.IP
	upstreamGetter func() *url.URL
}

func (h *GinHandler) initSrcIPGetter() func() (ipNet *net.IP) {
	var ips_ []*net.IP
	for _, ipStr := range h.svcConf.SourceAddresses {
		ip_, _, err := net.ParseCIDR(ipStr)
		if err != nil {
			ip_ = net.ParseIP(ipStr)
			if ip_ == nil {
				log.Error("invalid source address:", ipStr)
				continue
			}
		}
		ips_ = append(ips_, &ip_)
	}
	if len(ips_) == 0 {
		return nil
	}
	h.srcIPGetter = IterWithControlBit(ips_, true)
	return h.srcIPGetter
}

func (h *GinHandler) initUpstreamGetter() func() (upstream *url.URL) {
	var upstreams_ []*url.URL
	for _, endpoint := range h.svcConf.UpstreamEndpoints {
		upstream_, err := url.Parse(endpoint)
		if err != nil {
			log.Errorf("invalid upstream endpoint: %v", err)
			continue
		}
		upstreams_ = append(upstreams_, upstream_)
	}
	if len(upstreams_) == 0 {
		panic("no upstreams")
	}
	h.upstreamGetter = IterWithControlBit(upstreams_, false)
	return h.upstreamGetter
}

func NewGinHandler(conf *ServiceConfigModel) (h *GinHandler) {
	h = &GinHandler{
		svcConf: conf,
	}
	h.initSrcIPGetter()
	if h.srcIPGetter == nil {
		h.httpClient = &http.Client{}
	} else {
		h.httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: nil,
				DialContext: func(ctx context.Context, network string, addr string) (
					con net.Conn, err error) {

					srcIP_ := h.srcIPGetter()
					var dialer_ *net.Dialer
					if network == "tcp" {
						addr4Dialer_ := &net.TCPAddr{
							IP:   *srcIP_,
							Port: 0,
							Zone: "",
						}
						dialer_ = &net.Dialer{
							LocalAddr: addr4Dialer_,
						}
					} else if network == "udp" {
						addr4Dialer_ := &net.UDPAddr{
							IP:   *srcIP_,
							Port: 0,
							Zone: "",
						}
						dialer_ = &net.Dialer{
							LocalAddr: addr4Dialer_,
						}
					}
					log.Infof("dialing: %s %s <-> %s", network, srcIP_.String(), addr)
					con, err = dialer_.DialContext(ctx, network, addr)
					return
				},
			},
		}
	}
	h.initUpstreamGetter()
	return
}

func (h *GinHandler) HandleFun(c *gin.Context) {
	upStreamURL_ := h.upstreamGetter()
	newReq_ := &http.Request{
		Body:   c.Request.Body,
		Method: c.Request.Method,
		Header: c.Request.Header,
	}
	newReq_.URL, _ = url.Parse(c.Request.URL.String())
	newReq_.URL.Host = upStreamURL_.Host
	newReq_.URL.Scheme = upStreamURL_.Scheme
	newReq_.Header.Del("Host")
	newReq_.Header.Del("Referer")
	newReq_.Header.Del("Remote-Addr")
	newReq_.Header.Del("X-Real-IP")
	newReq_.Header.Del("X-Forwarded-For")
	newReq_.Header.Del("X-Forwarded-Proto")
	newReq_.Header.Del("X-Forwarded-Host")
	newReq_.Header.Del("X-Forwarded-Port")
	newReq_.Header.Del("X-Forwarded-Server")
	newReq_.Header.Del("X-Forwarded-Prefix")
	newReq_.Header.Del("X-Forwarded-Ssl")
	newReq_.Header.Del("X-Forwarded-Uri")
	newReq_.Header.Del("X-Forwarded-Scheme")
	rsp_, err := h.httpClient.Do(newReq_)
	defer func() {
		if rsp_ != nil {
			_ = rsp_.Body.Close()
		}
	}()
	if err != nil {
		log.Error(c.Error(err))
		return
	}
	c.DataFromReader(rsp_.StatusCode,
		rsp_.ContentLength,
		rsp_.Header.Get("Content-Type"),
		rsp_.Body,
		nil)
}
