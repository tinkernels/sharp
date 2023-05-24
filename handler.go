package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"net/url"
	"time"
)

type GinHandler struct {
	svcConf         *ServiceConfigModel
	transportGetter func() *http.Transport
	upstreamGetter  func() *url.URL
}

func (h *GinHandler) initTransportGetter() func() (tr *http.Transport) {
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
		return func() *http.Transport {
			return &http.Transport{
				Proxy: nil,
			}
		}
	}
	var trs_ []*http.Transport
	for _, ip := range ips_ {
		ip_ := ip
		tr_ := &http.Transport{
			Proxy:                 nil,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: func(ctx context.Context, network string, addr string) (
				con net.Conn, err error) {

				var dialer_ *net.Dialer
				if network == "tcp" {
					addr4Dialer_ := &net.TCPAddr{
						IP:   *ip_,
						Port: 0,
						Zone: "",
					}
					dialer_ = &net.Dialer{
						LocalAddr: addr4Dialer_,
						Timeout:   3 * time.Second, // 3 seconds to connect
					}
				} else if network == "udp" {
					addr4Dialer_ := &net.UDPAddr{
						IP:   *ip_,
						Port: 0,
						Zone: "",
					}
					dialer_ = &net.Dialer{
						LocalAddr: addr4Dialer_,
					}
				}
				log.Debugf("dialing: %s %s <-> %s", network, ip_.String(), addr)
				con, err = dialer_.DialContext(ctx, network, addr)
				return
			},
		}
		trs_ = append(trs_, tr_)
	}
	h.transportGetter = IterWithControlBit(trs_, true)
	return h.transportGetter
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
	h.initTransportGetter()
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
	cli_ := &http.Client{
		Transport: h.transportGetter(),
	}
	rsp_, err := cli_.Do(newReq_)
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
