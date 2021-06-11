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
	setList := util.NewSetList()
	fileCount := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() && hasExtension(info.Name(), filter) {
			fs, err := util.NewFileStruct(path)
			if err != nil {
				log.Println(err)
			}
			util.AddTwin(&setList, fs, filter)
			fileCount++
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Scanned %d files in %d directories\nCleaning...")
	setList = util.CleanSetList(setList)

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
	for _, e := range filter.ExtensionList {
		if filepath.Ext(file) == e {
			return true
		}
	}
	return false
}