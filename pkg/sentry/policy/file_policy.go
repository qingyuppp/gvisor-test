package policy

import (
	"errors"
	"sync"
)

type FilePermission string

const (
	PermissionReadOnly  FilePermission = "ro"
	PermissionReadWrite FilePermission = "rw"
	PermissionDeny      FilePermission = "deny"
)

// FilePolicyManager 用于管理路径权限策略
type FilePolicyManager struct {
	mu         sync.RWMutex
	permission map[string]FilePermission
}

var GlobalFilePolicyManager *FilePolicyManager

func NewFilePolicyManager() *FilePolicyManager {
	return &FilePolicyManager{
		permission: make(map[string]FilePermission),
	}
}

// SetPermission 设置指定路径的权限
func (f *FilePolicyManager) SetPermission(path string, perm FilePermission) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.permission[path] = perm
}

// GetPermission 获取指定路径的权限，默认 rw
func (f *FilePolicyManager) GetPermission(path string) FilePermission {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if perm, ok := f.permission[path]; ok {
		return perm
	}
	return PermissionReadWrite
}

// CheckFileAccess 校验文件访问权限
func (f *FilePolicyManager) CheckFileAccess(path string, isWrite bool) error {
	perm := f.GetPermission(path)
	switch perm {
	case PermissionDeny:
		return errors.New("access denied by policy")
	case PermissionReadOnly:
		if isWrite {
			return errors.New("write denied by policy (read-only)")
		}
	}
	return nil
}