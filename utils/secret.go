package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type Secret struct {
	Password string            `json:"password"`
	Option   map[string]Option `json:"data"`
}

type Option struct {
	Deleted bool     `json:"deleted"`
	Record  []Record `json:"record"`
}

type Record struct {
	Time time.Time `json:"time"`
	Data string    `json:"data"`
}

func (t *Secret) GetOptionKeys() (result []string, err error) {
	for k, v := range t.Option {
		if !v.Deleted {
			result = append(result, k)
		}
	}
	// 排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if !ChineseLess(result[i], result[j]) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return
}

func (t *Secret) GetOptionDataByKey(password, key string) (result string, err error) {
	if !ComparePassword(t.Password, password) {
		return result, errors.New("密码错误")
	}
	option, ok := t.Option[key]
	if !ok || option.Deleted {
		return result, errors.New("解密项不存在")
	}

	var record = Record{}
	for _, v := range option.Record {
		if record.Time.Unix() < v.Time.Unix() {
			record = v
		}
	}

	be, err := base64.StdEncoding.DecodeString(record.Data)
	if err != nil {
		return result, errors.New(fmt.Sprintf("base64解密失败：%v", err))
	}
	b, err := DecryptAES(be, []byte(password))
	if err != nil {
		return result, errors.New(fmt.Sprintf("解密失败：%v", err))
	}
	result = string(b)
	return
}

func (t *Secret) Add(password, key, data string) error {

	if t.Password != "" {
		if !ComparePassword(t.Password, password) {
			return errors.New("密码错误")
		}
	} else {
		if !CheckPassword(password) {
			return errors.New("密码长度不正确")
		}
		p, err := EncryptionPassword(password)
		if err != nil {
			return errors.New("加密密码错误")
		}
		t.Password = p
	}
	if key == "" {
		return errors.New("加密项为空")
	}
	if data == "" {
		return errors.New("加密内容为空")
	}
	option, ok := t.Option[key]
	if ok && !option.Deleted {
		return errors.New("加密项已存在")
	}
	b, err := EncryptAES([]byte(data), []byte(password))
	if err != nil {
		return errors.New(fmt.Sprintf("加密失败：%v", err))
	}
	if ok {
		option.Deleted = false
		option.Record = append(option.Record, Record{
			Time: time.Now(),
			Data: base64.StdEncoding.EncodeToString(b),
		})
		t.Option[key] = option
	} else {
		t.Option[key] = Option{
			Deleted: false,
			Record: []Record{
				Record{
					Time: time.Now(),
					Data: base64.StdEncoding.EncodeToString(b),
				},
			},
		}
	}

	return nil
}

func (t *Secret) Update(password, key, data string) error {
	if !ComparePassword(t.Password, password) {
		return errors.New("密码错误")
	}
	if key == "" {
		return errors.New("加密项为空")
	}
	if data == "" {
		return errors.New("加密内容为空")
	}
	option, ok := t.Option[key]
	if !ok {
		return errors.New("加密项不存在")
	}
	b, err := EncryptAES([]byte(data), []byte(password))
	if err != nil {
		return errors.New(fmt.Sprintf("加密失败：%v", err))
	}
	option.Deleted = false
	option.Record = append(option.Record, Record{
		Time: time.Now(),
		Data: base64.StdEncoding.EncodeToString(b),
	})

	t.Option[key] = option

	return nil
}

func (t *Secret) Delete(password, key string) error {
	if !ComparePassword(t.Password, password) {
		return errors.New("密码错误")
	}
	option, ok := t.Option[key]
	if !ok {
		return errors.New("删除项不存在")
	}
	option.Deleted = true
	t.Option[key] = option
	return nil
}

func (t *Secret) ModifyPassword(password, newPassword string) error {
	if !ComparePassword(t.Password, password) {
		return errors.New("密码错误")
	}
	if !CheckPassword(newPassword) {
		return errors.New("密码长度不正确")
	}

	p, err := EncryptionPassword(newPassword)
	if err != nil {
		return errors.New("新密码错误")
	}
	t.Password = p

	for k, v := range t.Option {
		for i := range v.Record {
			be, err := base64.StdEncoding.DecodeString(v.Record[i].Data)
			if err != nil {
				return errors.New(fmt.Sprintf("base64解密失败：%v", err))
			}
			b, err := DecryptAES(be, []byte(password))
			if err != nil {
				return errors.New(fmt.Sprintf("解密失败：%v", err))
			}
			b, err = EncryptAES(b, []byte(newPassword))
			if err != nil {
				return errors.New(fmt.Sprintf("加密失败：%v", err))
			}
			v.Record[i].Data = base64.StdEncoding.EncodeToString(b)
		}
		t.Option[k] = v
	}

	return nil
}
