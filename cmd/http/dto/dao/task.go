package dao

import (
	"context"
	"time"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/http/dto/entity"

	"gorm.io/gorm"
)

type Task struct {
	DB *gorm.DB
}

func (t *Task) WithContext(ctx context.Context) *Task {
	t.DB, _ = core.GetDB(core.DB_VRW)
	t.DB.WithContext(ctx)
	return t
}

func (t *Task) Create(taskModel *entity.Task) error {
	return t.DB.Model(&entity.Task{}).
		Create(taskModel).Error
}

func (t *Task) DeleteByJobId(jobId string) error {
	return t.DB.Model(&entity.Task{}).
		Where("job_id = ?", jobId).
		Delete(&entity.Task{}).
		Error
}

func (t *Task) FindByNodeAndPid(node, pid string) (*entity.Task, error) {
	tModel := &entity.Task{}
	err := t.DB.Model(&entity.Task{}).
		Where("node = ? AND pid = ?", node, pid).
		Find(tModel).Error
	return tModel, err
}

func (t *Task) FindByJobId(jobId string) (*entity.Task, error) {
	tModel := &entity.Task{}
	err := t.DB.Model(&entity.Task{}).
		Where("big_one = 'bigOne' AND job_id = ?", jobId).
		Find(tModel).Error
	return tModel, err
}

func (t *Task) GetMaxCount(hostName string) (int64, error) {
	var id int64
	t.DB.Model(&entity.Task{}).
		Where("node = ?", hostName).
		Where("do_once = ?", consts.NotDoOnce).
		Select("Max(id) as id").
		Scan(&id)
	return id, nil
}

func (t *Task) GetMinCount(hostName string) (int64, error) {
	var id int64
	t.DB.Model(&entity.Task{}).
		Where("node=?", hostName).
		Where("do_once = ?", consts.NotDoOnce).
		Select("Min(id) as id").
		Scan(&id)
	return id, nil
}

func (t *Task) GetList(offset, batchSize int64, hostName string) ([]entity.Task, error) {
	db := t.DB.Model(&entity.Task{})
	list := []entity.Task{}
	err := db.
		Where("id >= ?", offset).
		Where("node=?", hostName).
		Where("do_once = ?", consts.NotDoOnce).
		Limit(int(batchSize)).
		Find(&list).Error
	return list, err
}

func (t *Task) UpdateHeartBeatTime(id int64) error {
	return t.DB.Model(&entity.Task{}).
		Where("id = ?", id).
		Update("heart_beat_time", time.Now()).
		Error
}

func (t *Task) UpdatePid(id int64, pid int) error {
	return t.DB.Model(&entity.Task{}).
		Where("id = ?", id).
		Updates(entity.Task{
			HeartBeatTime: time.Now(),
			Pid:           pid,
		}).Error
}

func (t *Task) GetNodeTaskCount() (map[string]int64, error) {
	var results []struct {
		Node  string
		Count int64
	}

	err := t.DB.Model(&entity.Task{}).
		Select("node, count(*) as count").
		Where("do_once = ?", consts.NotDoOnce). // 只统计长期运行的任务
		Group("node").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Node] = r.Count
	}

	return stats, nil
}

func (t *Task) GetTargetNodeTaskCount(node string) (int64, error) {
	var count int64
	err := t.DB.Model(&entity.Task{}).
		Where("node = ?", node).
		Where("do_once = ?", consts.NotDoOnce).
		Count(&count).Error

	return count, err
}
