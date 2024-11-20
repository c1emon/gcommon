package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var r *rand.Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(100)))
	r = rand.New(source)
}

var cnNums = [...]string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}

// Num2No 数字序号转中文
func Num2No(num int) string {
	if num == 0 {
		return cnNums[0]
	}

	numStr := strconv.Itoa(num)
	ret := ""
	for i := 0; i < len(numStr); i++ {
		id, err := strconv.Atoi(string(numStr[i]))
		if err != nil {
			panic(err)
		}
		ret = fmt.Sprintf("%s%s", ret, cnNums[id])
	}

	return ret
}

func PrettyMarshal(data any) string {
	b, _ := json.MarshalIndent(data, "", "    ")
	return string(b)
}

func RandStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
