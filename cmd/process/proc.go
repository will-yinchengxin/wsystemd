package process

import (
	"fmt"
	"github.com/go-kit/kit/log/level"
	procutil "github.com/shirou/gopsutil/process"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"wsystemd/cmd/log"
	"wsystemd/cmd/utils"
)

var (
	PManager *ProcManager
)

type ProcManager struct {
	lock  sync.RWMutex
	procs map[string]int
}

func NewProcManager() *ProcManager {
	return &ProcManager{
		procs: make(map[string]int),
	}
}

func (p *ProcManager) JobExist(jobId string) (int, bool) {
	p.lock.RLock()
	pid, ok := p.procs[jobId]
	p.lock.RUnlock()
	return pid, ok
}

func (m *ProcManager) StartProc(cmd string, args []string, outfile, errfile string, jobId string) (int, error) {
	fmt.Println("StartProc", cmd, args, outfile, errfile, jobId)
	outFile, err := utils.GetFile(outfile)
	if err != nil {
		level.Error(log.Logger).Log("Err", fmt.Sprintf("utils.GetFile(outfile) Err: %s", err.Error()))
		return 0, err
	}
	errFile, err := utils.GetFile(errfile)
	if err != nil {
		level.Error(log.Logger).Log("Err", fmt.Sprintf("utils.GetFile(errfile) Err: %s", err.Error()))
		return 0, err
	}
	wd, _ := os.Getwd()
	hostName, err := GetHostName()
	if err != nil {
		level.Error(log.Logger).Log("Err", fmt.Sprintf("GetHostName() Err: %s", err.Error()))
		return 0, err
	}
	procAtr := &os.ProcAttr{
		Dir: wd,
		Env: []string{"TASK_TOKEN=" + hostName + ":" + jobId},
		Files: []*os.File{
			os.Stdin,
			outFile,
			errFile,
		},
	}
	process, err := os.StartProcess(cmd, append([]string{cmd}, args...), procAtr)
	if err != nil {
		return 0, err
	}
	m.procs[jobId] = process.Pid
	return process.Pid, nil
}

func (m *ProcManager) StopProc(jobId string, pid int, force bool) (int, error) {
	var (
		err  error
		done = make(chan error, 1)
	)
	m.stopProc(pid, force)
	go func() {
		// 造成僵尸进程的原因
		// 	wait, err := procInfo.Wait() if err != nil { level.Error(log.Logger).Log("Err", fmt.Sprintf("Failed to wait for process %d Err: %s", pid, err.Error())) continue } if wait.ExitCode() == 0 { break } 和 _, err = syscall.Wait4(pid, nil, 0, nil) 功能分别是什么 ,冲突么
		_, err = syscall.Wait4(pid, nil, 0, nil)
		done <- err
	}()
	select {
	case err = <-done:
		if err != nil && err.Error() != "no child processes" {
			level.Error(log.Logger).Log("Err", fmt.Sprintf("syscall.Wait4(to kill son process) Err: %s", err.Error()))
			return -1, nil
		}
		level.Info(log.Logger).Log("msg", fmt.Sprintf("All processes %d are killed", pid))
	case <-time.After(time.Second * 5):
		level.Error(log.Logger).Log("Err", fmt.Sprintf("process %d is not killed after 5 seconds", pid))
		return -2, nil
	}

	return 0, nil
}

func (p *ProcManager) stopProc(pid int, force bool) error {
	forceStop := func() error {
		cmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
		err := cmd.Run()
		time.Sleep(5 * time.Millisecond)
		if err != nil {
			level.Error(log.Logger).Log("Err", fmt.Sprintf("Failed to kill process %d Err: %s", pid, err.Error()))
			return err
		}
		return nil
	}
	if force {
		return forceStop()
	}
	i := 0
	for ; i < 3; i++ {
		cmd := exec.Command("kill", "-15", fmt.Sprintf("%d", pid))
		err := cmd.Run()
		if err != nil {
			level.Error(log.Logger).Log("Err", fmt.Sprintf("GracefullyStop process %d Err: %s", pid, err.Error()))
		}
		time.Sleep(5 * time.Millisecond)
	}

	return forceStop()
}

func (m *ProcManager) IsAlive(pid int) bool {
	proc, err := procutil.NewProcess(int32(pid))
	if err != nil {
		level.Error(log.Logger).Log("procutil.NewProcess Err", err.Error())
		return false
	}

	s, err := proc.Status()
	level.Info(log.Logger).Log("msg", fmt.Sprintf("Get Process %d Status %s", pid, s))
	if err != nil {
		level.Error(log.Logger).Log("proc.Status Err", err.Error())
	}
	return s != "Z" && s != "T"
}

func (m *ProcManager) DelProc(jobId string) int {
	m.lock.Lock()
	defer m.lock.Unlock()
	proc, ok := m.procs[jobId]
	if ok {
		delete(m.procs, jobId)
	}
	return proc
}
