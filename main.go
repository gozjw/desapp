package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/gozjw/desapp/utils"
)

func recoverPanic() {
	if err := recover(); err != nil {
		multiline = ""
		tips(fmt.Sprintf("秘钥错误：%v", err))
	}
}

var (
	isDebug    = false
	fileName   = "mySecret.json"
	file       string
	password   string
	addOption  string
	decOption  string
	optionList []string
	selected   int32
	multiline  string

	// windowsTTCList = []string{"msyh.ttc", "msjhbd.ttc", "msyhbd.ttc"}
	windowTTCPath = "c:/Windows/Fonts/msyh.ttc"
	macTTCPath    = "/System/Library/Fonts/STHeiti Light.ttc"

	font     imgui.Font
	fontFile string

	secretData = utils.Secret{}

	mainDir = ""

	operaFlagDelete = 0
	operaFlagModify = 1
)

func updateCombo() {
	keys, err := secretData.GetOptionKeys()
	if err != nil {
		tips(fmt.Sprintf("获取解密项：%v", err))
		return
	}
	optionList = keys
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

func init() {
	secretData.Option = make(map[string]utils.Option)
	switch runtime.GOOS {
	case "darwin":
		fontFile = macTTCPath
	default:
		fontFile = windowTTCPath
	}
	mainDir = utils.GetFileDirectory()
	if isDebug {
		mainDir = utils.GetWorkDirectory()
	}

	file = strings.ReplaceAll(mainDir, "/desapp.app/Contents/MacOS", "") + "/" + fileName
}

func initFont() {
	fonts := giu.Context.IO().Fonts()
	font = fonts.AddFontFromFileTTFV(
		fontFile,                       // <- Font file path
		20,                             // <- Font size
		imgui.DefaultFontConfig,        // <- Usually you don't need to modify this
		fonts.GlyphRangesChineseFull()) // <- Add Chinese glyph ranges to display chinese characters
}

func tips(msg string) {
	giu.Msgbox("提示", msg, giu.MsgboxButtonsOk, func(i giu.DialogResult) {})
}
func confirm(msg string, opera int) {
	giu.Msgbox("确认", msg, giu.MsgboxButtonsYesNo, func(i giu.DialogResult) {
		if i == giu.DialogResultNo {
			return
		}
		switch opera {
		case operaFlagDelete:
			deleteOpera()
		case operaFlagModify:
			operaModify()
		}
	})
}
func operaModify() {
	defer recoverPanic()
	if err := secretData.Update(password, decOption, multiline); err != nil {
		tips(fmt.Sprintf("修改失败：%v", err))
		return
	}
	if err := writeFile(); err == nil {
		tips("修改成功")
	}
}

func deleteOpera() {
	if err := secretData.Delete(password, decOption); err != nil {
		tips(fmt.Sprintf("删除失败：%v", err))
		return
	}
	if err := writeFile(); err == nil {
		multiline = ""
		tips("删除成功")
	}
}

func loop() {
	if len(optionList) == 0 || selected >= int32(len(optionList)) {
		decOption = ""
	} else {
		decOption = optionList[selected]
	}

	giu.PushFont(font)
	giu.SingleWindow("加密器", giu.Layout{
		giu.Line(
			giu.InputTextV("文件", 0, &file, giu.InputTextFlagsNone, nil, nil),
			giu.Button("解析", func() {
				file = strings.TrimSpace(file)
				if file == "" {
					tips("文件为空")
					return
				}
				readFile()
			}),
		),
		giu.Line(
			giu.InputTextV("密码(<=16)", 0, &password, giu.InputTextFlagsPassword, nil, nil),
			giu.Button("查看", func() {
				if password != "" {
					tips(password)
				}
			}),
			giu.Button("清空密码", func() {
				password = ""
			}),
		),
		giu.Line(
			giu.InputText("加密项", 0, &addOption),
			giu.Button("添加", func() {
				defer recoverPanic()
				if err := secretData.Add(password, addOption, multiline); err != nil {
					tips(fmt.Sprintf("添加失败：%v", err))
					return
				}
				if err := writeFile(); err == nil {
					tips("添加成功")
				}
			}),
			giu.Button("清空加密项", func() {
				addOption = ""
			}),
		),
		giu.Line(
			giu.Combo("解密项", decOption, optionList, &selected, 0, 0, func() {
				multiline = ""
			}),
			giu.Button("解密", func() {
				defer recoverPanic()
				pt, err := secretData.GetOptionDataByKey(password, decOption)
				multiline = pt
				if err != nil {
					tips(fmt.Sprintf("解密失败：%v", err))
					return
				}
			}),
			giu.Button("修改", func() {
				confirm(fmt.Sprintf("确认修改“%s”？", decOption), operaFlagModify)
			}),
			giu.Button("删除", func() {
				confirm(fmt.Sprintf("确认删除“%s”？", decOption), operaFlagDelete)
			}),
		),
		giu.Line(
			giu.LabelWrapped("明文："),
			giu.Button("清空文本", func() {
				multiline = ""
			}),
		),
		giu.Line(
			giu.InputTextMultiline("", &multiline, -1, -1, giu.InputTextFlagsNone, nil, nil),
		),
		giu.PrepareMsgbox(),
	})
	giu.PopFont()
}

func main() {
	w := giu.NewMasterWindow("工具 v1.0.0 kiwi", 850, 600, 0, initFont)
	readFile()
	w.Main(loop)
}
