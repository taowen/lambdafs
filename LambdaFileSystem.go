package lambdafs

import (
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"path/filepath"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"time"
	"io/ioutil"
	"github.com/taowen/lambdafs/infra"
	"os"
)

type LambdaFileSystem struct {
	RootDir       string
	RwDir         string
	RoDir         string
	Delegate      pathfs.FileSystem
	UpdateFile    func(filePath string) ([]byte, error)
	FileUpdatedAt map[string]time.Time
}

func (fs *LambdaFileSystem) BeforeFileAccess(action string, path string) {
	if fs.UpdateFile == nil {
		return
	}
	updatedAt, hasBeenUpdated := fs.FileUpdatedAt[path]
	roPath := filepath.Join(fs.RoDir, path)
	fileInfo, err := os.Stat(roPath)
	if err != nil {
		// if file deleted from ro, it should not present in rw
		os.Remove(filepath.Join(fs.RwDir, path))
		return
	}
	if fileInfo.IsDir() {
		return
	}
	if hasBeenUpdated && !fileInfo.ModTime().After(updatedAt) {
		return
	}
	if infra.ShouldLogDebug() {
		infra.LogDebug("about to update file", "reason", action, "path", path)
	}
	content, err := fs.UpdateFile(filepath.Join(fs.RoDir, path))
	if err != nil {
		infra.LogError("failed to update file", "path", path, "err", err)
		return
	}
	if content == nil {
		fs.FileUpdatedAt[path] = time.Now()
		return
	}
	rwPath := filepath.Join(fs.RwDir, path)
	rwPathDir := filepath.Dir(rwPath)
	os.MkdirAll(rwPathDir, 0755) // if dir exists, the error is ignored
	err = ioutil.WriteFile(rwPath, content, 0644)
	if err != nil {
		infra.LogError("failed to write updated file", "path", path, "err", err)
		return
	} else if infra.ShouldLogDebug() {
		infra.LogDebug("updated file", "path", path)
		fs.FileUpdatedAt[path] = time.Now()
	}
}

func (fs *LambdaFileSystem) StatFs(name string) *fuse.StatfsOut {
	fs.BeforeFileAccess("StatFs", name)
	return fs.Delegate.StatFs(name)
}

func (fs *LambdaFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
}

func (fs *LambdaFileSystem) OnUnmount() {}

func (fs *LambdaFileSystem) GetPath(relPath string) string {
	return filepath.Join(fs.RootDir, relPath)
}

func (fs *LambdaFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	fs.BeforeFileAccess("GetAttr", name)
	return fs.Delegate.GetAttr(name, context)
}

func (fs *LambdaFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	return fs.Delegate.OpenDir(name, context)
}

func (fs *LambdaFileSystem) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	fs.BeforeFileAccess("Open", name)
	return fs.Delegate.Open(name, flags, context)
}

func (fs *LambdaFileSystem) Chmod(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Chmod", path)
	return fs.Delegate.Chmod(path, mode, context)
}

func (fs *LambdaFileSystem) Chown(path string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Chown", path)
	return fs.Delegate.Chown(path, uid, gid, context)
}

func (fs *LambdaFileSystem) Truncate(path string, offset uint64, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Truncate", path)
	return fs.Delegate.Truncate(path, offset, context)
}

func (fs *LambdaFileSystem) Readlink(name string, context *fuse.Context) (out string, code fuse.Status) {
	fs.BeforeFileAccess("Readlink", name)
	return fs.Delegate.Readlink(name, context)
}

func (fs *LambdaFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (code fuse.Status) {
	return fs.Delegate.Mknod(name, mode, dev, context)
}

func (fs *LambdaFileSystem) Mkdir(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fs.Delegate.Mkdir(path, mode, context)
}

// Don't use os.Remove, it removes twice (unlink followed by rmdir).
func (fs *LambdaFileSystem) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Unlink", name)
	return fs.Delegate.Unlink(name, context)
}

func (fs *LambdaFileSystem) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Rmdir", name)
	return fs.Delegate.Rmdir(name, context)
}

func (fs *LambdaFileSystem) Symlink(pointedTo string, linkName string, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Symlink", linkName)
	return fs.Delegate.Symlink(pointedTo, linkName, context)
}

func (fs *LambdaFileSystem) Rename(oldPath string, newPath string, context *fuse.Context) (codee fuse.Status) {
	fs.BeforeFileAccess("Rename", oldPath)
	return fs.Delegate.Rename(oldPath, newPath, context)
}

func (fs *LambdaFileSystem) Link(orig string, newName string, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Link", newName)
	return fs.Delegate.Link(orig, newName, context)
}

func (fs *LambdaFileSystem) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Access", name)
	return fs.Delegate.Access(name, mode, context)
}

func (fs *LambdaFileSystem) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	return fs.Delegate.Create(path, flags, mode, context)
}

func (fs *LambdaFileSystem) GetXAttr(name string, attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	fs.BeforeFileAccess("GetXAttr", name)
	return fs.Delegate.GetXAttr(name, attribute, context)
}

func (fs *LambdaFileSystem)  ListXAttr(name string, context *fuse.Context) (attributes []string, code fuse.Status) {
	fs.BeforeFileAccess("ListXAttr", name)
	return fs.Delegate.ListXAttr(name, context)
}

func (fs *LambdaFileSystem)  RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	fs.BeforeFileAccess("RemoveXAttr", name)
	return fs.Delegate.RemoveXAttr(name, attr, context)
}

func (fs *LambdaFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	fs.BeforeFileAccess("SetXAttr", name)
	return fs.Delegate.SetXAttr(name, attr, data, flags, context)
}

func (fs *LambdaFileSystem) SetDebug(debug bool) {
	fs.Delegate.SetDebug(debug)
}

func (fs *LambdaFileSystem) String() string {
	return fs.Delegate.String()
}

func (fs *LambdaFileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	fs.BeforeFileAccess("Utimens", name)
	return fs.Delegate.Utimens(name, Atime, Mtime, context)
}