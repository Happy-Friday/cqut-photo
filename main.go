package main

import (
	"blog/tools"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

const (
	//教务系统返回的不存在的头像
	ErrImage   = "e56fe09f4771d1f0a9a3b87f338aad7046d3d8aa57b58a8f93ceb9043798d365"
	BreakFile  = "break.json"
	ConfigFile = "config.json"
	Introduce  = `
run: start to run script
clean: delete the break file
-----------------------------
config.json
{
	from: int, [the grade of start]
	to : int, [the grade of end]
	peopleCount: int, [the total number of a class]
	duration: float64, [how long does script run once]
	username: string,
	password: string
}
`
)

type ResponseOrder struct {
	Path string
	Rep  *http.Response
}

//图片写入文件
func WriteImage(name string, r io.ReadCloser) {
	name += ".gif"
	os.MkdirAll(path.Dir(name), os.ModeDir)

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic("读取文件失败，怀疑是被封锁，请重启脚本继续执行")
		return
	}
	if tools.Sha256(buf) != ErrImage {
		f, _ := os.Create(name)
		defer func() {
			f.Close()
			r.Close()
		}()
		f.Write(buf)
	}
}

//获取所以的学院代号+班级代码
func GetAllId(fname string) ([]string, error) {
	f, err := os.Open("id.txt")
	if err != nil {
		return []string{}, err
	}

	defer f.Close()
	var ids []string
	r := bufio.NewReader(f)

	for {
		if buf, _, err := r.ReadLine(); err == nil {
			ids = append(ids, string(buf))
		} else {
			break
		}
	}

	return ids, nil
}

func SaveJson(m *map[string]interface{}, filename string) error {
	fio, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer fio.Close()

	encoder := json.NewEncoder(fio)

	return encoder.Encode(m)
}

func ReadJson(filename string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	fio, err := os.Open(filename)

	if err != nil {
		return map[string]interface{}{}, err
	}

	reader := json.NewDecoder(fio)

	reader.Decode(&m)

	return m, nil
}

func ReadBreakPoint(gD, iD, cD int64, filename string) (int64, int64, int64) {
	m, _ := ReadJson(filename)

	gDefault, iDefault, cDefault := gD, iD, cD
	if v, ok := m["grade"]; ok {
		gDefault = int64(v.(float64))
	}
	if v, ok := m["index"]; ok {
		iDefault = int64(v.(float64))
	}
	if v, ok := m["count"]; ok {
		cDefault = int64(v.(float64))
	}

	return gDefault, iDefault, cDefault
}

func SaveBreakPoint(g, i, c int64, filename string) error {
	m := map[string]interface{}{
		"grade": g,
		"index": i,
		"count": c,
	}

	return SaveJson(&m, filename)
}

func DeleteBreak() {
	os.Remove(BreakFile)
}

func StartRun() {
	var wg sync.WaitGroup
	//获取基本配置文件
	//包括起始年级，结束年级，每个班的人数
	conf, err := ReadJson(ConfigFile)
	if err != nil {
		return
	}

	var (
		//结束年级
		to = int64(conf["to"].(float64))
		//开始年级
		from = int64(conf["from"].(float64))
		//爬虫间隔
		duration = time.Duration(conf["duration"].(float64))
		//一个班的人数
		peopleCount = int64(conf["peopleCount"].(float64))
		//用户名
		username = conf["username"].(string)
		//密码
		password = conf["password"].(string)
	)

	//新建cqut对象, 并且获取学工系统cookie
	cqut := NewCqut()
	cqut.Login(username, password)
	//获取所以的专业代号加学号
	ids, _ := GetAllId("id.txt")

	reps := make(chan *ResponseOrder, 1000)
	//启动goroutine来写文件
	go func() {
		for rep := range reps {
			WriteImage(rep.Path, rep.Rep.Body)
		}
	}()

	var (
		//成绩， 代码下标，班级计数
		grade, index, count int64
		gradeDefault        int64
		indexDefault        int64
		countDefault        int64
		countInit           = false
		indexInit           = false
	)

	gradeDefault, indexDefault, countDefault = ReadBreakPoint(from, 0, 1, BreakFile)
	for grade = gradeDefault; grade <= to; grade++ {
		//第一次读取配置文件按照指定值
		//后面从0开始
		if indexInit {
			index = 1
		} else {
			index = indexDefault
			indexInit = true
		}
		for ; index < int64(len(ids)); index++ {
			//第一次读取配置文件按照指定值
			//后面从1开始
			if countInit {
				count = 1
			} else {
				count = countDefault
				countInit = true
			}

			for ; count <= peopleCount; count++ {
				//保存目录
				dir := fmt.Sprintf("img/%02d/%s/", grade, ids[index])
				//学号
				snum := fmt.Sprintf("1%02d%s%02d", grade, ids[index], count)
				//保存路径
				path := dir + snum

				wg.Add(1)
				go func(snum, path string, cqut *Cqut) {
					defer wg.Done()
					log.Println("开始获取学号为", snum, "的同学....")
					reps <- &ResponseOrder{Path: path, Rep: cqut.GetHead(snum)()}
					log.Println("学号为", snum, "的同学抓取完毕")
				}(snum, path, cqut)

				//保存一次断点，失败后重启可以从原来位置开始抓取
				SaveBreakPoint(grade, index, count, BreakFile)
				//1.5s刷一次
				<-time.After(duration)
			}
		}
	}
	wg.Wait()
}

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		//运行脚本
		case "run":
			StartRun()
		//删除断点文件
		case "clean":
			DeleteBreak()
		case "help":
			fmt.Println(Introduce)
		}
	}
}
