package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
	Namee string
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuSet != "" {
			//fmt.Println("11", res.CpuSet)
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.mems"), []byte(strconv.Itoa(0)), 0644); err != nil {
				//fmt.Println("22")
				return fmt.Errorf("set cgroup cpuset fail %v", err)
			}
			//fmt.Println("33")
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
				//fmt.Println("44")
				return fmt.Errorf("set cgroup cpuset fail %v", err)
			}
			//fmt.Println("55")
		}
		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (s *CpusetSubSystem) Name() string {
	return s.Namee
}
