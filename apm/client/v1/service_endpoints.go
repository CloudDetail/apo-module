package client

import (
	"sort"

	"github.com/CloudDetail/apo-module/model/v1"
)

type serviceEndPoints struct {
	services []*serviceEndPoint
}

func newServiceEndPoints() *serviceEndPoints {
	return &serviceEndPoints{
		services: make([]*serviceEndPoint, 0),
	}
}

func (services *serviceEndPoints) getOrCreateService(serviceName string, url string) *serviceEndPoint {
	var service *serviceEndPoint
	for _, serviceEndPoint := range services.services {
		if serviceEndPoint.ServiceName == serviceName && serviceEndPoint.EndPoint == url {
			service = serviceEndPoint
			break
		}
	}
	if service == nil {
		service = &serviceEndPoint{
			ServiceName: serviceName,
			EndPoint:    url,
		}
		services.services = append(services.services, service)
	}
	return service
}

func (services *serviceEndPoints) addMutateSpan(traceSpan *model.TraceTreeNode) {
	if traceSpan.MutatedValue <= 0 {
		return
	}

	service := services.getOrCreateService(traceSpan.ServiceName, traceSpan.Url)
	service.SelfTime += traceSpan.SelfTime
	service.MutatedValue += traceSpan.MutatedValue
	service.Spans = append(service.Spans, traceSpan)
}

type serviceEndPoint struct {
	ServiceName  string
	EndPoint     string
	MutatedValue int64
	SelfTime     uint64
	Spans        []*model.TraceTreeNode
}

func (service *serviceEndPoint) getMutateNode() (*model.TraceTreeNode, bool) {
	sort.Sort(byMuatedValue(service.Spans))

	top := service.Spans[0]
	if top.IsProfiled {
		return top, true
	}

	// Case - 0.3 0.3 0.3, take any noe.
	// Case - 0.5 0.3 0.2, take first.
	// Case - 0.8 0.2, take first.
	// Case - 0.5 0.4 0.1, take first or second.
	topPercent := top.SelfTime * 100 / service.SelfTime
	if topPercent >= 55 {
		// 60 40
		return top, false
	}
	var suggestDiff uint64 = 0
	if topPercent >= 40 {
		// 55 45
		// 40 30
		suggestDiff = 10
	} else if topPercent >= 20 {
		// 30 25
		// 25 20
		suggestDiff = 5
	} else {
		// 18 16
		suggestDiff = 2
	}

	for i := 1; i < len(service.Spans); i++ {
		spanNode := service.Spans[i]
		percent := spanNode.SelfTime * 100 / service.SelfTime
		if percent+suggestDiff < topPercent {
			break
		}
		if spanNode.IsProfiled {
			return spanNode, true
		}
	}
	return top, false
}
