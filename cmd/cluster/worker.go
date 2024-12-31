package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/log"
	"wsystemd/cmd/utils"

	"github.com/go-kit/kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	WkMg *WorkerManager
)

type Worker struct {
	ID        string
	Hostname  string
	IP        string
	Port      string
	Status    string
	LastBeat  time.Time
	Resources ResourceInfo
}

type ResourceInfo struct {
	CPUUsage    float64
	MemoryUsage float64
	TaskCount   int
}

type WorkerManager struct {
	etcd      *clientv3.Client
	worker    Worker
	taskCount int64
	mu        sync.RWMutex
}

type Task struct {
	ID         string
	Command    string
	Args       []string
	Status     string
	WorkerID   string
	CreateTime time.Time
}

func NewWorkerManager(endpoints []string, workerID string) (*WorkerManager, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      5 * time.Second,
		AutoSyncInterval: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	hostname, err := utils.GetHostName()
	if err != nil {
		return nil, err
	}

	ip, err := utils.GetLocalIP()
	if err != nil {
		return nil, err
	}

	WkMg = &WorkerManager{
		etcd: cli,
		worker: Worker{
			ID:       workerID,
			Hostname: hostname,
			IP:       ip,
			Port:     consts.ServerPort,
			Status:   "active",
			LastBeat: time.Now(),
		},
	}
	return WkMg, nil
}

func (wm *WorkerManager) Register(ctx context.Context) error {
	workerKey := fmt.Sprintf("/workers/%s", wm.worker.ID)
	resp, err := wm.etcd.Get(context.Background(), workerKey)
	if err != nil {
		return err
	}

	if len(resp.Kvs) > 0 {
		var workerData map[string]interface{}
		if err := json.Unmarshal(resp.Kvs[0].Value, &workerData); err != nil {
			return err
		}

		wm.worker.IP = workerData["ip"].(string)
		wm.worker.Hostname = workerData["hostname"].(string)
		wm.worker.Port = workerData["port"].(string)
		wm.worker.Status = workerData["status"].(string)
		wm.worker.LastBeat = time.Now()

		if count, ok := workerData["taskCount"].(float64); ok {
			wm.taskCount = int64(count)
		}

		level.Info(log.Logger).Log("msg", "Worker already registered, syncing data", "id", wm.worker.ID)
		return nil
	}

	workerData, _ := json.Marshal(wm.worker)
	_, err = wm.etcd.Put(context.Background(), workerKey, string(workerData))
	if err != nil {
		return err
	}

	level.Info(log.Logger).Log("msg", "Worker registered successfully", "id", wm.worker.ID)

	// 启动任务计数同步和心跳
	go wm.syncTaskCount(ctx)
	go wm.heartbeat(ctx)

	return nil
}

func (wm *WorkerManager) heartbeat(ctx context.Context) {
	workerKey := fmt.Sprintf("/workers/%s", wm.worker.ID)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			level.Info(log.Logger).Log("msg", "Worker stopping, cleaning up registration", "id", wm.worker.ID)
			_, err := wm.etcd.Delete(context.Background(), workerKey)
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to delete worker info from etcd", "error", err, "id", wm.worker.ID)
			} else {
				level.Info(log.Logger).Log("msg", "Worker registration cleaned up successfully", "id", wm.worker.ID)
			}
			return
		case <-ticker.C:
			workerData := map[string]interface{}{
				"id":        wm.worker.ID,
				"hostname":  wm.worker.Hostname,
				"ip":        wm.worker.IP,
				"port":      wm.worker.Port,
				"status":    wm.worker.Status,
				"lastBeat":  time.Now(),
				"taskCount": wm.GetTaskCount(),
			}
			data, _ := json.Marshal(workerData)

			_, err := wm.etcd.Put(context.Background(), workerKey, string(data))
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to update worker info", "error", err, "id", wm.worker.ID)
			}
		}
	}
}

func (wm *WorkerManager) IncrTaskCount() {
	wm.mu.Lock()
	wm.taskCount++
	wm.mu.Unlock()
}

func (wm *WorkerManager) DecrTaskCount() {
	wm.mu.Lock()
	wm.taskCount--
	wm.mu.Unlock()
}

func (wm *WorkerManager) GetTaskCount() int64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.taskCount
}

func (wm *WorkerManager) syncTaskCount(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count := wm.GetTaskCount()
			workerData := map[string]interface{}{
				"id":        wm.worker.ID,
				"hostname":  wm.worker.Hostname,
				"ip":        wm.worker.IP,
				"taskCount": count,
				"lastBeat":  time.Now(),
			}
			data, _ := json.Marshal(workerData)

			key := fmt.Sprintf("/workers/%s", wm.worker.ID)
			_, err := wm.etcd.Put(context.Background(), key, string(data))
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to sync task count", "error", err)
			}
		}
	}
}

func GetWorkerTaskCounts(cli *clientv3.Client) (map[string]int64, error) {
	resp, err := cli.Get(context.Background(), "/workers/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, kv := range resp.Kvs {
		var workerData map[string]interface{}
		if err := json.Unmarshal(kv.Value, &workerData); err != nil {
			continue
		}

		if count, ok := workerData["taskCount"].(float64); ok {
			hostname := workerData["hostname"].(string)
			counts[hostname] = int64(count)
		}
	}
	return counts, nil
}
