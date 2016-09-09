package server

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"
	"serverless/config"
	"serverless/images"
	"serverless/model"
	"serverless/util"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

const MAX_CODE_SIZE = 1024 * 1024 * 100             // code file can't exceed 100M
const SERVERLESS_FUNC_LIST = "serverless_func_list" // code index table, store function code name into list

func getRedisConn() (redis.Conn, error) {
	return redis.Dial(config.Conf.Redis_network, config.Conf.Redis_address)
}

func CreateFunction(c *gin.Context) {
	defer util.Trace("CreateFunction")()

	var function model.Function
	e := c.BindJSON(&function)
	if e != nil {
		fmt.Println("CreateFunction, BindJSON error:", e)
		c.JSON(400, "Bad Request")
		return
	}

	conn, err := getRedisConn()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal error"})
		fmt.Println("CreateFunction, redis.Dial error:", err)
		return
	}

	defer conn.Close()

	hash := "/functions/" + function.FunctionName
	fmt.Println("HMSET hash:", hash)

	codesize := len(function.FuncCode.File)
	fmt.Printf("CreateFunction, codetype:%s, code size:%d\n", function.FuncCode.CodeType, codesize)

	if codesize > MAX_CODE_SIZE {
		errorinfo := "Code size: " + strconv.Itoa(codesize) + ", exceed max size: 100M"
		c.JSON(400, gin.H{"CodeStorageExceededException": errorinfo})
		return
	}
	exists, _ := redis.Bool(conn.Do("EXISTS", hash))
	if exists == true {
		c.JSON(409, gin.H{"ResourceConflictException": "Function existed"})
		return
	}

	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	t := tm.Format("2006-01-02 15:04:05")

	// insert record to function detail table
	_, err = conn.Do("HMSET", hash,
		"CodeSha256", "xxxx",
		"Code.S3Bucket", function.FuncCode.S3Bucket, "Code.S3Key", function.FuncCode.S3Key, "Code.S3ObjectVersion", function.FuncCode.S3ObjectVersion,
		"Code.File", function.FuncCode.File, "Code.CodeType", function.FuncCode.CodeType, "Code.CodeSize", codesize,
		"FunctionName", function.FunctionName, "Description", function.Description,
		"Handler", function.Handler, "MemorySize", function.MemorySize, "Publish", function.Publish,
		"Runtime", function.Runtime, "Timeout", function.Timeout, "LastModified", t)
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal error"})
		return
	}

	//insert record to functionlist table
	_, err = conn.Do("rpush", SERVERLESS_FUNC_LIST, hash)
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal error"})
		return
	}

	tag, err := images.BuildFuncDockerImage(function.FunctionName, function.FuncCode.CodeType, function.FuncCode.File)
	if err == nil {
		updateFuncImageLabel(function.FunctionName, tag)
		c.JSON(200, gin.H{"CodeSha256": "xxxx",
			"CodeSize":     codesize,
			"Description":  function.Description,
			"FunctionName": function.FunctionName,
			"Handler":      function.Handler,
			"LastModified": t,
			"MemorySize":   function.MemorySize,
			"Runtime":      function.Runtime,
			"Timeout":      function.Timeout})
		return
	}
	c.JSON(500, gin.H{"ServiceException": "Internal error"})

}

func updateFuncImageLabel(funcname string, tag string) error {
	conn, err := getRedisConn()

	if err == nil {
		defer conn.Close()
		hash := "/images/" + funcname
		_, e := conn.Do("HMSET", hash, "FuncName", funcname, "image", tag)
		if e != nil {
			fmt.Println("updateFuncImageLabel error", e)
			return e
		}
		return nil
	}
	return err
}

func GetFunctionList(c *gin.Context) {

	marker, _ := strconv.Atoi(c.Query("Marker"))
	maxItems, _ := strconv.Atoi(c.Query("MaxItems"))
	fmt.Printf("GetFunctionList, marker: %d, maxItems: %d\n", marker, maxItems)

	conn, e1 := getRedisConn()

	if e1 != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		fmt.Printf("GetFunctionList, redis.Dial:%s\n", e1.Error())
		return
	}
	defer conn.Close()
	reply, e2 := redis.Values(conn.Do("LRANGE", SERVERLESS_FUNC_LIST, marker, maxItems))
	if e2 != nil {
		fmt.Println(e2)
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	n := len(reply)
	rst := make([]map[string]interface{}, n)

	i := 0
	for len(reply) > 0 {
		var v string
		reply, e2 = redis.Scan(reply, &v)
		if e2 != nil {
			fmt.Println(e2)
			c.JSON(500, gin.H{"ServiceException": "Internal Error"})
			return
		}
		fmt.Println("Function Hash: ", v)
		r, s := getFunctionInfo(v)
		if s != 200 {
			c.JSON(s, gin.H(r.(map[string]interface{})))
			fmt.Println("GetFunctionList, not found function: ", v)
			continue
		}
		rst[i] = r.(map[string]interface{})
		i++
	}

	m := marker + maxItems
	if n < maxItems {
		m = marker + n
	}

	c.JSON(200, gin.H{"Functions": rst, "NextMarker": m})

}

