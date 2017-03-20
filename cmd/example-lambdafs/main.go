// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/unionfs"
	"github.com/taowen/lambdafs"
	"io/ioutil"
	"strings"
)

func main() {
	debug := flag.Bool("debug", false, "debug on")
	portable := flag.Bool("portable", false, "use 32 bit inodes")

	entry_ttl := flag.Float64("entry_ttl", 1.0, "fuse entry cache TTL.")
	negative_ttl := flag.Float64("negative_ttl", 1.0, "fuse negative entry cache TTL.")

	delcache_ttl := flag.Float64("deletion_cache_ttl", 5.0, "Deletion cache TTL in seconds.")
	branchcache_ttl := flag.Float64("branchcache_ttl", 5.0, "Branch cache TTL in seconds.")
	deldirname := flag.String(
		"deletion_dirname", "GOUNIONFS_DELETIONS", "Directory name to use for deletions.")

	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Println("Usage:\n  unionfs MOUNTPOINT RW-DIRECTORY RO-DIRECTORY ...")
		os.Exit(2)
	}

	ufsOptions := &unionfs.UnionFsOptions{
		DeletionCacheTTL: time.Duration(*delcache_ttl * float64(time.Second)),
		BranchCacheTTL:   time.Duration(*branchcache_ttl * float64(time.Second)),
		DeletionDirName:  *deldirname,
	}
	rootDir := flag.Arg(0)
	rwDir := flag.Arg(1)
	lambdafs_, err := lambdafs.NewLambdaFileSystem(rwDir, flag.Arg(2), ufsOptions)
	if err != nil {
		lambdafs.LogError("create lambdafs failed", "err", err)
		os.Exit(1)
	}
	lambdafs_.UpdateFile = func(filePath string) ([]byte, error) {
		if !strings.HasSuffix(filePath, ".php") {
			return nil, nil
		}
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		content = append(content, []byte("\nhello\n")...)
		return content, nil
	}
	nodeFs := pathfs.NewPathNodeFs(lambdafs_, &pathfs.PathNodeFsOptions{ClientInodes: true})
	mOpts := nodefs.Options{
		EntryTimeout:    time.Duration(*entry_ttl * float64(time.Second)),
		AttrTimeout:     time.Duration(*entry_ttl * float64(time.Second)),
		NegativeTimeout: time.Duration(*negative_ttl * float64(time.Second)),
		PortableInodes:  *portable,
		Debug:           *debug,
	}
	mountState, _, err := nodefs.MountRoot(rootDir, nodeFs.Root(), &mOpts)
	if err != nil {
		log.Fatal("Mount fail:", err)
	}

	mountState.Serve()
}