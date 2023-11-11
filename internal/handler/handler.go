package handler

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sacloud/iaas-api-go"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		err := ctx.Err()
		if err == nil {
			return
		}

		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func wrap(fn func(ctx *gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := fn(ctx); err != nil {
			_ = ctx.Error(err)
		}
	}
}

type ServiceDiscovery struct {
	sacloudClient *iaas.Client
	zone          string
}

func NewServiceDiscovery(sacloudClient *iaas.Client, zone string) *ServiceDiscovery {
	return &ServiceDiscovery{
		sacloudClient: sacloudClient,
		zone:          zone,
	}
}

type discoveryItem struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}
type discoveryResponse []*discoveryItem

func (h *ServiceDiscovery) DiscoveryServer() gin.HandlerFunc {
	return wrap(h.discoveryServer)
}

func (h *ServiceDiscovery) discoveryServer(ctx *gin.Context) error {
	serverOp := iaas.NewServerOp(h.sacloudClient)
	findRes, err := serverOp.Find(ctx, h.zone, &iaas.FindCondition{})
	if err != nil {
		return err
	}

	servers := []*iaas.Server{}
	for _, s := range findRes.Servers {
		if isScrape(s) {
			servers = append(servers, s)
		}
	}

	res := discoveryResponse{}
	for _, s := range servers {
		ifaces := s.GetInterfaces()
		if len(ifaces) == 0 {
			log.Println("NOTICE: server doesn't have interface")
			continue
		}
		iface := ifaces[0]
		var ip string
		if iface.IPAddress != "" {
			ip = iface.IPAddress
		}
		if iface.UserIPAddress != "" {
			ip = iface.UserIPAddress
		}

		if ip == "" {
			log.Println("NOTICE: interface doesn't set ip address")
			continue
		}

		res = append(res, &discoveryItem{
			Targets: []string{net.JoinHostPort(ip, getPort(s))},
			Labels: map[string]string{
				"__meta_server_id":   s.ID.String(),
				"__meta_server_name": s.Name,
				"__meta_server_zone": h.zone,
			},
		})
	}
	ctx.JSON(http.StatusOK, res)
	return nil
}

func (h *ServiceDiscovery) DiscoveryLoadbalancer() gin.HandlerFunc {
	return wrap(h.discoveryLoadbalancer)
}

func (h *ServiceDiscovery) discoveryLoadbalancer(ctx *gin.Context) error {
	loadbalancerOp := iaas.NewLoadBalancerOp(h.sacloudClient)
	findRes, err := loadbalancerOp.Find(ctx, h.zone, &iaas.FindCondition{})
	if err != nil {
		return err
	}

	lbs := []*iaas.LoadBalancer{}
	for _, lb := range findRes.LoadBalancers {
		if isScrape(lb) {
			lbs = append(lbs, lb)
		}
	}

	res := discoveryResponse{}
	for _, lb := range lbs {
		excludes := getExcludes(lb, "sd/exclude")
		port := getPort(lb)
		for _, vip := range lb.VirtualIPAddresses {
			if isExclude(excludes, net.JoinHostPort(vip.VirtualIPAddress, vip.Port.String())) {
				continue
			}
			vipPort := net.JoinHostPort(vip.VirtualIPAddress, vip.Port.String())
			item := &discoveryItem{
				Targets: []string{},
				Labels: map[string]string{
					"__meta_loadbalancer_id":       lb.ID.String(),
					"__meta_loadbalancer_name":     lb.Name,
					"__meta_loadbalancer_zone":     h.zone,
					"__meta_loadbalancer_vip_port": vipPort,
				},
			}
			for _, sv := range vip.Servers {
				if sv.Enabled {
					item.Targets = append(item.Targets, net.JoinHostPort(sv.IPAddress, port))
				}
			}
			res = append(res, item)
		}
	}
	ctx.JSON(http.StatusOK, res)
	return nil
}