func GetFunctionCode(c *gin.Context) {
	name := c.Param("name")
	hash := "/functions/" + name
	conn, e := getRedisConn()
	if e != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
	}
	reply, _ := redis.Values(conn.Do("HMGET", hash, "Code.File", "Code.CodeType"))

	if len(reply) > 0 {

		var v string
		var t string
		_, e = redis.Scan(reply, &v, &t)
		if e != nil {
			fmt.Println(e)
			c.JSON(500, gin.H{"ServiceException": "Internal Error"})
			return
		}

		c.JSON(200, gin.H{"CodeType": t, "Code": v})
		return
	}
	c.JSON(500, gin.H{"ServiceException": "Internal Error"})

}

func getFunctionInfo(hash string) (interface{}, int) {
	kv := make(map[string]interface{})
	conn, e1 := getRedisConn()

	defer conn.Close()
	if e1 != nil {
		kv = map[string]interface{}{"ServiceException": "Internal Error"}
		return kv, 500
	}

	exists, _ := redis.Bool(conn.Do("EXISTS", hash))
	if exists == false {
		kv = map[string]interface{}{"ResourceNotFoundException": "Function Not found"}
		return kv, 404
	}

	var k [9]string = [9]string{"CodeSha256", "Code.CodeSize", "Description", "FunctionName", "Handler",
		"LastModified", "MemorySize", "Runtime", "Timeout"}
	reply, e2 := redis.Values(
		conn.Do("HMGET", hash, k[0], k[1], k[2], k[3], k[4], k[5], k[6], k[7], k[8]))
	if e2 != nil {
		fmt.Println(e2)
		kv = map[string]interface{}{"ServiceException": "Internal Error"}
		return kv, 500
	}

	i := 0

	for len(reply) > 0 {
		var v string
		reply, e2 = redis.Scan(reply, &v)
		if e2 != nil {
			fmt.Println(e2)
			kv = map[string]interface{}{"ServiceException": "Internal Error"}
			return kv, 500
		}
		if k[i] == "Code.CodeSize" || k[i] == "MemorySize" || k[i] == "Timeout" {
			index := k[i]
			if index == "Code.CodeSize" {
				index = "CodeSize"
			}
			temp, _ := strconv.Atoi(v)
			kv[index] = temp
		} else {
			kv[k[i]] = v
		}

		i++
	}

	return kv, 200
}

func UpdateFunctionCode(c *gin.Context) {
	var f model.Function
	e := c.BindJSON(&f)
	if e != nil {
		fmt.Println("UpdateFunctionCode, BindJSON error:", e)
		c.JSON(400, "Bad Request")
		return
	}

	name := c.Param("name")
	hash := "/functions/" + name
	conn, err := getRedisConn()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	defer conn.Close()

	exists, _ := redis.Bool(conn.Do("EXISTS", hash))
	if exists == false {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
		return
	}

	codesize := len(f.FuncCode.File)
	fmt.Printf("UpdateFunctionCode, codetype:%s, code size:%d\n", f.FuncCode.CodeType, codesize)

	if codesize > MAX_CODE_SIZE {
		errorinfo := "Code size: " + strconv.Itoa(codesize) + ", exceed max size: 100M"
		c.JSON(400, gin.H{"CodeStorageExceededException": errorinfo})
		return
	}

	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	t := tm.Format("2006-01-02 15:04:05")

	// update code content in redis
	_, e = conn.Do("HMSET", hash, "Code.CodeType", f.FuncCode.CodeType, "Code.File", f.FuncCode.File, "Code.CodeSize", codesize, "LastModified", t)
	if e != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}

	// delete old function image
	imagehash := "/images/" + name
	img, err := getImageName(imagehash)
	if err == nil {
		images.RemoveImage(img)
	}

	// create new function image
	tag, err := images.BuildFuncDockerImage(f.FunctionName, f.FuncCode.CodeType, f.FuncCode.File)
	if err == nil {
		updateFuncImageLabel(f.FunctionName, tag)
	}

	r, s := getFunctionInfo(hash)
	if s != 200 {
		c.JSON(s, gin.H{"ServiceException": "Internal Error"})
		return
	}

	var rst = r.(map[string]interface{})
	c.JSON(200, gin.H(rst))
}

func UpdateFunctionConfig(c *gin.Context) {
	var f model.Function

	e := c.BindJSON(&f)
	if e != nil {
		fmt.Println("UpdateFunctionConfig, BindJSON error:", e)
		c.JSON(400, "Bad Request")
		return
	}

	name := c.Param("name")
	hash := "/functions/" + name
	conn, err := getRedisConn()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	defer conn.Close()

	exists, _ := redis.Bool(conn.Do("EXISTS", hash))
	if exists == false {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
		return
	}

	_, e = conn.Do("HMSET", hash, "Runtime", f.Runtime, "Handler", f.Handler, "Description", f.Description, "MemorySize", f.MemorySize, "Timeout", f.Timeout)
	if e != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	r, s := getFunctionInfo(hash)
	if s != 200 {
		c.JSON(s, gin.H{"ServiceException": "Internal Error"})
		return
	}

	var rst = r.(map[string]interface{})
	c.JSON(200, gin.H(rst))
}

