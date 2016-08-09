package server

import (
    "github.com/gin-gonic/gin"
    "github.com/garyburd/redigo/redis"
    "serverless/model"
)

func CreateFunction (c *gin.Context) {
    var function model.Function
    if c.BindJSON(&function) == nil {
        conn, err := redis.Dial("tcp", ":6379")
        if err == nil {
            defer conn.Close ()
            conn.Do("SET", "k1", 1)
        } else {
            c.JSON (500, gin.H{"err":"Internel error"})
        }
    } 
    c.JSON (200, gin.H{"CodeSize": 1000}) 
}

