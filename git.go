package stacker

import (
	"os/exec"
	"strings"
)

// GitHash generates a version string similar to what git describe --always
// does, with -dirty on the end if the git repo had local changes.
func gitHash(path string, short bool) (string, error) {

	// Check if there are local changes
	args := []string{"-C", path, "status", "--porcelain", "--untracked-files=no"}
	output, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", err
	}
	isClean := len(output) == 0

	// Get hash
	args = []string{"-C", path, "rev-parse", "HEAD"}
	if short {
		args = []string{"-C", path, "rev-parse", "--short", "HEAD"}
	}
	output, err = exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", err
	}
	hash := strings.TrimSpace(string(output))

	if isClean {
		return hash, nil
	}

	return hash + "-dirty", nil

}

// GitVersion generates a version string similar to what git describe --always
// does, with -dirty on the end if the git repo had local changes.
func GitVersion(path string) (string, error) {
	return gitHash(path, false)
}

// GitUsername determines the local git username
func GitUsername(path string) (string, error) {
	output, err := exec.Command("git", "-C", path, "config", "user.email").CombinedOutput()
	if err != nil {
		return "", err
	}
	email := strings.TrimSpace(string(output))
	pieces := strings.SplitN(email, "@", 2)

	// Username could be obtained
	if len(pieces) == 2 {
		return pieces[0], nil
	}

	// Username could not be obtained, this is not necessarily an error
	return "", nil
}


// GitLayerTag version generates a commit-<if> tag to be used for uploading an image to a docker registry
func NewGitLayerTag(path string) (string, error) {
	tag := ""

	// Determine git hash
	hash, err := gitHash(path, true)
	if err != nil {
		return "", err
	}

	// Determine git username
	username, err := GitUsername(path)
	if err != nil {
		return "", err
	}

	// Set username in tag
	if len(username) != 0 {
		tag = username + "-"
	}

	// Add commit id in tag
	tag += "commit-" + hash

	return tag, nil
}