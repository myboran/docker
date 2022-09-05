package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}

	setUpMount()

	// 改动,调用 exec.LookPath, 可以在系统的 PATH 里面寻找命令的绝对路径
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err = syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	log.Infof("readUserCommand: %v", msgStr)
	return strings.Split(msgStr, " ")
}

/**
Init 挂载点
*/
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	if err = pivotRoot(pwd); err != nil {
		log.Errorf("pivotRoot err:%v", err)
		return
	}

	//mount proc
	//https://blog.csdn.net/qq_37857224/article/details/125421976
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	//syscall.Mount("", "/", "", syscall.MS_SLAVE, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {
	// 为了使当前 root 的老 root 和新 root 不在同一个文件系统下, 我们把 root 重新 mount 了一次, bind mount 是把相同的内容换了一个挂载点的挂载方式
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error:%v", err)
	}
	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的 rootfs, 老的 old_root 现在挂载在 rootfs/.pivot_root 上
	// 挂载点目前依然可以在 mount 命令中看到
	fmt.Println(root, pivotDir)
	//https://github.com/xianlubird/mydocker/issues/62
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root err:%v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / err:%v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	//syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount2 pivot_root:%v dir err:%v", pivotDir, err)
	}
	return os.Remove(pivotDir)
}
