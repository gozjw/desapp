package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	font     imgui.Font
	fontFile string

	dataMap = make(map[string]Data, 0)
)

func updateCombo() {
	optionList = []string{}
	for k := range dataMap {
		optionList = append(optionList, k)
	}
}

func GetMainDirectory() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

type Data struct {
	Data string `json:"data"`
}

func readFile() {
	if !FileExist(file) {
		tips("文件不存在")
		return
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		tips(fmt.Sprintf("读取文件错误：%v", err))
		return
	}
	if len(b) == 0 {
		b = []byte("{}")
	}
	err = json.Unmarshal(b, &dataMap)
	if err != nil {
		tips(fmt.Sprintf("解析json错误：%v", err))
		return
	}
	updateCombo()
}

func writeFile() {
	if !FileExist(file) {
		tips("文件不存在")
		return
	}
	b, err := json.Marshal(dataMap)
	if err != nil {
		tips(fmt.Sprintf("构造json错误：%v", err))
		return
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		tips(fmt.Sprintf("打开文件错误：%v", err))
		return
	}
	_, err = f.Write(b)
	if err != nil {
		tips(fmt.Sprintf("写入文件错误：%v", err))
		return
	}
	f.Close()
	updateCombo()
	tips("添加成功")
}

func init() {
	fontFile = windowTTCPath
	file = GetMainDirectory() + "/" + fileName
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
	giu.Msgbox("提示", msg, 0, func(i giu.DialogResult) {})
}

func loop() {
	if len(optionList) == 0 {
		decOption = ""
	} else {
		decOption = optionList[selected]
	}

	giu.PushFont(font)
	giu.SingleWindow("Window", giu.Layout{
		giu.Line(
			giu.InputText("文件位置", 0, &file),
		),
		giu.Line(
			giu.InputTextV("密码(小于等于16位)", 0, &password, giu.InputTextFlagsPassword, nil, nil),
			giu.Button("查看密码", func() {
				if password != "" {
					tips(password)
				}
			}),
			giu.Label(""),
		),
		giu.Line(
			giu.InputText("加密项", 0, &addOption),
			giu.Button("添加", func() {
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
				writeFile()
			}),
		),
		giu.Line(
			giu.Combo("选择解密项", decOption, optionList, &selected, 0, 0, func() {}),
			giu.Button("解密", func() {
				if decOption == "" {
					tips("请选择需要解密的项")
					return
				}
				if password == "" {
					tips("密码为空")
					return
				}
				data, ok := dataMap[decOption]
				fmt.Println(data, ok, decOption)
				if !ok {
					tips("解密项不存在")
					return
				}
				be, err := base64.StdEncoding.DecodeString(data.Data)
				if err != nil {
					tips(fmt.Sprintf("base64解密失败：%v", err))
					return
				}
				b, err := decryptAES(be, []byte(password))
				if err != nil {
					tips(fmt.Sprintf("解密失败：%v", err))
					return
				}
				multiline = string(b)
			}),
			giu.Button("修改", func() {}),
			giu.Button("删除", func() {}),
		),
		giu.Line(
			giu.Label("明文："),
		),
		giu.Line(
			giu.InputTextMultiline("", &multiline, -1, -1, 0, nil, nil),
		),
		giu.PrepareMsgbox(),
	})
	giu.PopFont()
}

func main() {
	w := giu.NewMasterWindow("加密器", 850, 600, 0, initFont)
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