func GetFunction(c *gin.Context) {
	name := c.Param("name")
	hash := "/functions/" + name

	conn, e1 := getRedisConn()

	if e1 != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	defer conn.Close()
	exists, _ := redis.Bool(conn.Do("EXISTS", hash))
	if exists == false {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
		return
	}

	var k [9]string = [9]string{"CodeSha256", "Code.CodeSize", "Description", "FunctionName", "Handler",
		"LastModified", "MemorySize", "Runtime", "Timeout"}
	reply, e2 := redis.Values(
		conn.Do("HMGET", hash, k[0], k[1], k[2], k[3], k[4], k[5], k[6], k[7], k[8]))
	if e2 != nil {
		fmt.Println(e2)
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}

	i := 0

	kv := make(map[string]interface{})

	for len(reply) > 0 {
		var v string
		reply, e2 = redis.Scan(reply, &v)
		if e2 != nil {
			fmt.Println(e2)
			c.JSON(500, gin.H{"ServiceException": "Internal Error"})
			return
		}
		if k[i] == "Code.CodeSize" || k[i] == "MemorySize" || k[i] == "Timeout" {
			index := k[i]
			if index == "Code.CodeSize" {
				index = "CodeSize"
			}
			temp, _ := strconv.Atoi(v)
			kv[index] = temp

		} else {
			kv[k[i]] = v
		}

		i++
	}

	mi, _ := json.MarshalIndent(kv, "", "")
	fmt.Println("GetFunction, Marshal indent:", string(mi))

	c.JSON(200, gin.H(kv))
}

func DeleteFunction(c *gin.Context) {
	name := c.Param("name")
	funchash := "/functions/" + name
	imagehash := "/images/" + name

	conn, err := getRedisConn()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}
	defer conn.Close()

	reply, err := redis.Int(conn.Do("DEL", funchash))

	fmt.Println("DeleteFunction, delete rows:", reply)
	if reply != 0 {
		c.Status(204)
	} else {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
	}

	_, err = redis.Int(conn.Do("LREM", SERVERLESS_FUNC_LIST, 0, funchash))

	if err == nil {
		c.Status(204)
	} else {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
	}

	//delete func image
	img, err := getImageName(imagehash)
	if err == nil {
		images.RemoveImage(img)
	}

	reply, _ = redis.Int(conn.Do("DEL", imagehash))

	fmt.Println("DeleteFunction, delete image rows:", reply)
}

func getImageName(hash string) (string, error) {
	conn, err := getRedisConn()
	var v string
	if err != nil {
		return v, err
	}
	defer conn.Close()

	reply, err := redis.Values(conn.Do("HMGET", hash, "image"))
	if err != nil {
		fmt.Println("getImageName error: ", err)
		return v, err
	}
	if len(reply) > 0 {
		_, err = redis.Scan(reply, &v)
		if err != nil {
			fmt.Println(err)

			return v, err
		}
	}
	fmt.Println("getImageName: ", v)
	return v, nil
}

func InvocateFunction(c *gin.Context) {
	defer util.Trace("InvocateFunction")()
	name := c.Param("name")
	imagehash := "/images/" + name
	img, err := getImageName(imagehash)

	c.Writer.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		fmt.Println("InvocateFunction, can not found image for the function: " + name)
		c.Writer.WriteHeader(http.StatusNotFound)
		body := "\"ResourceNotFoundException\": \"Function Not found\""
		c.Writer.Write([]byte(body))
		return
	}

	/*
		//////////////////////
		var payload map[string]interface{} = make(map[string]interface{})

		e := c.BindJSON(&payload)
		if e != nil {
			fmt.Printf("InvocateFunction %s, BindJSON error:%s\n", name, e.Error())
			c.JSON(400, "Bad Request")
			return
		}

		userName := payload["UserName"].(string)
		/////////////////////
	*/

	url := config.Conf.LoadBalancer_address + "/runLambda/" + img
	fmt.Printf("InvocateFunction, url:%s\n", url)
	/*
		buf, err := ioutil.ReadAll(c.Request.Body)
		fmt.Println("InvocateFunction, request body:", string(buf))
	*/
	req, err := http.NewRequest("POST", url, c.Request.Body)

	client := &http.Client{}
	resp, err := client.Do(req)
	status := http.StatusOK
	if err != nil {
		fmt.Println("InvocateFunction, error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	fmt.Println("InvocateFunction response Status:", resp.Status)
	fmt.Println("InvocateFunction response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("InvocateFunction response Body:", string(body))
	c.Writer.WriteHeader(status)
	c.Writer.Write(body)
}

