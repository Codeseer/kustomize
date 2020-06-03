// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filesys"
)

// Cloner is a function that can clone a git repo.
type Cloner func(repoSpec *RepoSpec) error

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec) error {
	gitProgram, err := exec.LookPath("git")
	if err != nil {
		return errors.Wrap(err, "no 'git' program on path")
	}
	repoSpec.Dir, err = filesys.NewCacheConfirmedDir(repoSpec.OrgRepo + repoSpec.Ref)
	if err != nil {
		return err
	}

	if repoSpec.Ref == "" {
		repoSpec.Ref = "master"
	}

	if _, err := os.Stat(path.Join(repoSpec.Dir.String(), ".git")); os.IsNotExist(err) {
		//cache has not been populated with initial clone yet.
		cmd := exec.Command(
			gitProgram,
			"clone",
			"--depth=1",
			repoSpec.CloneSpec(),
			"-b",
			repoSpec.Ref,
			repoSpec.Dir.String())
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err = cmd.Run()
		if err != nil {
			log.Printf("Error cloning git repo: %s", out.String())
			return errors.Wrapf(
				err,
				"trouble cloning git repo %v in %s",
				repoSpec.CloneSpec(), repoSpec.Dir.String())
		}

		cmd = exec.Command(
			gitProgram,
			"submodule",
			"update",
			"--init",
			"--recursive")
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Dir = repoSpec.Dir.String()
		err = cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "trouble fetching submodules for %s", repoSpec.CloneSpec())
		}
	}

	return nil
}

// DoNothingCloner returns a cloner that only sets
// cloneDir field in the repoSpec.  It's assumed that
// the cloneDir is associated with some fake filesystem
// used in a test.
func DoNothingCloner(dir filesys.ConfirmedDir) Cloner {
	return func(rs *RepoSpec) error {
		rs.Dir = dir
		return nil
	}
}
