package main

import (
	"fmt"

	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/xi123/libgo/utils/crypto/aes"
	db "github.com/xi123/libgo/utils/dbwraper"

	"github.com/cwloo/server/src/common/mongoop"
	"github.com/cwloo/server/src/config"
)

func init() {
	cmd.InitArgs(func(arg *cmd.ARG) {
		arg.SetConf("config/conf.ini")
	})
}

const (
	AES_KEY = "dstar!@#$01234561234567890@dubai"
	AES_IV  = "dstar67890@dubai"
)
const (
	AES_KEY_FORCLIENT = "yaoxing8901234561234567890123488"
	AES_IV_FORCLIENT  = "yaoxing890123488"
)

func Test() {

	aes.CBCTest()
	aes.ECBTest()
	src := "hello,world"

	enc := aes.CBCEncryptPKCS7([]byte(src), []byte(AES_KEY_FORCLIENT), []byte(AES_IV_FORCLIENT))

	dst := aes.CBCDecryptPKCS7(enc, []byte(AES_KEY_FORCLIENT), []byte(AES_IV_FORCLIENT))

	fmt.Println(dst)

	// key := "0123456789abcdef"
	// key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	// crypted := codec.AesEncryptPKCS5([]byte(src), []byte(key))
	// fmt.Printf("加密---%v\n", string(crypted))
	// fmt.Printf("原始---%v\n", string(src))
	// decrypted := codec.AesDecryptPKCS5(crypted, []byte(key))
	// fmt.Printf("%v\n", string(decrypted))
	// fmt.Printf("解密---%v\n", string(decrypted))
	// fmt.Print("\n\n\n\n\n")

	// decrypted = codec.AesDecryptECB(crypted, []byte(key))
	// fmt.Printf("%v\n", string(decrypted))
	// fmt.Printf("解密---%v\n", string(decrypted))
	fmt.Print("")
}

func Test2() {

	// src := "hello,world"
	// key := "0123456789abcdef"

	// crypted := codec.AesCBCEncryptPKCS7([]byte(src), []byte(key))
	// fmt.Printf("加密---%v\n", string(crypted))
	// fmt.Printf("原始---%v\n", string(src))
	// decrypted := codec.AesCBCDecryptPKCS7(crypted, []byte(key))
	// fmt.Printf("%v\n", string(decrypted))
	// fmt.Printf("解密---%v\n", string(decrypted))
	fmt.Print("")
}

func main() {
	Test()
	Test2()
	return
	// codec.Test2()
	cmd.ParseArgs()
	config.InitConfig(cmd.Conf())
	db.Init(
		config.Config.Redis,
		config.Config.Mongo,
		config.Config.Mysql,
		config.Config.Mysql)
	mongoop.CreateIndex()
	mongoop.InitAutoIncrement()
	logs.Close()
}
