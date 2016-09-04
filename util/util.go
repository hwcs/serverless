package util

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"encoding/base64"
	"serverless/model"
)

var OPEN_LAMBDA_HOME string

var REDIS_CONF = make(map[string]string)

var LOAD_BALANCER_ADDRESS string

var REGISTRY_ADDRESS string

func Init() error {
	OPEN_LAMBDA_HOME = os.Getenv("OPEN_LAMBDA_HOME")
	fmt.Println("OPEN_LAMBDA_HOME:", OPEN_LAMBDA_HOME)

	e := GetRedisConfig()
	if e != nil {
		return e
	}

	e = InitFuncTempDir()
	if e != nil {
		return e
	}

	e = GetLoadBalancerAddress()
	if e != nil {
		return e
	}

	e = GetRegistryAddress()
	if e != nil {
		return e
	}

	return nil
}

func GetRedisConfig() error {
	file := OPEN_LAMBDA_HOME + "/util/cluster/redis.json"
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("GetRedisConfig, ReadFile:", err)
		return err
	}

	datajson := []byte(data)
	err = json.Unmarshal(datajson, &REDIS_CONF)
	if err != nil {
		fmt.Println("GetRedisConfig, Unmarshal:", err)
		return err
	}

	fmt.Printf("GetRedisConfig: %s:%s\n", REDIS_CONF["network"], REDIS_CONF["port"])
	return nil
}

func GetRegistryAddress() error {
	file := OPEN_LAMBDA_HOME + "/util/cluster/registry.json"

	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("GetRegistryAddress, ReadFile:", err)
		return err
	}

	datajson := []byte(data)
	var v map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(datajson, &v)
	if err != nil {
		fmt.Println("GetRegistryAddress, Unmarshal:", err)
		return err
	}
	REGISTRY_ADDRESS = v["ip"].(string) + ":5000" // + v["host_port"].(string)

	fmt.Printf("GetRegistryAddress: %s\n", REGISTRY_ADDRESS)
	return nil
}

func GetLoadBalancerAddress() error {
	balancerFile := OPEN_LAMBDA_HOME + "/util/cluster/loadbalancer-1.json"

	data, err := ioutil.ReadFile(balancerFile)
	if err != nil {
		fmt.Println("GetLoadBalancerAddress, ReadFile:", err)
		return err
	}

	datajson := []byte(data)
	var v map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(datajson, &v)
	if err != nil {
		fmt.Println("GetLoadBalancerAddress, Unmarshal:", err)
		return err
	}
        fmt.Println (v)
	LOAD_BALANCER_ADDRESS = "http://" + v["ip"].(string) + ":" + v["port"].(string)

	fmt.Printf("GetLoadBalancerAddress: %s\n", LOAD_BALANCER_ADDRESS)
	return nil
}

func Trace(msg string) func() {
	start := time.Now()
	log.Printf("enter %s ...", msg)
	return func() {
		log.Printf("exit %s (%s)", msg, time.Since(start))
	}
}

func InitFuncTempDir() error {
	dir := OPEN_LAMBDA_HOME + "/tmp"
	_, err := os.Stat(dir)

	if err == nil {
		return nil
	}

	err = os.Mkdir(dir, 0755)
	if err == nil {
		fmt.Println("create temp succeed, path:", dir)
	} else {
		fmt.Println("create temp failed, path:", dir)
	}
	return err
}

func GetRandomString(funcname string) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" + funcname
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 12; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	fmt.Println("GetRandomString", string(result))
	v := strings.ToLower(string(result))

	return v
}

func unzipFuncFile(path string, file string) error {
	defer Trace("unzipFunFile")()

	r, err := zip.OpenReader(file)
	if err != nil {
		fmt.Println("unzipFuncFile, open reader: ", err)
		return err
	}
	for _, k := range r.Reader.File {
		if k.FileInfo().IsDir() {
			subdir := path + "/" + k.Name
			err := os.MkdirAll(subdir, 0755)
			if err != nil {
				fmt.Println("unzipFuncFile, make dir error:", err)
			}
			continue
		}
		r, err := k.Open()
		if err != nil {
			fmt.Println("unzipFuncFile, open file:", err)
			continue
		}

		defer r.Close()

		//newfile, err := os.Create(path + "/lambda_func.py")
		newfile, err := os.Create(path + "/" + k.Name)
		if err != nil {
			fmt.Println("unzipFuncFile: ", err)
			continue
		}
		io.Copy(newfile, r)

		newfile.Close()
	}
	os.Remove(file)
	return nil
}

func WriteFuncfile(funcname string, codetype string, content string) (string, error) {
	tag := GetRandomString(funcname)
	fmt.Println("WriteFuncfile, image tag:", tag)
	path := OPEN_LAMBDA_HOME + "/tmp/" + tag

	err := os.Mkdir(path, 0755)
	if err != nil {
		fmt.Println("WriteFuncfile, create func temp dir error:", path)
		return tag, err
	}
	if codetype == model.CODE_TYPE_INLINE {
		fn := path + "/lambda_func.py"
		e := ioutil.WriteFile(fn, []byte(content), 0666)
		if e != nil {
			fmt.Println("WriteFuncfile, error write file")
			return tag, e
		}
		return tag, nil
	}

	appPath := path + "/"
	fn := appPath + funcname + ".zip"
	value, _ := base64.StdEncoding.DecodeString(content)
	e := ioutil.WriteFile(fn, value, 0666)
	if e != nil {
		fmt.Println("WriteFuncfile, error write file")
		return "", nil
	}
	unzipFuncFile(path, fn)
	return tag, nil
}

func CopyFile(srcName, dstName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		fmt.Println("CopyFile error:", err)
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("CopyFile error:", err)
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func DeleteFuncTmpDir(dir string) {
	e := os.RemoveAll(dir)
	if e != nil {
		fmt.Println("DeleteFuncTmpDir: %s, error:", e.Error())
	}
}
