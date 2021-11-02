package main

import (
	"bufio"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// GetCrt 获取根目录下所有crt文件
func GetCrt() []string {
	glob, err := filepath.Glob("./*.crt")
	if err != nil {
		return nil
	}
	return glob
}

func ExecCertmgr(success int,value string) int{
	cmd := exec.Command("cmd", "/c", "certmgr.exe -c -add "+value+" -s root")
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return success
	}
	go func() {
		//如果有多证书就安装第一个
		defer stdinPipe.Close()
		io.WriteString(stdinPipe, string(1))
	}()
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmd.CombinedOutput()
	if err != nil {
		println(err.Error())
		return success
	}
	success++
	return success
}

// Certmgr 写入证书到根
func Certmgr(glob []string) int {
	success := 0
	for _, value := range glob {
		success=ExecCertmgr(success,value)
	}
	return success
}

// SettingFont 设置中文字体
func SettingFont() {
	fontPath, err := findfont.Find("simhei.ttf")
	if err != nil {
		return
	}
	os.Setenv("FYNE_FONT", fontPath)
}

//GetAllFile 遍历所有子目录，找到文件
func GetAllFile(pathname string, file string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "\\" + fi.Name()
			s, err = GetAllFile(fullDir, file, s)
			if err != nil {
				return s, err
			}
		} else {
			if fi.Name() == file {
				fullName := pathname + "\\" + fi.Name()
				s = append(s, fullName)
			}
		}
	}
	return s, nil
}

//ReadLine 读取每行
func ReadLine(filePth string, hookfn func([]byte)) error {
	f, err := os.Open(filePth)
	if err != nil {
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	for {
		line,  _, err := bfRd.ReadLine()
		hookfn(line)    //放在错误处理前面，即使发生错误，也会处理已经读取到的数据。
		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

// SettingFirefox 设置火狐企业根偏好
func SettingFirefox() {
	current, err := user.Current()
	if err != nil {
		return
	}
	firefoxPath := current.HomeDir + "\\AppData\\Roaming\\Mozilla\\Firefox"
	if _, err := os.Stat(firefoxPath); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}
	var s []string
	s, err = GetAllFile(firefoxPath, "prefs.js", s)
	if err != nil {
		return
	}
	if len(s) == 0 {
		return
	}
	for _, file := range s {
		//读取每行删掉有关“security.enterprise_roots.enabled”的行
		var content string
		err = ReadLine(file, func(bytes []byte) {
			if !strings.Contains(string(bytes), "security.enterprise_roots.enabled")&&string(bytes)!="" {
				content+=string(bytes)+"\n"
			}
		})
		if err != nil {
			return
		}
		content+="user_pref(\"security.enterprise_roots.enabled\", true);"
		//闭包，读写方式打开
		func() {
			openFile, err := os.OpenFile(file, os.O_RDWR, 0666)
			if err != nil {
				return
			}
			defer openFile.Close()
			err = openFile.Truncate(0)
			if err != nil {
				return
			}
			openFile.Seek(0, 0)
			_, err = openFile.WriteString(content)
			if err != nil {
				return
			}
		}()
	}

}

func main() {
	SettingFont()
	a := app.New()
	w := a.NewWindow("安装证书")
	glob := GetCrt()
	label := widget.NewLabel("检测到目录内有" + strconv.Itoa(len(glob)) + "个证书")
	button := widget.NewButton("安装证书", func() {
		SettingFirefox()
		success := Certmgr(glob)
		dialog.ShowInformation("", "成功安装"+strconv.Itoa(success)+"个证书", w)
	})
	w.SetContent(container.NewCenter(container.NewVBox(
		label, button,
	)))
	w.Resize(fyne.NewSize(240, 180))
	w.CenterOnScreen()
	w.ShowAndRun()
	os.Unsetenv("FYNE_FONT")
}
