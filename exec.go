package main

import (
	"golang.org/x/sys/unix"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func getPidForRunningContainer(containerID string) int {
	containers, err := getRunningContainers()
	if err != nil {
		log.Fatalf("Unable to get running containers: %v\n", err)
	}
	for _, container := range containers {
		if container.containerId == containerID {
			return container.pid
		}
	}
	return 0
}

func execInContainer(containerId string)  {
	pid := getPidForRunningContainer(containerId)
	if pid == 0 {
		log.Fatalf("No such container!")
	}
	baseNsPath := "/proc/" + strconv.Itoa(pid) + "/ns"
	ipcFd, ipcErr := os.Open(baseNsPath + "/ipc")
	mntFd, mntErr := os.Open(baseNsPath + "/mnt")
	netFd, netErr := os.Open(baseNsPath + "/net")
	pidFd, pidErr := os.Open(baseNsPath + "/pid")
	utsFd, utsErr := os.Open(baseNsPath + "/uts")

	if ipcErr != nil || mntErr != nil || netErr != nil ||
		pidErr != nil || utsErr != nil {
		log.Fatalf("Unable to open namespace files!")
	}
	syscall.Unshare(syscall.CLONE_NEWIPC)
	syscall.Unshare(syscall.CLONE_NEWNS)
	syscall.Unshare(syscall.CLONE_NEWNET)
	syscall.Unshare(syscall.CLONE_NEWPID)
	syscall.Unshare(syscall.CLONE_NEWUTS)

	unix.Setns(int(ipcFd.Fd()), syscall.CLONE_NEWIPC)
	unix.Setns(int(mntFd.Fd()), syscall.CLONE_NEWNS)
	unix.Setns(int(netFd.Fd()), syscall.CLONE_NEWNET)
	unix.Setns(int(pidFd.Fd()), syscall.CLONE_NEWPID)
	unix.Setns(int(utsFd.Fd()), syscall.CLONE_NEWUTS)

	containerMntPath := getGockerContainersPath() + "/" + containerId + "/fs/mnt"
	doOrDieWithMsg(syscall.Chroot(containerMntPath), "Unable to chroot")
	os.Chdir("/")
	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	doOrDieWithMsg(cmd.Run(), "Unable to exec command in container")
}
