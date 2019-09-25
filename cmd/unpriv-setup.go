package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/anuvu/stacker"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var unprivSetupCmd = cli.Command{
	Name:   "unpriv-setup",
	Usage:  "do the necessary unprivileged setup for stacker build to work without root",
	Action: doUnprivSetup,
	Before: beforeUnprivSetup,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "uid",
			Usage: "the user to do setup for (defaults to $SUDO_UID from env)",
			Value: os.Getenv("SUDO_UID"),
		},
		cli.StringFlag{
			Name:  "gid",
			Usage: "the group to do setup for (defaults to $SUDO_GID from env)",
			Value: os.Getenv("SUDO_GID"),
		},
	},
}

func beforeUnprivSetup(ctx *cli.Context) error {
	if ctx.String("uid") == "" {
		return fmt.Errorf("please specify --uid or run unpriv-setup with sudo")
	}

	if ctx.String("gid") == "" {
		return fmt.Errorf("please specify --gid or run unpriv-setup with sudo")
	}

	return nil
}

func recursiveChown(dir string, uid int, gid int) error {
	return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return os.Chown(p, uid, gid)
	})
}

func warnAboutNewuidmap() {
	_, err := exec.LookPath("newuidmap")
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: no newuidmap binary present. LXC will not work correctly.")
	}

	_, err = exec.LookPath("newgidmap")
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: no newgidmap binary present. LXC will not work correctly.")
	}
}

func doUnprivSetup(ctx *cli.Context) error {
	_, err := os.Stat(config.StackerDir)
	if err == nil {
		return fmt.Errorf("stacker dir %s already exists, aborting setup", config.StackerDir)
	}

	uid, err := strconv.Atoi(ctx.String("uid"))
	if err != nil {
		return errors.Wrapf(err, "couldn't convert uid %s", ctx.String("uid"))
	}

	gid, err := strconv.Atoi(ctx.String("gid"))
	if err != nil {
		return errors.Wrapf(err, "couldn't convert gid %s", ctx.String("gid"))
	}

	err = os.MkdirAll(path.Join(config.StackerDir), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(config.RootFSDir), 0755)
	if err != nil {
		return err
	}

	size := int64(100 * 1024 * 1024 * 1024)
	err = stacker.MakeLoopbackBtrfs(path.Join(config.StackerDir, "btrfs.loop"), size, uid, gid, config.RootFSDir)
	if err != nil {
		return err
	}

	err = recursiveChown(config.StackerDir, uid, gid)
	if err != nil {
		return err
	}

	err = recursiveChown(config.RootFSDir, uid, gid)
	if err != nil {
		return err
	}

	warnAboutNewuidmap()
	return nil
}
