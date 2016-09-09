package images

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"serverless/config"
	"serverless/util"
)

func BuildFuncDockerImage(funcname string, codetype string, code string) (string, error) {

	imgName, _ := util.WriteFuncfile(funcname, codetype, code)

	appPath := config.Conf.Open_lambda_home + "/tmp/" + imgName + "/"
	servercode := config.Conf.Open_lambda_home + "/lambda-generator/pyserver/server.py"
	configfile := config.Conf.Open_lambda_home + "/lambda-generator/pyserver/config.json"

	fmt.Printf("BuildFuncDockerImage, app path:%s, server code:%s\n", appPath, servercode)

	_, e := util.CopyFile(servercode, appPath+"/server.py")
	if e != nil {
		fmt.Println("BuildFuncDockerImage copy file error, ", servercode)
		return imgName, e
	}
	_, e = util.CopyFile(configfile, appPath+"/config.json")
	if e != nil {
		fmt.Println("BuildFuncDockerImage copy file error, ", configfile)
		return imgName, e
	}

	CreateDockfile(appPath)

	fmt.Println("******* start building image....")
	start := time.Now()

	build := "docker build -t " + imgName + " " + appPath
	fmt.Println("build image cmd: ", build)
	cmd := exec.Command("/bin/sh", "-c", build)
	err := cmd.Run()
	if err != nil {
		fmt.Println("BuildFuncDockerImage, %s", err.Error())
		return imgName, err
	}
	fmt.Println("******* finish build image, spend time: ", time.Since(start))

	e = PushFuncImage2DockerRegistry(imgName)
	if e != nil {
		fmt.Println("PushFuncImage2DockerRegistry, %s", e.Error())
		return imgName, e
	}
	util.DeleteFuncTmpDir(appPath)
	return imgName, nil
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
	s2 := "docker rmi -f " + config.Conf.Registry_address + "/" + img
	cmd = exec.Command("/bin/sh", "-c", s2)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("RemoveImage: %s, error: %s\n", s2, err.Error)
	}
	fmt.Printf("RemoveImage succeed: %s\n", s2)
}

// generate dockerfile
func CreateDockfileBak(path string) {

	dockerfile := path + "/Dockerfile"
	f, e := os.OpenFile(dockerfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if e != nil {
		fmt.Println("CreateDockfile, error open file")
	}

	fmt.Fprintln(f, "FROM "+config.Conf.Registry_address+"/ubuntu-py")

	if config.Conf.Use_proxy == "true" {
		fmt.Fprintln(f, "ENV http_proxy "+config.Conf.Http_proxy)
		fmt.Fprintln(f, "ENV https_proxy "+config.Conf.Https_proxy)
	}
	fmt.Fprintln(f, "RUN apt-get -y install python-pip")
	fmt.Fprintln(f, "RUN pip --proxy "+config.Conf.Http_proxy+" --trusted-host pypi.python.org install rethinkdb")
	fmt.Fprintln(f, "RUN pip --proxy "+config.Conf.Http_proxy+" --trusted-host pypi.python.org install Flask")

	if config.Conf.Use_proxy == "true" {
		fmt.Fprintln(f, "ENV http_proxy \"\"")
		fmt.Fprintln(f, "ENV https_proxy \"\"")
	}

	fmt.Fprintln(f, "ADD / /")
	fmt.Fprintln(f, "CMD python /server.py")
}

// Create dockerfile from alpine base image
func CreateDockfile(path string) {

	dockerfile := path + "/Dockerfile"
	f, e := os.OpenFile(dockerfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if e != nil {
		fmt.Println("CreateDockfile, error open file")
	}

	fmt.Fprintln(f, "FROM "+config.Conf.Registry_address+"/"+config.Conf.Base_image)

	if config.Conf.Use_proxy == "true" {
		fmt.Fprintln(f, "ENV http_proxy "+config.Conf.Http_proxy)
		fmt.Fprintln(f, "ENV https_proxy "+config.Conf.Https_proxy)
	}

	if config.Conf.Base_image == "alpine" {
		fmt.Fprintln(f, "RUN apk add --update python py-pip")
	} else if config.Conf.Base_image == "ubuntu-py" {
		fmt.Fprintln(f, "RUN apt-get -y install python-pip")
	}

	if config.Conf.Use_proxy == "true" {
		fmt.Fprintln(f, "RUN pip --proxy "+config.Conf.Http_proxy+" --trusted-host pypi.python.org install rethinkdb")
		fmt.Fprintln(f, "RUN pip --proxy "+config.Conf.Http_proxy+" --trusted-host pypi.python.org install Flask")
	} else {
		fmt.Fprintln(f, "RUN pip install rethinkdb")
		fmt.Fprintln(f, "RUN pip install Flask")
	}

	if config.Conf.Use_proxy == "true" {
		fmt.Fprintln(f, "ENV http_proxy \"\"")
		fmt.Fprintln(f, "ENV https_proxy \"\"")
	}

	fmt.Fprintln(f, "ADD / /")
	fmt.Fprintln(f, "CMD python /server.py")
}

func PushFuncImage2DockerRegistry(name string) error {
	defer util.Trace("PushFuncImage2DockerRegistry")()

	img := config.Conf.Registry_address + "/" + name
	tagCmd := "docker tag " + name + " " + img

	fmt.Println("PushFuncImage2DockerRegistry, docker tag:", tagCmd)
	out, err := exec.Command("/bin/sh", "-c", tagCmd).Output()
	if err != nil {
		fmt.Println("PushFuncImage2DockerRegistry, docker tag error: ", err)
		return err
	}
	fmt.Println("PushFuncImage2DockerRegistry, docker tag output: ", out)

	pushImgCmd := "docker push " + img
	fmt.Println("PushFuncImage2DockerRegistry, push image:", pushImgCmd)

	out, err = exec.Command("/bin/sh", "-c", pushImgCmd).Output()
	if err != nil {
		fmt.Println("PushFuncImage2DockerRegistry, docker push error: ", err)
	} else {
		fmt.Println("PushFuncImage2DockerRegistry, docker push output: ", out)
	}
	return err
}
