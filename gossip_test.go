package gossip

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const testdata = "./testdata"

func dirsEqual(dir1, dir2 string) bool {
	out, err := exec.Command("diff", "-r", dir1, dir2).Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			log.Println(string(out))
			return false
		default:
			panic(err)
		}
	}
	return true
}

func TestGenerateSite(t *testing.T) {
	td, err := ioutil.TempDir("", "gossip_test")
	defer os.RemoveAll(td)
	if err != nil {
		panic(err)
	}
	site := NewSite(filepath.Join(testdata, "src"), td)
	if err = site.Generate(); err != nil {
		t.Errorf("error generating site: %v", err)
	}
	if !dirsEqual(td, filepath.Join(testdata, "dest")) {
		t.Errorf("expected generated site doesn't match actual")
	}
}
