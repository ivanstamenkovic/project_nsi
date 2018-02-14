package controllers

import (
	"encoding/json"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/go-redis/redis"
	"github.com/ivanstamenkovic/project_nsi/models"
)

type RedisController struct {
	beego.Controller
}

var DnsClient *redis.Client
var ValidationClient *redis.Client

func init() {
	DnsClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	ValidationClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func (this *RedisController) ResolveIP() {
	req := struct {
		url string
	}{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &req)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = err.Error()
	} else {
		ipAddress, errResolve := DnsClient.Get(req.url).Result()

		if errResolve != nil {
			this.Ctx.Output.SetStatus(400)
			this.Data["json"] = errResolve.Error()
		} else {
			this.Ctx.Output.SetStatus(200)
			this.Data["json"] = struct{ ip string }{ipAddress}
		}
	}
	this.ServeJSON()
}

func (this *RedisController) ServerCheckIn() {
	ipAddress := this.Ctx.Input.IP()
	if ipAddress == "::1" {
		ipAddress = "127.0.0.1"
	}

	req := struct {
		Url      string
		Username string
	}{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &req)

	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = err.Error()
	} else {
		finalURL := req.Url + "." + req.Username + ".idee.com"
		beego.Debug(ipAddress)
		errDNS := DnsClient.Set(finalURL, ipAddress, 90*time.Second).Err()
		if errDNS != nil {
			beego.Debug(errDNS)
			this.Ctx.Output.SetStatus(400)
			this.Data["json"] = errDNS.Error()
		} else {
			this.Ctx.Output.SetStatus(200)
			this.Data["json"] = "Success"
		}

	}
	this.ServeJSON()
}

func (this *RedisController) VerifyUser() {
	verificationLink := this.Ctx.Input.Param(":link")
	if verificationLink == "" {
		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = "No link supplied"
	} else {
		id, err := ValidationClient.Get(verificationLink).Result()
		if err != nil {
			this.Ctx.Output.SetStatus(400)
			this.Data["json"] = err.Error()
		} else {
			var user models.User
			o := orm.NewOrm()
			errFind := o.QueryTable("user").Filter("Username", id).One(&user)

			if errFind != nil {
				this.Ctx.Output.SetStatus(400)
				this.Data["json"] = errFind.Error()
			} else {
				user.Verified = true
				o.Update(&user)

				this.Ctx.Output.SetStatus(200)
				this.Data["json"] = user
			}
		}
	}

	this.ServeJSON()
}
