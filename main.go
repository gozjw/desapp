package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
)

var (
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

	dataMap = make(map[string]Data, 0)

	operaFlagDelete = 0
	operaFlagModify = 1
)

func recoverPanic() {
	if err := recover(); err != nil {
		multiline = ""
		tips(fmt.Sprintf("秘钥错误：%v", err))
	}
}

func updateCombo() {
	optionList = []string{}
	for k := range dataMap {
		optionList = append(optionList, k)
	}
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

type Data struct {
	Data string `json:"data"`
}

func readFile() error {
	if !FileExist(file) {
		return errors.New("文件不存在")
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("读取文件错误：%v", err)
	}
	if len(b) == 0 {
		b = []byte("{}")
	}
	err = json.Unmarshal(b, &dataMap)
	if err != nil {
		return fmt.Errorf("解析json错误：%v", err)
	}
	updateCombo()
	return nil
}

func writeFile() error {
	b, err := json.Marshal(dataMap)
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

// GetMainDirectory 获取项目根目录
func GetMainDirectory() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	//兼容linux "\\"转"/"
	return strings.Replace(dir, "\\", "/", -1)
}

func init() {
	switch runtime.GOOS {
	case "darwin":
		fontFile = macTTCPath
	default:
		fontFile = windowTTCPath
	}

	file = strings.ReplaceAll(GetMainDirectory(), "/desapp.app/Contents/MacOS", "") + "/" + fileName
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
	if password == "" {
		tips("密码为空")
		return
	}
	if !checkPassword(password) {
		tips("密码过长")
		return
	}
	if decOption == "" {
		tips("解密项为空")
		return
	}
	if _, ok := dataMap[decOption]; !ok {
		tips("解密项已存在")
		return
	}
	if multiline == "" {
		tips("明文为空")
		return
	}
	b, err := encryptAES([]byte(multiline), []byte(password))
	if err != nil {
		tips(fmt.Sprintf("加密失败：%v", err))
		return
	}
	dataMap[decOption] = Data{Data: base64.StdEncoding.EncodeToString(b)}
	if err := writeFile(); err == nil {
		tips("修改成功")
	}
}

func deleteOpera() {
	delete(dataMap, decOption)
	multiline = ""
	if err := writeFile(); err == nil {
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
			giu.InputTextV("密码(16位)", 0, &password, giu.InputTextFlagsPassword, nil, nil),
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
				if password == "" {
					tips("密码为空")
					return
				}
				if !checkPassword(password) {
					tips("密码过长")
					return
				}
				if addOption == "" {
					tips("加密项为空")
					return
				}
				if _, ok := dataMap[addOption]; ok {
					tips("加密项已存在")
					return
				}
				if multiline == "" {
					tips("明文为空")
					return
				}
				b, err := encryptAES([]byte(multiline), []byte(password))
				if err != nil {
					tips(fmt.Sprintf("加密失败：%v", err))
					return
				}
				dataMap[addOption] = Data{Data: base64.StdEncoding.EncodeToString(b)}
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
				if decOption == "" {
					tips("请选择需要解密的项")
					return
				}
				if password == "" {
					multiline = ""
					tips("密码为空")
					return
				}
				data, ok := dataMap[decOption]
				if !ok {
					multiline = ""
					tips("解密项不存在")
					return
				}
				be, err := base64.StdEncoding.DecodeString(data.Data)
				if err != nil {
					multiline = ""
					tips(fmt.Sprintf("base64解密失败：%v", err))
					return
				}
				b, err := decryptAES(be, []byte(password))
				if err != nil {
					multiline = ""
					tips(fmt.Sprintf("解密失败：%v", err))
					return
				}
				multiline = string(b)
			}),
			giu.Button("修改", func() {
				defer recoverPanic()
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

// 填充数据
func padding(src []byte, blockSize int) []byte {
	padNum := blockSize - len(src)%blockSize
	pad := bytes.Repeat([]byte{byte(padNum)}, padNum)
	return append(src, pad...)
}

// 去掉填充数据
func unpadding(src []byte) []byte {
	n := len(src)
	unPadNum := int(src[n-1])
	return src[:n-unPadNum]
}

// 加密
func encryptAES(src []byte, key []byte) ([]byte, error) {
	key = padding(key, 16)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	src = padding(src, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	blockMode.CryptBlocks(src, src)
	return src, nil
}

// 解密
func decryptAES(src []byte, key []byte) ([]byte, error) {
	key = padding(key, 16)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	blockMode.CryptBlocks(src, src)
	src = unpadding(src)
	return src, nil
}

func checkPassword(password string) bool {
	if len(password) > 16 {
		return false
	}
	return true
}
