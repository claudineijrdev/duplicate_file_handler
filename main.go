package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type SortOption int

const (
	Descending SortOption = iota
	Ascending
)

type File struct {
	Name string
	Size int64
	Hash string
	id   int
}

func NewFile(name string, size int64, hash string) File {
	return File{name, size, hash, 0}
}

type Files []File

func (f Files) Descending() Files {
	sort.Slice(f, func(i, j int) bool {
		if f[i].Size == f[j].Size {
			if f[i].Hash == f[j].Hash {
				return f[i].Name > f[j].Name
			}
			return f[i].Hash > f[j].Hash
		}
		return f[i].Size > f[j].Size
	})
	return f
}

func (f Files) Ascending() Files {
	sort.Slice(f, func(i, j int) bool {
		if f[i].Size == f[j].Size {
			if f[i].Hash == f[j].Hash {
				return f[i].Name < f[j].Name
			}
			return f[i].Hash < f[j].Hash
		}
		return f[i].Size < f[j].Size
	})
	return f
}

func (f *Files) Sort(option SortOption) Files {
	switch option {
	case Descending:
		return f.Descending()
	case Ascending:
		return f.Ascending()
	default:
		return f.Descending()
	}
}

func (f Files) MapDuplicates() Files {
	duplicates := make(map[string][]File)
	duplicatedFiles := make([]File, 0)
	for _, file := range f {
		for _, file2 := range f {
			if file.Name != file2.Name && file.Size == file2.Size && file.Hash == file2.Hash {
				duplicates[file.Hash] = append(duplicates[file.Hash], file)
				break
			}
		}
	}

	id := 1
	for value := range duplicates {
		for _, file := range duplicates[value] {
			file.id = id
			duplicatedFiles = append(duplicatedFiles, file)
			id++
		}
	}

	return duplicatedFiles
}

func (f Files) DeleteFiles(filesToDelete []int) (int, error) {
	freeSize := 0
	for _, fileNumber := range filesToDelete {
		for _, file := range f {
			if file.id == fileNumber {
				err := os.Remove(file.Name)
				if err != nil {
					return 0, err
				}
				freeSize += int(file.Size)
			}
		}
	}
	return freeSize, nil

}

func ScanFormat() string {
	var inputFormat string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter file format:")
	scanner.Scan()
	inputFormat = scanner.Text()
	return inputFormat
}

func ScanSortingOption() int {
	var sortingOption int
	fmt.Println("Size sorting option:")
	fmt.Println("1. Descending")
	fmt.Println("2. Ascending")
	for sortingOption != 1 && sortingOption != 2 {
		fmt.Println("Enter a sorting option:")
		fmt.Scan(&sortingOption)
		if sortingOption != 1 && sortingOption != 2 {
			fmt.Println("Wrong option")
		}
	}
	return sortingOption - 1
}

func ScanDuplicatesOption() bool {
	var duplicatesOption string
	fmt.Println("Check for duplicates?")
	for duplicatesOption != "yes" && duplicatesOption != "no" {
		fmt.Scan(&duplicatesOption)
		if duplicatesOption != "yes" && duplicatesOption != "no" {
			fmt.Println("Wrong option")
		}
	}
	return duplicatesOption == "yes"
}

func ScanDeleteFiles(files Files) []int {
	var optToDelete string
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	fmt.Println("Delete files?")
	for optToDelete != "yes" && optToDelete != "no" {
		fmt.Scan(&optToDelete)
		if optToDelete != "yes" && optToDelete != "no" {
			fmt.Println("Wrong option")
		}
	}

	if optToDelete == "yes" {
		var filesToDelete []int
		fmt.Println("Enter file numbers to delete:")
		for filesToDelete == nil {
			var input string
			scanner.Scan()
			input = scanner.Text()
			splitInput := strings.Split(input, " ")
			for _, num := range splitInput {
				fileNumber, err := strconv.Atoi(strings.TrimSpace(num))
				if err != nil {
					fmt.Println("Wrong format")
					break
				}

				if fileNumber > 0 && fileNumber <= len(files) {
					filesToDelete = append(filesToDelete, fileNumber)
				} else {
					filesToDelete = nil
					fmt.Println("Wrong format")
					break
				}
			}
		}
		return filesToDelete
	} else {
		return nil
	}

}

func MapFiles(dir string, extention string) (Files, error) {
	var files Files
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if extention != "" && filepath.Ext(path) != "."+extention {
				return nil
			}
			fileStat, _ := os.Stat(path)
			hash := md5.New()
			file, _ := os.Open(path)
			defer file.Close()
			if _, err := io.Copy(hash, file); err != nil {
				return err
			}
			hashInBytes := hash.Sum(nil)[:16]
			hashString := fmt.Sprintf("%x", hashInBytes)

			fileSize := fileStat.Size()
			files = append(files, NewFile(path, fileSize, hashString))
		}
		return nil
	})
	return files, err
}

func PrintDuplicates(files Files) {
	sizeControl := 0
	hashControl := ""

	for _, file := range files {
		if sizeControl != int(file.Size) {
			fmt.Printf("%v bytes\n", file.Size)
			sizeControl = int(file.Size)
		}
		if hashControl != file.Hash {
			fmt.Printf("Hash: %v\n", file.Hash)
			hashControl = file.Hash
		}
		fmt.Printf("%v. %v\n", file.id, file.Name)
	}
}

func PrintFilesSizes(files Files) {
	sizeControl := 0
	for _, file := range files {
		if sizeControl != int(file.Size) {
			fmt.Printf("%v bytes\n", file.Size)
			sizeControl = int(file.Size)
		}
		fmt.Println(file.Name)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print("Directory is not specified")
		return
	}
	dir := os.Args[1]

	inputFormat := ScanFormat()
	sortingOption := ScanSortingOption()
	files, err := MapFiles(dir, inputFormat)

	if err != nil {
		fmt.Println("Directory is not specified")
		return
	}

	files.Sort(SortOption(sortingOption))
	PrintFilesSizes(files)
	checkDuplicates := ScanDuplicatesOption()

	if checkDuplicates {
		duplicates := files.MapDuplicates()
		PrintDuplicates(duplicates)
		filesToDelete := ScanDeleteFiles(duplicates)

		if filesToDelete != nil {
			freeSize, err := duplicates.DeleteFiles(filesToDelete)
			if err != nil {
				fmt.Println("Failed to delete files")
			}

			fmt.Printf("Total freed up space: %v bytes\n", freeSize)
		}

	}

}
