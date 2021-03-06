package controllers

import (
	"github.com/astaxie/beego"
	"github.com/hoisie/redis"
	"log"
)

func initialRedis() *redis.Client {
	db, _ := beego.AppConfig.Int("redisdb")
	rc := &redis.Client{
		Addr:     beego.AppConfig.String("redisaddr"),
		Db:       db,
		Password: beego.AppConfig.String("redispassword"),
	}
	return rc
}

type Host struct {
	Domain string `form:"domain"`
	IP     string `form:"ip"`
}

type DNSController struct {
	beego.Controller
	rc *redis.Client
}

// http basic auth
// init redis connect
func (c *DNSController) Prepare() {
	CheckAuth(c.Ctx)
	c.rc = initialRedis()
}

func (c *DNSController) Get() {
	var HostsRecord = make(map[string]string)
	bindkey := beego.AppConfig.String("bindkey")
	err := c.rc.Hgetall(bindkey, HostsRecord)
	if err != nil {
		panic(err)
	}
	c.Data["Redis"] = beego.AppConfig.String("redisaddr")
	c.Data["Hosts"] = HostsRecord
	c.Layout = "layout.html"
	c.TplNames = "dns.html"
	c.Render()
}

func (c *DNSController) Post() {
	h := new(Host)
	if err := c.ParseForm(h); err != nil {
		c.Ctx.Abort(400, "Invalid post data")
		return
	}
	if h.Domain == "" || h.IP == "" {
		c.Ctx.Abort(400, "Both domain and ip needed")
		return
	}
	bindkey := beego.AppConfig.String("bindkey")

	if _, err := c.rc.Hset(bindkey, h.Domain, []byte(h.IP)); err != nil {
		c.Ctx.Abort(500, "Save hosts record failed")
		log.Println(err)
		return
	}
	log.Printf("Insert [%s:%s] into redis", h.Domain, h.IP)

}

type DNSDelController struct {
	beego.Controller
	rc *redis.Client
}

func (c *DNSDelController) Prepare() {
	CheckAuth(c.Ctx)
	c.rc = initialRedis()
}

func (c *DNSDelController) Post() {
	h := new(Host)
	if err := c.ParseForm(h); err != nil {
		c.Ctx.Abort(400, "Invalid post data")
		return
	}
	bindkey := beego.AppConfig.String("bindkey")
	if ok, err := c.rc.Hdel(bindkey, h.Domain); !ok {
		c.Ctx.Abort(500, "Delete hosts record failed")
		log.Println(err)
		return
	}
	log.Printf("Delete [%s:%s] from redis", h.Domain, h.IP)

}
