package util

import (
	"os"

	"github.com/spf13/viper"
)

func init() {
	work, _ := os.Getwd()      // 获取目录路径
	viper.SetConfigName("app") // 设置文件名
	viper.SetConfigType("yml") // 设置文件类型
	//viper.AddConfigPath(work + "/config") // 执行单文件路径
	viper.AddConfigPath(work + "/config") // 执行go run文件路径
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func ReadStringConcif(parm string) string {
	value := viper.GetString(parm)

	return value
}

func ReadSliceConcif(parm string) []string {
	value := viper.GetStringSlice(parm)

	return value
}
