package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"serverless/model"
)

const URL = "http://10.67.202.161:8888"

func ReadFuncZipFile(file string) ([]byte, error) {
	var buf []byte
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("error read file!", file)
		return nil, err
	}
	return buf, nil
}

func CreateFunction(name string) {

	var f model.Function
	funcfile, err := ReadFuncZipFile("HelloWorld.zip")
	if err != nil {
		return
	}

	f.FuncCode.File = funcfile
	f.FuncCode.S3Bucket = "testbucket"
	f.FuncCode.S3Key = "tests3key"
	f.FuncCode.S3ObjectVersion = "1.0"
	f.FuncCode.CodeType = model.CODE_TYPE_ZIPFILE
	f.Description = "Hello world function description"
	f.FunctionName = name
	f.Handler = "HelloWorldHandler"
	f.MemorySize = 2048
	f.Publish = true
	f.Runtime = "Node.js"
	f.Timeout = 60
	data, err := json.Marshal(f)

	url := URL + "/functions/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("CreateFunction response Status:", resp.Status)
	fmt.Println("CreateFunction response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("CreateFunction response Body:", string(body))
}

func UpdateFunctionCode(name string) {
	var code struct {
		Publish bool
		ZipFile []byte
	}
	funcfile, err := ReadFuncZipFile("testfunc.zip")
	if err != nil {
		return
	}

	code.ZipFile = funcfile
	code.Publish = true
	data, err := json.Marshal(code)

	url := URL + "/functions/" + name + "/code"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(data)))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("CreateFunction response Status:", resp.Status)
	fmt.Println("CreateFunction response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("CreateFunction response Body:", string(body))
}

func DeleteFunction(name string) {
	var body []byte
	url := URL + "/functions/" + name
	fmt.Println(url)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("DeleteFunction response Status:", resp.Status)
	fmt.Println("DeleteFunction response Headers:", resp.Header)
	body1, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("DeleteFunction response Body:", string(body1))

}

func GetFunction(name string) {

	url := URL + "/functions/" + name
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("GetFunction response Status:", resp.Status)
	fmt.Println("GetFunction response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("GetFunction response Body:", string(body))

	var f model.Function
	e := json.Unmarshal(body, &f)
	if e != nil {
		fmt.Println("Unmarshal error")
		panic(e)
	}

	/*
		fn := f.FunctionName + ".zip"
		err1 := ioutil.WriteFile(fn, f.FuncCode.ZipFile, 0666)
		if err1 != nil {
			fmt.Println("error write file")
		}
	*/
}

func GetFunctionList(marker string) {
	url := URL + "/functions/?Marker=" + marker + "&MaxItems=2"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("GetFunctionList response Status:", resp.Status)
	fmt.Println("GetFunctionList response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("GetFunctionList response Body:", string(body))
}

func GetFunctionCode(name string) {

	url := URL + "/functions/" + name + "/code"
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("GetFunctionCode response Status:", resp.Status)
	fmt.Println("GetFunctionCode response Headers:", resp.Header)
	body1, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println("GetFunctionCode response Body:", string(body1))

	var v struct{ Code string }
	e := json.Unmarshal(body1, &v)

	if e != nil {
		fmt.Println("GetFunctionCode Decode error: ", e)

	}

	//fmt.Println("zip content:", v.Zip)

	vv, _ := base64.StdEncoding.DecodeString(v.Code)

	fn := name + ".zip"
	err1 := ioutil.WriteFile(fn, vv, 0666)
	if err1 != nil {
		fmt.Println("GetFunctionCode, error write file")
	}

}

func main() {

	args := os.Args
	if len(args) < 3 {
		fmt.Println("usage: client <ACTION> <FunctionParam>")
		fmt.Println("ACTION param:")
		fmt.Println("  create FunctionName: create funciton")
		fmt.Println("  get FunctionName: get funciton")
		fmt.Println("  getlist Marker: get funciton list")
		fmt.Println("  getcode FunctionName: get funciton code")
		fmt.Println("  delete FunctionName: delete funciton")
		fmt.Println("  update FunctionName: update funciton")
		fmt.Println("  invocate FunctionName: invocate funciton")
		return
	}
	action := args[1]
	param := args[2]

	switch action {
	case "create":
		CreateFunction(param)
	case "get":
		GetFunction(param)
	case "getlist":
		GetFunctionList(param)
	case "getcode":
		GetFunctionCode(param)
	case "updatecode":
		UpdateFunctionCode(param)
	case "delete":
		DeleteFunction(param)
	default:
		fmt.Println("default", action)
		return
	}

}

