package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/mylxsw/mysql-diff/util"
	"github.com/pmezard/go-difflib/difflib"
)

type Stringer interface {
	String() string
}

type Diff struct {
	diff string
	save func() error
}

func (d Diff) String() string {
	return d.diff
}

func (d Diff) Save() error {
	return d.save()
}

func (d Diff) PrintAndSave() error {
	if d.diff == "" {
		return nil
	}

	fmt.Println(d.String())
	return d.Save()
}

type DiffService struct {
	tmpDir      string
	contextLine int
}

func NewDiffService(tmpDir string, contextLine int) *DiffService {
	return &DiffService{tmpDir: tmpDir, contextLine: contextLine}
}

func (ds DiffService) Diff(name string, target Stringer) Diff {
	var original []byte
	idx, _ := ioutil.ReadFile(filepath.Join(ds.tmpDir, name+".idx"))
	if string(idx) != "" {
		idxFilepath := filepath.Join(ds.tmpDir, string(idx))
		if util.FileExist(idxFilepath) {
			original, _ = ioutil.ReadFile(idxFilepath)
		}
	}

	targetStr := target.String()
	diffRes := ds.diff(string(idx), string(original), name+".new", targetStr)

	return Diff{
		diff: diffRes,
		save: func() error {
			targetName := fmt.Sprintf("%s.%s.stat", name, time.Now().Format("20060102150405"))
			_ = ioutil.WriteFile(filepath.Join(ds.tmpDir, targetName+".diff"), []byte(diffRes), os.ModePerm)
			_ = ioutil.WriteFile(filepath.Join(ds.tmpDir, targetName), []byte(targetStr), os.ModePerm)

			return ioutil.WriteFile(filepath.Join(ds.tmpDir, name+".idx"), []byte(targetName), os.ModePerm)
		},
	}
}

func (ds DiffService) diff(s1name, s1, s2name, s2 string) string {
	udiff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(s1),
		B:        difflib.SplitLines(s2),
		FromFile: s1name,
		ToFile:   s2name,
		Context:  ds.contextLine,
	}

	text, _ := difflib.GetUnifiedDiffString(udiff)
	return text
}
