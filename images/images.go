package images

import (
	"fmt"

	"os"
	"os/exec"
	"serverless/util"
        "log"
)

func BuildFuncDockerImage(funcname string, codetype string, code string) (string, error) {
	tag, _ := util.WriteFuncfile(funcname, codetype, code)

	appPath := util.OPEN_LAMBDA_HOME + "/tmp/" + tag + "/"
	servercode := util.OPEN_LAMBDA_HOME + "/lambda-generator/pyserver/server.py"
	configfile := util.OPEN_LAMBDA_HOME + "/lambda-generator/pyserver/config.json"

	fmt.Printf("BuildFuncDockerImage, app path:%s, server code:%s\n", appPath, servercode)

	_, e := util.CopyFile(servercode, appPath+"/server.py")
	if e != nil {
		fmt.Println("BuildFuncDockerImage copy file error, ", servercode)
		return tag, e
	}
	_, e = util.CopyFile(configfile, appPath+"/config.json")
	if e != nil {
		fmt.Println("BuildFuncDockerImage copy file error, ", configfile)
		return tag, e
	}

	CreateDockfile(appPath)
	build := "docker build -t " + tag + " " + appPath
	fmt.Println("build image cmd: ", build)
	cmd := exec.Command("/bin/sh", "-c", build)
	err := cmd.Run()
	if err != nil {
		fmt.Println("BuildFuncDockerImage, %s", err.Error())
		return tag, err
	}

	err = PushFuncImage2DockerRegistry(tag)
	if err != nil {
		fmt.Println("PushFuncImage2DockerRegistry, %s", err.Error())
		return tag, err
	}
	util.DeleteFuncTmpDir(appPath)
	return tag, nil
}

func RemoveImage(img string) {
	if img == "" {
		fmt.Printf("RemoveImage Not Found the image")
		return
	}

	// remove image in host
	s1 := "docker rmi -f " + img
	cmd := exec.Command("/bin/sh", "-c", s1)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("RemoveImage: %s, error: %s\n", s1, err.Error)
	}
	fmt.Printf("RemoveImage succeed: %s\n", s1)

	// remove image in registry node
	s2 := "docker rmi -f " + util.REGISTRY_ADDRESS + "/" + img
	cmd = exec.Command("/bin/sh", "-c", s2)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("RemoveImage: %s, error: %s\n", s2, err.Error)
	}
	fmt.Printf("RemoveImage succeed: %s\n", s2)
}

// generate dockerfile
func CreateDockfile(path string) {

	dockerfile := path + "/Dockerfile"
	f, e := os.OpenFile(dockerfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if e != nil {
		fmt.Println("CreateDockfile, error open file")
	}
	fmt.Fprintln(f, "FROM alpine")
	//fmt.Fprintln(f, "ENV http_proxy http://10.67.202.161:3128")  // proxy环境下用户设置代理，测试用
	//fmt.Fprintln(f, "ENV https_proxy http://10.67.202.161:3128") // proxy环境下用户设置代理，测试用
	fmt.Fprintln(f, "RUN apk add --update python py-pip")

	//fmt.Fprintln(f, "RUN pip --proxy http://10.67.202.161:3128 --trusted-host pypi.python.org install rethinkdb") // proxy环境下用户设置代理，测试用
	//fmt.Fprintln(f, "RUN pip --proxy http://10.67.202.161:3128 --trusted-host pypi.python.org install Flask")     // proxy环境下用户设置代理，测试用
	//fmt.Fprintln(f, "ENV http_proxy \"\"")                                                                        // proxy环境下用户设置代理，测试用
	//fmt.Fprintln(f, "ENV https_proxy \"\"")    
	fmt.Fprintln(f, "RUN pip install rethinkdb")
	fmt.Fprintln(f, "RUN pip install Flask")   
	fmt.Fprintln(f, "ADD / /")
	fmt.Fprintln(f, "CMD python /server.py")

}

func PushFuncImage2DockerRegistry(appname string) error {
	img := util.REGISTRY_ADDRESS + "/" + appname
	tag := "docker tag " + appname + " " + img

	fmt.Println("docker tag:", tag)
	out, err := exec.Command("/bin/sh", "-c", tag).Output()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))

	pushimg := "docker push " + img
	fmt.Println("push image:", pushimg)

	out, err = exec.Command("/bin/sh", "-c", pushimg).Output()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))

	return err
}

