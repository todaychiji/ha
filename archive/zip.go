package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/hahalab/qi/config"
	"encoding/json"
)

// Build 把目录打包成 cwd()/code.zip
func Build(dir string, hintMessage chan string) error {
	c, err := config.LoadConfig(path.Join(dir, "qi.yml"))
	if err != nil || c == nil {
		return err
	}

	hintMessage <- "Compiling"
	if err := executeBuild(dir, *c); err != nil {
		return err
	}

	// create qi.json for fc proxy
	qiJsonPath := path.Join(dir, "qi.json")
	qiJson, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(qiJsonPath, qiJson, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(qiJsonPath); err != nil {
			panic(err)
		}
	}()

	output, err := os.Create("code.zip")
	if err != nil {
		return err
	}
	defer output.Close()

	tw := zip.NewWriter(output)
	// Write index.py to zip

	hintMessage <- "Injecting"
	err = injectProxy(tw)
	if err != nil {
		return err
	}
	hintMessage <- "Building"
	// Write files to tw
	files := c.Files
	files = append(files, qiJsonPath)

	for _, f := range files {
		targetPath := path.Join(dir, f)
		err = injectDir(targetPath, dir, tw)
		if err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return err
}

func injectProxy(tw *zip.Writer) error {
	input, err := codeZipBytes()
	if err != nil {
		return err
	}
	r, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return err
	}
	for _, file := range r.File {
		//fmt.Printf("Add file %s\n", file.Name)
		info := file.FileInfo()
		header, err := zip.FileInfoHeader(info)
		header.Name = file.Name

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		if info.Mode()&os.ModeSymlink != 0 {
			header.SetMode(0)
		}

		//fmt.Printf("create heade %+v\n", header)
		writer, err := tw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		r, err := file.Open()
		if err != nil {
			return err
		}
		defer r.Close()

		_, err = io.Copy(writer, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func injectDir(dir string, baseDir string, tw *zip.Writer) error {
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		zipPath := strings.TrimPrefix(filePath, baseDir)
		if info == nil {
			return nil
		}
		header, err := zip.FileInfoHeader(info)

		header.Name = zipPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		if info.Mode()&os.ModeSymlink != 0 {
			header.SetMode(0)
		}

		//fmt.Printf("create heade %+v\n", header)
		writer, err := tw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func executeBuild(dir string, c config.CodeConfig) error {

	if c.Build == "" {
		return nil
	}

	oldPwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err = os.Chdir(dir); err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", c.Build)

	out := []byte{}
	cmd.Stdout = bytes.NewBuffer(out)
	err = cmd.Run()
	if err != nil {
		return err
	}

	if err = os.Chdir(oldPwd); err != nil {
		return err
	}
	return nil
}
