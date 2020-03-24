package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/andlabs/ui"
	"github.com/gozjw/desapp/utils"
)

func recoverPanic() {
	if err := recover(); err != nil {
		tips(fmt.Sprintf("秘钥错误：%v", err))
	}
}

func tips(s string) {
	ui.MsgBox(mainwin, "提示", s)
}

func confirm(s string) {
	confirmWin := ui.NewWindow("确认", 200, 150, true)
	yesButton := ui.NewButton("Yes")
	yesButton.OnClicked(func(b *ui.Button) {
		confirmWin.Destroy()
	})
	noButton := ui.NewButton("No")
	noButton.OnClicked(func(b *ui.Button) {
		confirmWin.Destroy()
	})

	confirmWin.SetMargined(true)
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	grid := ui.NewGrid()
	grid.SetPadded(true)
	vbox.Append(grid, false)
	grid.Append(ui.NewLabel(s), 0, 0, 2, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	grid.Append(yesButton, 1, 1, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(noButton, 1, 1, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	confirmWin.SetChild(vbox)
	confirmWin.OnClosing(func(*ui.Window) bool {
		return true
	})
	confirmWin.Show()
}

var (
	mainwin = ui.NewWindow("工具 v1.0.0 kiwi", 640, 600, true)

	grid = ui.NewGrid()

	openFilebutton = ui.NewButton("打开文件")
	fileEntry      = ui.NewEntry()
	passwordEntry  = ui.NewPasswordEntry()
	optionCombobox = ui.NewEditableCombobox()

	watchButton   = ui.NewButton("查看密码")
	decryptButton = ui.NewButton("解密")
	addButton     = ui.NewButton("添加")
	updateButton  = ui.NewButton("修改")
	deleteButton  = ui.NewButton("删除")

	readButton           = ui.NewButton("重新解析文件")
	emptyPasswordButton  = ui.NewButton("清空密码框")
	emptyPlaintextButton = ui.NewButton("清空明文框")

	fileName = "mySecret.json"
	file     = ""

	secretData = utils.Secret{}
	mainDir    = ""
)

func init() {
	secretData.Option = make(map[string]utils.Option)
	mainDir = utils.GetWorkDirectory()
	file = strings.ReplaceAll(mainDir, "/desapp.app/Contents/MacOS", "") + "/" + fileName
}

func updateCombo() {
	keys, err := secretData.GetOptionKeys()
	if err != nil {
		tips(fmt.Sprintf("获取解密项：%v", err))
		return
	}
	optionCombobox.Hide()
	optionCombobox = ui.NewEditableCombobox()
	for _, v := range keys {
		optionCombobox.Append(v)
	}
	grid.Append(optionCombobox, 1, 2, 5, 1, true, ui.AlignFill, false, ui.AlignFill)
}

func readFile() error {
	if !utils.FileExist(file) {
		return errors.New("文件不存在")
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("读取文件错误：%v", err)
	}
	if len(b) == 0 {
		b = []byte("{}")
	}
	err = json.Unmarshal(b, &secretData)
	if err != nil {
		return fmt.Errorf("解析json错误：%v", err)
	}
	updateCombo()
	return nil
}

func writeFile() error {
	b, err := json.Marshal(secretData)
	if err != nil {
		return fmt.Errorf("构造json错误：%v", err)
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("打开文件错误：%v", err)
	}
	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("写入文件错误：%v", err)
	}
	f.Close()
	updateCombo()
	return nil
}

func mySecret() ui.Control {
	//垂直框
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	grid.SetPadded(true)
	vbox.Append(grid, false)

	// 操作
	fileEntry.SetText(file)
	openFilebutton.OnClicked(func(*ui.Button) {
		fileEntry.SetText(ui.OpenFile(mainwin))
	})
	watchButton.OnClicked(func(*ui.Button) {
		password := passwordEntry.Text()
		if password != "" {
			// tips(password)
			confirm("")
		}
	})

	// 布局
	grid.Append(ui.NewLabel("文件"), 0, 0, 1, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	grid.Append(fileEntry, 1, 0, 4, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(openFilebutton, 5, 0, 1, 1, true, ui.AlignFill, false, ui.AlignFill)

	grid.Append(ui.NewLabel("密码"), 0, 1, 1, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	grid.Append(passwordEntry, 1, 1, 5, 1, true, ui.AlignFill, false, ui.AlignFill)

	grid.Append(ui.NewLabel("加密项"), 0, 2, 1, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	// grid.Append(optionCombobox, 1, 2, 5, 1, true, ui.AlignFill, false, ui.AlignFill)

	grid.Append(ui.NewLabel("操作"), 0, 3, 1, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	grid.Append(watchButton, 1, 3, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(decryptButton, 2, 3, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(addButton, 3, 3, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(updateButton, 4, 3, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(deleteButton, 5, 3, 1, 1, true, ui.AlignFill, false, ui.AlignFill)

	grid.Append(ui.NewLabel("输入"), 0, 4, 1, 1, false, ui.AlignCenter, false, ui.AlignCenter)
	grid.Append(readButton, 1, 4, 1, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(emptyPasswordButton, 2, 4, 2, 1, true, ui.AlignFill, false, ui.AlignFill)
	grid.Append(emptyPlaintextButton, 4, 4, 2, 1, true, ui.AlignFill, false, ui.AlignFill)

	group := ui.NewGroup("明文")
	group.SetMargined(true)
	vbox.Append(group, true)

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)
	group.SetChild(entryForm)

	entryForm.Append("", ui.NewMultilineEntry(), true)

	err := readFile()
	if err != nil {
		tips(fmt.Sprintf("%v", err))
	}
	return vbox
}

//设置UI界面
func setupUI() {
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	tab := ui.NewTab()
	mainwin.SetChild(tab)
	mainwin.SetMargined(true)

	tab.Append("加密器", mySecret())
	tab.SetMargined(0, true)

	mainwin.Show()
}

func main() {
	ui.Main(setupUI)

}
