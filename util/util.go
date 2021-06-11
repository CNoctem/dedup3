package util

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Filtertype int

const (
	UNSET Filtertype = 0
	EQ    Filtertype = 1
	NEQ   Filtertype = 2
)

type Filter struct {
	CheckSize     Filtertype
	CheckDate     Filtertype
	ExtensionList []string
}

func NewFilter(checkSize, checkDate Filtertype, extensionList []string) Filter {
	return Filter{CheckSize: checkSize, CheckDate: checkDate, ExtensionList: extensionList}
}

type Set struct {
	Elements map[FileStruct]bool
}

func (s Set) Sneak() (*FileStruct, error) {
	for k, _ := range s.Elements {
		return &k, nil
	}
	return nil, errors.New("empty set")
}

func (s Set) Separate() (*FileStruct, *Set) {
	var separated FileStruct
	remainderSet := newEmptySet()
	separationDone := false
	for f, _ := range s.Elements {
		if !separationDone {
			separated = f
			separationDone = true
		} else {
			addFile(&remainderSet, f)
		}
	}
	return &separated, &remainderSet
}

func (s Set) String() string {
	toString := make([]string, 0)
	for f, _ := range s.Elements {
		toString = append(toString, f.FilePath)
	}
	return "[" + strings.Join(toString, ", ") + "]"
}

type SetList struct {
	Sets []Set
}

type FileStruct struct {
	Info       os.FileInfo
	FilePath   string
	ParentPath string
}

func (ft FileStruct) String() string {
	return ft.FilePath
}

func NewSetList() SetList {
	return SetList{make([]Set, 0)}
}

func CleanSetList(sl SetList) SetList {
	cleanList := NewSetList()
	for _, s := range sl.Sets {
		if len(s.Elements) > 1 {
			addSet(&cleanList, s)
		}
	}
	return cleanList
}

func NewFileStruct(fullPath string) (FileStruct, error) {
	fullPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("creating file", fullPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return FileStruct{}, nil
	}
	if !info.Mode().IsRegular() {
		return FileStruct{}, errors.New("not a regular file")
	}
	return FileStruct{info, fullPath, filepath.Dir(fullPath)}, nil
}

func AddTwin(setList *SetList, fileStruct FileStruct, filter Filter) {
	for _, set := range setList.Sets {
		if set.containsTwin(fileStruct, filter) {
			fmt.Println("twins:", fileStruct.FilePath, "->", set)
			addFile(&set, fileStruct)
			return
		}
	}
	addSet(setList, newSet(fileStruct))
}

func (sl SetList) NumSets() int {
	return len(sl.Sets)
}

func (sl SetList) NumElements() int {
	e := 0
	for _, s := range sl.Sets {
		e += len(s.Elements)
	}
	return e
}

func (s Set) containsTwin(fileStruct FileStruct, f Filter) bool {
	for k, _ := range s.Elements {
		if isTwin(k, fileStruct, f) {
			return true
		}
	}
	return false
}

func addFile(set *Set, fileStruct FileStruct) {
	set.Elements[fileStruct] = true
}

func addSet(setList *SetList, set Set) {
	setList.Sets = append(setList.Sets, set)
}

func isTwin(f1, f2 FileStruct, f Filter) bool {
	if f1.FilePath == f2.FilePath {
		return false
	}
	if f.CheckDate == EQ && f1.Info.ModTime() != f2.Info.ModTime() {
		return false
	}
	if f.CheckSize == EQ && f1.Info.Size() != f2.Info.Size() {
		return false
	}
	if f.CheckDate == NEQ && f1.Info.ModTime() == f2.Info.ModTime() {
		return false
	}
	if f.CheckSize == NEQ && f1.Info.Size() == f2.Info.Size() {
		return false
	}
	return f1.Info.Name() == f2.Info.Name()
}

func newSet(fileStruct FileStruct) Set {
	elements := make(map[FileStruct]bool, 1)
	elements[fileStruct] = true
	return Set{elements}
}

func newEmptySet() Set {
	elements := make(map[FileStruct]bool, 1)
	return Set{elements}
}
