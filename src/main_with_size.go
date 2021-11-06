package main

import (
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/jessevdk/go-flags"
	"io/fs"
	"io/ioutil"
	"math"
	"path/filepath"
	"sort"
	"strconv"
)

type Dir struct {
	node           string
	name           string
	subdirectories []Dir
	files          []fs.FileInfo
	size           float64
}

type SubdirParams struct {
	LevelIndex int
	DirPrefix  string
	FilePrefix string
}

func RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
} //Thanks for https://www.socketloop.com/tutorials/golang-byte-format-example

func ByteFormat(inputNum float64, precision int) string {

	if precision <= 0 {
		precision = 1
	}

	var unit string
	var returnVal float64
	CurrColor := color.Reset
	if inputNum >= 1000000000000000000000000 {
		returnVal = RoundUp((inputNum / 1208925819614629174706176), precision)
		unit = " YB" // yottabyte
		CurrColor = color.Purple

	} else if inputNum >= 1000000000000000000000 {
		returnVal = RoundUp((inputNum / 1180591620717411303424), precision)
		unit = " ZB" // zettabyte
		CurrColor = color.Purple

	} else if inputNum >= 10000000000000000000 {
		returnVal = RoundUp((inputNum / 1152921504606846976), precision)
		unit = " EB" // exabyte
		CurrColor = color.Purple

	} else if inputNum >= 1000000000000000 {
		returnVal = RoundUp((inputNum / 1125899906842624), precision)
		unit = " PB" // petabyte
		CurrColor = color.Purple

	} else if inputNum >= 1000000000000 {
		returnVal = RoundUp((inputNum / 1099511627776), precision)
		unit = " TB" // terrabyte
		CurrColor = color.Purple
	} else if inputNum >= 1000000000 {
		returnVal = RoundUp((inputNum / 1073741824), precision)
		unit = " GB" // gigabyte
		CurrColor = color.Red
	} else if inputNum >= 1000000 {
		returnVal = RoundUp((inputNum / 1048576), precision)
		unit = " MB" // megabyte
		CurrColor = color.White
	} else if inputNum >= 1000 {
		returnVal = RoundUp((inputNum / 1024), precision)
		unit = " KB"
		CurrColor = color.Gray
	} else {
		returnVal = inputNum
		unit = " bytes" // byte
	}
	if Options.Colorize == "yes" {
		return CurrColor + strconv.FormatFloat(returnVal, 'f', precision, 64) + unit + color.Reset
	}
	return strconv.FormatFloat(returnVal, 'f', precision, 64) + unit

} //Thanks for https://www.socketloop.com/tutorials/golang-byte-format-example

