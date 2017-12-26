package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

const (
	BaseUrl      = "http://xgxt.i.cqut.edu.cn/xgxt/xsxx_xsgl.do?method=showPhoto&xh="
	URLPortal    = "http://i.cqut.edu.cn/portal.do"
	URLLoginGet  = "http://i.cqut.edu.cn/zfca/login?service=http%3A%2F%2Fi.cqut.edu.cn%2Fportal.do"
	URLLoginPost = "http://i.cqut.edu.cn/zfca/login"
	URLJWXT      = "http://i.cqut.edu.cn/zfca?yhlx=student&login=0122579031373493728&url=xs_main.aspx"
	URLXGXT      = "http://i.cqut.edu.cn/zfca?yhlx=student&login=122579031373493679&url=stuPage.jsp"
)

//Cookie罐头
//实现Jar接口
type Jar struct {
	cookieMap map[*url.URL][]*http.Cookie
}

func (this *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	this.cookieMap[u] = append(this.cookieMap[u], cookies...)
}

func (this *Jar) Cookies(u *url.URL) []*http.Cookie {
	return this.cookieMap[u]
}

//将jar中所有cookie加到request上
func (this *Jar) TachRequest(r *http.Request) {
	for _, cookie := range this.AllCookies() {
		r.AddCookie(cookie)
	}
}

//获取已经截获到的所有cookie
func (this *Jar) AllCookies() []*http.Cookie {
	cs := []*http.Cookie{}

	for _, cookies := range this.cookieMap {
		for _, v := range cookies {
			cs = append(cs, v)
		}
	}

	return cs
}

//Cqut类
//封装了获取学工系统，正方成绩cookie的方法
type Cqut struct {
	Jar   *Jar
	cli   *http.Client
	cliNo *http.Client
}

func NewCqut() *Cqut {
	cqut := new(Cqut)
	cqut.Jar = &Jar{
		cookieMap: make(map[*url.URL][]*http.Cookie),
	}
	//设置不跟随跳转的Client
	cqut.cliNo = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: cqut.Jar,
	}
	//设置跟随跳转的Client
	//并且把每次请求的截取的cookie都加到跟随的request上去
	cqut.cli = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			cqut.Jar.TachRequest(req)
			return nil
		},
		Jar: cqut.Jar,
	}

	return cqut
}

//登陆的第一个cookie获取， 不能跟随重定向
func (c *Cqut) portalDo() *http.Response {
	req, _ := http.NewRequest("GET", URLPortal, nil)

	log.Println("请求portal.do...")
	rep, _ := c.cliNo.Do(req)
	log.Println("请求portal.do成功...")
	return rep
}

//登陆的第二个cookie获取
func (c *Cqut) loginGet() *http.Response {
	req, _ := http.NewRequest("GET", URLLoginGet, nil)
	log.Println("请求login/GET...")
	rep, _ := c.cli.Do(req)
	log.Println("请求login/GET成功...")
	return rep
}

//登陆
func (c *Cqut) loginPost(v url.Values) *http.Response {
	req, _ := http.NewRequest("POST", URLLoginPost, strings.NewReader(v.Encode()))
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.101 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://i.cqut.edu.cn/zfca/login")
	req.Header.Set("Origin", "http://i.cqut.edu.cn")
	req.Header.Set("Host", "i.cqut.edu.cn")
	c.Jar.TachRequest(req)
	log.Println("请求login/POST...")
	rep, _ := c.cli.Do(req)
	log.Println("请求login/POST成功...")
	return rep
}

//获取学工系统cookie
func (c *Cqut) Xgxt() *http.Response {
	req, _ := http.NewRequest("GET", URLXGXT, nil)
	c.Jar.TachRequest(req)
	log.Println("请求学工系统...")
	rep, _ := c.cli.Do(req)
	log.Println("请求学工系统成功...")
	return rep
}

//获取获取学工系统的头像
//id 为学号
func (c *Cqut) GetHead(id string) func() *http.Response {
	return func() *http.Response {
		cli := &http.Client{}

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", BaseUrl, id), nil)

		c.Jar.TachRequest(req)
		rep, _ := cli.Do(req)
		return rep
	}
}

//获取所以要进入学工系统的cookies
func (c *Cqut) Login(username, password string) {
	c.portalDo()
	doc, _ := goquery.NewDocumentFromResponse(c.loginGet())
	lt, _ := doc.Find(`input[name="lt"]`).Attr("value")
	c.loginPost(url.Values{
		"lt":              {lt},
		"ip":              {""},
		"username":        {username},
		"password":        {password},
		"_eventId":        {"submit"},
		"useValidateCode": {"0"},
		"isremenberme":    {"0"},
		"losetime":        {"30"},
	})
	c.Xgxt()
}
