package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	targetStr := target.String()
	originalStr, _ := ioutil.ReadFile(filepath.Join(ds.tmpDir, name+".stat"))
	diffRes := ds.diff(name+".last", string(originalStr), name+".new", targetStr)
	return Diff{
		diff: diffRes,
		save: func() error {
			_ = ioutil.WriteFile(filepath.Join(ds.tmpDir, name+".diff"), []byte(diffRes), os.ModePerm)
			return ioutil.WriteFile(filepath.Join(ds.tmpDir, name+".stat"), []byte(targetStr), os.ModePerm)
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
