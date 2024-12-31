package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/utils"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetEtcdAddr() string {
	if core.CoreConfig["etcd"].(string) == "" {
		return ""
	}
	return core.CoreConfig["etcd"].(string) + ":2379"
}

func GetEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   []string{GetEtcdAddr()},
		DialTimeout: 5 * time.Second,
	})
}

func GetWorkerInfo(nodeName string) (*Worker, error) {
	cli, err := GetEtcdClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	resp, err := cli.Get(context.Background(), fmt.Sprintf("/workers/%s", nodeName))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("worker not found")
	}

	var worker Worker
	if err := json.Unmarshal(resp.Kvs[0].Value, &worker); err != nil {
		return nil, err
	}

	return &worker, nil
}

func ForwardToWorker(worker *Worker, path string, body interface{}) (interface{}, error) {
	targetURL := fmt.Sprintf("http://%s:%s%s", worker.IP, worker.Port, path)
	return utils.ForwardRequest(targetURL, body)
}

func FindLeastLoadedNode(nodeStats map[string]int64) string {
	var targetNode string
	minTasks := int64(^uint64(0) >> 1)

	for node, count := range nodeStats {
		if count < minTasks {
			minTasks = count
			targetNode = node
		}
	}

	return targetNode
}
