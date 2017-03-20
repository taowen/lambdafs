package lambdafs

import (
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"path/filepath"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"time"
	"io/ioutil"
	"os"
	"sync"
	"github.com/hanwen/go-fuse/unionfs"
)

type LambdaFileSystem struct {
	UpdateFile        func(filePath string) ([]byte, error)
	tempDir           string
	origDir           string
	delegate          pathfs.FileSystem
	fileUpdatedAt     map[string]time.Time
	fileUpdatedAtLock *sync.Mutex
}

func NewLambdaFileSystem(tempDir string, origDir string, opts *unionfs.UnionFsOptions) (*LambdaFileSystem, error) {
	ufs, err := unionfs.NewUnionFsFromRoots([]string{
		tempDir/*rw*/,
		origDir/*ro*/,
	}, opts, false)
	if err != nil {
		LogError("failed to create unionfs", "err", err)
		return nil, err
	}
	lambdafs_ := &LambdaFileSystem{
		tempDir: tempDir,
		origDir: origDir,
		delegate: ufs,
		fileUpdatedAt: map[string]time.Time{},
		fileUpdatedAtLock: &sync.Mutex{},
	}
	return lambdafs_, nil
}

func (fs *LambdaFileSystem) beforeFileAccess(action string, path string) {
	if fs.UpdateFile == nil {
		return
	}
	fs.fileUpdatedAtLock.Lock()
	defer fs.fileUpdatedAtLock.Unlock()
	updatedAt, hasBeenUpdated := fs.fileUpdatedAt[path]
	roPath := filepath.Join(fs.origDir, path)
	fileInfo, err := os.Stat(roPath)
	if err != nil {
		// if file deleted from ro, it should not present in rw
		os.Remove(filepath.Join(fs.tempDir, path))
		return
	}
	if fileInfo.IsDir() {
		return
	}
	if hasBeenUpdated && !fileInfo.ModTime().After(updatedAt) {
		return
	}
	if ShouldLogDebug() {
		LogDebug("about to update file", "reason", action, "path", path)
	}
	content, err := fs.UpdateFile(filepath.Join(fs.origDir, path))
	if err != nil {
		LogError("failed to update file", "path", path, "err", err)
		return
	}
	if content == nil {
		fs.fileUpdatedAt[path] = time.Now()
		return
	}
	rwPath := filepath.Join(fs.tempDir, path)
	rwPathDir := filepath.Dir(rwPath)
	os.MkdirAll(rwPathDir, 0755) // if dir exists, the error is ignored
	err = ioutil.WriteFile(rwPath, content, 0644)
	if err != nil {
		LogError("failed to write updated file", "path", path, "err", err)
		return
	} else if ShouldLogDebug() {
		LogDebug("updated file", "path", path)
		fs.fileUpdatedAt[path] = time.Now()
	}
}

func (fs *LambdaFileSystem) StatFs(name string) *fuse.StatfsOut {
	fs.beforeFileAccess("StatFs", name)
	return fs.delegate.StatFs(name)
}

func (fs *LambdaFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
	fs.delegate.OnMount(nodeFs)
}

func (fs *LambdaFileSystem) OnUnmount() {
	fs.delegate.OnUnmount()
}

func (fs *LambdaFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	fs.beforeFileAccess("GetAttr", name)
	return fs.delegate.GetAttr(name, context)
}

func (fs *LambdaFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	return fs.delegate.OpenDir(name, context)
}

func (fs *LambdaFileSystem) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	fs.beforeFileAccess("Open", name)
	return fs.delegate.Open(name, flags, context)
}

func (fs *LambdaFileSystem) Chmod(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Chmod", path)
	return fs.delegate.Chmod(path, mode, context)
}

func (fs *LambdaFileSystem) Chown(path string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Chown", path)
	return fs.delegate.Chown(path, uid, gid, context)
}

func (fs *LambdaFileSystem) Truncate(path string, offset uint64, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Truncate", path)
	return fs.delegate.Truncate(path, offset, context)
}

func (fs *LambdaFileSystem) Readlink(name string, context *fuse.Context) (out string, code fuse.Status) {
	fs.beforeFileAccess("Readlink", name)
	return fs.delegate.Readlink(name, context)
}

func (fs *LambdaFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (code fuse.Status) {
	return fs.delegate.Mknod(name, mode, dev, context)
}

func (fs *LambdaFileSystem) Mkdir(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fs.delegate.Mkdir(path, mode, context)
}

// Don't use os.Remove, it removes twice (unlink followed by rmdir).
func (fs *LambdaFileSystem) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Unlink", name)
	return fs.delegate.Unlink(name, context)
}

func (fs *LambdaFileSystem) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Rmdir", name)
	return fs.delegate.Rmdir(name, context)
}

func (fs *LambdaFileSystem) Symlink(pointedTo string, linkName string, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Symlink", linkName)
	return fs.delegate.Symlink(pointedTo, linkName, context)
}

func (fs *LambdaFileSystem) Rename(oldPath string, newPath string, context *fuse.Context) (codee fuse.Status) {
	fs.beforeFileAccess("Rename", oldPath)
	return fs.delegate.Rename(oldPath, newPath, context)
}

func (fs *LambdaFileSystem) Link(orig string, newName string, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Link", newName)
	return fs.delegate.Link(orig, newName, context)
}

func (fs *LambdaFileSystem) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Access", name)
	return fs.delegate.Access(name, mode, context)
}

func (fs *LambdaFileSystem) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	return fs.delegate.Create(path, flags, mode, context)
}

func (fs *LambdaFileSystem) GetXAttr(name string, attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	fs.beforeFileAccess("GetXAttr", name)
	return fs.delegate.GetXAttr(name, attribute, context)
}

func (fs *LambdaFileSystem)  ListXAttr(name string, context *fuse.Context) (attributes []string, code fuse.Status) {
	fs.beforeFileAccess("ListXAttr", name)
	return fs.delegate.ListXAttr(name, context)
}

func (fs *LambdaFileSystem)  RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	fs.beforeFileAccess("RemoveXAttr", name)
	return fs.delegate.RemoveXAttr(name, attr, context)
}

func (fs *LambdaFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	fs.beforeFileAccess("SetXAttr", name)
	return fs.delegate.SetXAttr(name, attr, data, flags, context)
}

func (fs *LambdaFileSystem) SetDebug(debug bool) {
	fs.delegate.SetDebug(debug)
}

func (fs *LambdaFileSystem) String() string {
	return fs.delegate.String()
}

func (fs *LambdaFileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	fs.beforeFileAccess("Utimens", name)
	return fs.delegate.Utimens(name, Atime, Mtime, context)
}