func (subdir *Dir) CalcTotalSize() float64 {
	print("Start to", subdir.node, "(size=", subdir.size, ")\n")
	TotalSize := float64(0)
	for _, tmpfile := range subdir.files {
		TotalSize += float64(tmpfile.Size())
	}
	fmt.Println(subdir.node, "(files)=", TotalSize)
	for _, tmpsubdir := range subdir.subdirectories {
		TotalSize += tmpsubdir.CalcTotalSize()
	}
	(*subdir).size = TotalSize
	fmt.Println("Continue to ", subdir.node, "(size=", subdir.size, ")\n")

	return TotalSize
}
func (subdir *Dir) SubDir2String(optSubdirParams ...SubdirParams) (error, string) {
	var fileArrow = "├───"
	var dirArrow = "──Dir─>"
	var lastFfileArrow = "└───"
	var TreeArrow = "├───"
	var EmptyArrow = "          "
	var VerticalArrow = "│"
	CurrentSubdirParams := SubdirParams{
		1,
		"",
		""}
	if len(optSubdirParams) > 0 {
		CurrentSubdirParams = optSubdirParams[0]
	}
	CurrentSubdirParams.LevelIndex += 1
	CurrentSubdirParams.DirPrefix += EmptyArrow //+"_"+strconv.Itoa(CurrentSubdirParams.LevelIndex)
	tmpSubdirDirprefix := CurrentSubdirParams.DirPrefix
	result := ""
	if CurrentSubdirParams.LevelIndex-1 > Options.MaxDepth && Options.MaxDepth != 0 {
		return nil, ""
	}

	for dirNum, CurrentSubdir := range subdir.subdirectories {
		if dirNum+1 != len(subdir.subdirectories) {
			result += CurrentSubdirParams.DirPrefix + TreeArrow + dirArrow + CurrentSubdir.name + " (" + ByteFormat(subdir.subdirectories[dirNum].size, 1) + ")\n"
		} else {
			result += CurrentSubdirParams.DirPrefix + TreeArrow + dirArrow + CurrentSubdir.name + " (" + ByteFormat(subdir.subdirectories[dirNum].size, 1) + ")\n"

		}

		CurrentSubdirParams.DirPrefix = tmpSubdirDirprefix + VerticalArrow

		if dirNum == len(subdir.subdirectories)-1 && len(CurrentSubdir.files) == 0 {
			CurrentSubdirParams.DirPrefix = tmpSubdirDirprefix
		}

		if err, tmp := CurrentSubdir.SubDir2String(CurrentSubdirParams); err == nil {
			result += tmp
		} else {
			result += CurrentSubdirParams.DirPrefix + "some error:" + fmt.Sprint(err) + "\n"
		}
		CurrentSubdirParams.DirPrefix = tmpSubdirDirprefix
		if Options.OnlyDirs != "yes" {

			for i, file := range CurrentSubdir.files {
				if dirNum == len(subdir.subdirectories)-1 {
					if i+1 != len(CurrentSubdir.files) {

						result += CurrentSubdirParams.DirPrefix + VerticalArrow + EmptyArrow + fileArrow + file.Name() + "(" + strconv.Itoa(CurrentSubdirParams.LevelIndex) + ")\n"
					} else {
						result += CurrentSubdirParams.DirPrefix + VerticalArrow + EmptyArrow + lastFfileArrow + file.Name() + "(" + strconv.Itoa(CurrentSubdirParams.LevelIndex) + ")\n"
					}

				} else {
					if i+1 != len(CurrentSubdir.files) {

						result += CurrentSubdirParams.DirPrefix + VerticalArrow + EmptyArrow + fileArrow + file.Name() + "(" + ByteFormat(float64(file.Size()), 1) + ")\n"
					} else {
						result += CurrentSubdirParams.DirPrefix + VerticalArrow + EmptyArrow + lastFfileArrow + file.Name() + "(" + ByteFormat(float64(file.Size()), 1) + ")\n"
					}
				}

			}
		}

	}
	return nil, result
}
func GetSubDir(currentDir Dir) (error, Dir) {
	files, err := ioutil.ReadDir(currentDir.node)
	if err != nil {
		return err, currentDir
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	for _, f := range files {
		if f.IsDir() {
			err, tmpdir := GetSubDir(Dir{node: filepath.Join(currentDir.node, f.Name()), name: f.Name()})
			if err == nil {
				currentDir.subdirectories = append(currentDir.subdirectories, tmpdir)
				currentDir.size += tmpdir.size
			} else {
				continue
			}
		} else {
			currentDir.files = append(currentDir.files, f)
			currentDir.size += float64(f.Size())
		}
	}
	//s ,err := DirSize(currentDir.node)
	//if err == nil {
	//	currentDir.size = float64(s)
	//}

	return nil, currentDir
}

var Options struct {
	MaxDepth int    `long:"max-depth" description:"maximum depth in output" default:"0" optional:"yes"`
	OnlyDirs string `long:"only-dirs" description:"print only directory sizes (no files) [yes|no]" optional:"yes" default:"no"`
	Colorize string `long:"colorize" description:"Use colorized output [yes|no]" optional:"yes" default:"yes"`
	Args     struct {
		Path string
	} `positional-args:"yes"`
}

func main() {
	var directory = Dir{node: "./"}

	_, err := flags.Parse(&Options)
	if err != nil {
		panic(err)
	}
	if Options.Args.Path != "" {
		directory = Dir{node: Options.Args.Path}
	}
	_, directory = GetSubDir(directory)
	_, stringResult := directory.SubDir2String()
	fmt.Println(directory.node, "(", ByteFormat(directory.size, 1), ")")
	fmt.Println(stringResult)
}
