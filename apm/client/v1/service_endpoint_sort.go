package client

import "github.com/CloudDetail/apo-module/model/v1"

type byServiceMutatedValue []*serviceEndPoint

func (m byServiceMutatedValue) Len() int {
	return len(m)
}

func (m byServiceMutatedValue) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m byServiceMutatedValue) Less(i, j int) bool {
	return m[i].MutatedValue > m[j].MutatedValue
}

type byMuatedValue []*model.TraceTreeNode

func (m byMuatedValue) Len() int {
	return len(m)
}

func (m byMuatedValue) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m byMuatedValue) Less(i, j int) bool {
	return m[i].MutatedValue > m[j].MutatedValue
}
