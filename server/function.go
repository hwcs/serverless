package server

import (
	"encoding/json"
	"fmt"

	"serverless/model"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

const MAX_CODE_SIZE = 1024 * 1024 * 100 // 代码不超过100M

func CreateFunction(c *gin.Context) {
	var function model.Function

	if c.BindJSON(&function) == nil {
		conn, err := redis.Dial("tcp", ":6379")
		if err == nil {
			defer conn.Close()

			hash := "/functions/" + function.FunctionName
			fmt.Println("HMSET hash:", hash)

			codesize := len(function.FuncCode.ZipFile)

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

			conn.Do("HMSET", hash,
				"CodeSha256", "xxxx",
				"Code.S3Bucket", function.FuncCode.S3Bucket, "Code.S3Key", function.FuncCode.S3Key,
				"Code.S3ObjectVersion", function.FuncCode.S3ObjectVersion, "Code.ZipFile", function.FuncCode.ZipFile,
				"FunctionName", function.FunctionName, "Description", function.Description,
				"Handler", function.Handler, "MemorySize", function.MemorySize, "Publish", function.Publish,
				"Runtime", function.Runtime, "Timeout", function.Timeout)

			timestamp := time.Now().Unix()
			tm := time.Unix(timestamp, 0)

			c.JSON(201, gin.H{"CodeSha256": "xxxx",
				"CodeSize":     len(function.FuncCode.ZipFile),
				"Description":  function.Description,
				"FunctionName": function.FunctionName,
				"Handler":      function.Handler,
				"LastModified": tm.Format("2006-01-02 15:04:05"),
				"MemorySize":   function.MemorySize,
				"Runtime":      function.Runtime,
				"Timeout":      function.Timeout})

		} else {

			c.JSON(500, gin.H{"ServiceException": "Internal error"})
		}
	}

}

func GetFunction(c *gin.Context) {
	name := c.Param("name")
	hash := "/functions/" + name

	conn, err := redis.Dial("tcp", ":6379")
	defer conn.Close()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}

	reply, _ := redis.Values(conn.Do("HGETALL", hash))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}

	n := len(reply)
	if n == 0 {
		c.JSON(404, gin.H{"ResourceNotFoundException": "not found function"})
		return
	}

	r := make(map[string]interface{})
	mc := make(map[string]interface{})

	for len(reply) > 0 {
		var k, v string

		reply, err = redis.Scan(reply, &k, &v)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{"ServiceException": "Internal Error"})
			return
		}

		switch k {
		case "Code.S3Bucket":
			mc["S3Bucket"] = v

		case "Code.S3Key":
			mc["S3Key"] = v

		case "Code.S3ObjectVersion":
			mc["S3ObjectVersion"] = v

		case "Code.ZipFile":
			mc["ZipFile"] = v

		case "FunctionName", "Description", "Handler", "Runtime":
			r[k] = v
		case "MemorySize", "Timeout":
			r[k], _ = strconv.Atoi(v)

		case "Publish":
			temp, _ := strconv.Atoi(v)
			if temp == 1 {
				r[k] = true
			} else {
				r[k] = false
			}

		default:
			break

		}

	}
	r["FuncCode"] = mc

	mi, _ := json.MarshalIndent(r, "", "")
	fmt.Println("Marshal indent:", string(mi))

	c.JSON(200, gin.H(r))

}

func DeleteFunction(c *gin.Context) {
	name := c.Param("name")
	hash := "/functions/" + name

	conn, err := redis.Dial("tcp", ":6379")
	defer conn.Close()
	if err != nil {
		c.JSON(500, gin.H{"ServiceException": "Internal Error"})
		return
	}

	reply, err := redis.Int(conn.Do("DEL", hash))

	fmt.Println("delete rows:", reply)
	if reply != 0 {
		c.Status(204)
		//c.JSON(204)
	} else {
		c.JSON(404, gin.H{"ResourceNotFoundException": "Function Not found"})
	}

}

