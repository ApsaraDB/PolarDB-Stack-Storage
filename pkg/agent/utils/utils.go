/*
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */

package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	smslog "polardb-sms/pkg/log"
	"strings"
	"syscall"
	"time"
)

const (
	TimeFormatMilli    = "2006-01-02T15:04:05.000Z07:00"
	TimeFormat1        = "20060102150405"
	CmdDefaultTimeout  = 5 * time.Second
	CmdDefaultPriority = -20
)

func ExecCommand(args string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return execPriorityCommand(ctx, args, CmdDefaultPriority)
}

func Exec(args string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return execCommand(ctx, args)
}

func execPriorityCommand(ctx context.Context, args string, priority int) (string, string, error) {
	return execCommand(ctx, fmt.Sprintf("nice %d %s", priority, args))
}

func execCommand(ctx context.Context, args string) (string, string, error) {
	var (
		stdout string
		stderr string
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	cmd := exec.CommandContext(ctx, "bash", append([]string{"-c"}, args)...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	start := time.Now()
	smslog.Debugf("command begin, cmd: %s, time: %s", args, start.Format(TimeFormatMilli))
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	end := time.Now()
	smslog.Debugf("command end, cmd: %s, exit code: %d, stdout: %s, stderr: %s, err: %v, time: %s, cost: %dms",
		cmd, cmd.ProcessState.ExitCode(), stdout, stderr, err, end.Format(TimeFormatMilli), end.Sub(start).Milliseconds())
	return stdout, stderr, err
}

type FileLock struct {
	file *os.File
	lock *syscall.Flock_t
}

func LockPath(path string) (*FileLock, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0x700)
	if err != nil {
		return nil, fmt.Errorf("failed to create pid file %s, err: %s", path, err)
	}

	l := syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Start:  0,
		Len:    0,
		Whence: io.SeekCurrent,
	}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &l); err != nil {
		return nil, fmt.Errorf("failed to acquire lock for pid file %s, err: %s", path, err)
	}

	pid := os.Getpid()
	// TODO check write bytes
	if _, err := syscall.Write(int(f.Fd()), []byte(fmt.Sprintf("%d", pid))); err != nil {
		return nil, fmt.Errorf("failed to write pid %d to file %s", pid, path)
	}
	smslog.Infof("successfully lock pid file %s", path)

	return &FileLock{file: f, lock: &l}, nil
}

func UnlockPath(f *FileLock) error {
	path := f.file.Name()

	if err := syscall.FcntlFlock(f.file.Fd(), syscall.F_UNLCK, f.lock); err != nil {
		return fmt.Errorf("unlock pid file %s failed, err: %s", path, err)
	}

	if err := f.file.Close(); err != nil {
		return fmt.Errorf("failed to close pid file %s", path)
	}

	if err := os.Remove(f.file.Name()); err != nil {
		return fmt.Errorf("failed to remove pid file %s", path)
	}

	smslog.Infof("successfully unlock and remove pid file %s", path)
	return nil
}

func CheckNvmeVolume(volumeName string) bool {
	var (
		err            error
		defaultTimeout = 5 * time.Second
	)

	findNvme := fmt.Sprintf("nvme list|grep %s", volumeName)
	outInfo, errInfo, err := ExecCommand(findNvme, defaultTimeout)
	if err != nil || len(errInfo) != 0 || len(outInfo) == 0 {
		smslog.Warnf("Checkout NVMe-of failed, errInfo: %v, err: %v", errInfo, err)
		return false
	}
	return true
}

func CheckNvmeVolumeStartWith3(volumeName string) bool {
	realVolumeName := strings.TrimPrefix(volumeName, "3")
	return CheckNvmeVolume(realVolumeName)
}
