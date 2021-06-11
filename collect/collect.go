package collect

import (
	"dedup3/util"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Collect(root, collection string, filter util.Filter) {
	setList := List(root, filter)

	fmt.Println("Collecting...")
	for _, s := range setList.Sets {
		twinOne, remainderSet := s.Separate()
		fmt.Println("\n--- ", twinOne.Info.Name(), "{")

		copiedFile, err := copyFile(*twinOne, collection)
		if err != nil {
			fmt.Println(err)
		}

		//remove twinOne
		//link: copiedFile -> twinOne
		err = removeAndLink(*copiedFile, *twinOne)
		if err != nil {
			log.Fatal(err)
		}

		for twin, _ := range (*remainderSet).Elements {
			fmt.Println()
			err = removeAndLink(*copiedFile, twin)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("} ", twinOne.Info.Name())
	}
}

func List(root string, filter util.Filter) util.SetList {
	fmt.Println("Scanning...")
	setList := util.NewSetList()
	fileCount := 0
	dirCount := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() && hasExtension(info.Name(), filter) {
			fs, err := util.NewFileStruct(path)
			if err != nil {
				log.Println(err)
			}
			util.AddTwin(&setList, fs, filter)
			fileCount++
		} else if info.IsDir() {
			dirCount++
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\nScanned %d files in %d directories\nFinding duplicates...\n", fileCount, dirCount)
	setList = util.CleanSetList(setList)
	fmt.Printf("Found %d duplicates in %d sets\n\n", setList.NumElements(), setList.NumSets())

	for _, s := range setList.Sets {
		fmt.Println(s)
	}

	return setList
}

func copyFile(src util.FileStruct, destDir string) (*util.FileStruct, error){
	dir, err := os.Stat(destDir)
	if err != nil {
		log.Fatal(err)
	}
	if !dir.IsDir() {
		log.Fatal("Not a dir", dir)
	}

	destDir = filepath.Join(destDir, src.Info.Name())

	bytesRead, err := ioutil.ReadFile(src.FilePath)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(destDir, bytesRead, 0755)
	if err != nil {
		return nil, err
	}
	newFile, err := util.NewFileStruct(destDir)

	if err != nil {
		return nil, err
	}

	fmt.Println("\tCopied ", src.FilePath, "->", destDir)
	fmt.Println()

	return &newFile, nil
}

func removeAndLink(copiedFile, twin util.FileStruct) error {
	err := os.Remove(twin.FilePath)
	if err != nil {
		return err
	}
	fmt.Println("\tRemoved ", twin.FilePath)

	err = os.Symlink(copiedFile.FilePath, twin.FilePath)
	if err != nil {
		return err
	}

	fmt.Println("\tLinked ", copiedFile.FilePath, "->", twin.FilePath)
	return nil
}

func hasExtension(file string, filter util.Filter) bool {
	if len(filter.ExtensionList) == 0 {
		return true
	}
	for _, e := range filter.ExtensionList {
		if filepath.Ext(file) == e {
			return true
		}
	}
	return false
}
