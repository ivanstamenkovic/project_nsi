package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/astaxie/beego/orm"

	"golang.org/x/crypto/bcrypt"

	"github.com/astaxie/beego"
	"github.com/ivanstamenkovic/project_nsi/models"

	"github.com/mitchellh/mapstructure"

	"gopkg.in/gomail.v2"
)

type UserController struct {
	beego.Controller
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func createToken(username, password string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": username,
		"Password": password,
	})
	tokenString, errToken := token.SignedString([]byte("NeverDoThis"))

	return tokenString, errToken
}

func ValidateToken(tokenString string) (interface{}, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Encoding error")
		}
		return []byte("NeverDoThis"), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ret := struct {
			Username string
			Password string
		}{}
		mapstructure.Decode(claims, &ret)
		return ret, nil
	} else {
		return nil, fmt.Errorf("Invalid token")
	}
}

func (this *UserController) CreateUser() {

	req := struct {
		Username string
		Password string
		Email    string
	}{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &req)

	beego.Debug(req)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = err.Error()
	} else {

		hash, errHash := hashPassword(req.Password)
		if errHash != nil {
			this.Ctx.Output.SetStatus(400)
			this.Data["json"] = errHash.Error()
		} else {
			var user models.User
			user.Username = req.Username
			user.Password = req.Password
			user.Email = req.Email
			user.Verified = false
			user.Password = hash

			tokenString, errToken := createToken(user.Username, user.Password)
			if errToken != nil {
				this.Ctx.Output.SetStatus(500)
				this.Data["json"] = errToken.Error()
			} else {

				user.Token = tokenString

				o := orm.NewOrm()
				_, errSave := o.Insert(&user)

				if errSave != nil {
					this.Ctx.Output.SetStatus(400)
					this.Data["json"] = errSave.Error()
				} else {
					partLink := strconv.FormatInt(time.Now().Unix(), 10)
					fullLink := "127.0.0.1:3000/verify/" + partLink
					ValidationClient.Set(partLink, user.Username, 24*3600*time.Second)

					m := gomail.NewMessage()
					m.SetHeader("From", "idee.verify@gmail.com")
					m.SetHeader("To", user.Email)
					m.SetHeader("Subject", "Account Verification")
					m.SetBody("text/html", `<h1> Idee DNS</h1> 
						<br /> Click on the link below to verify your account
						<br /> <link href = "`+fullLink+`">`+fullLink+`</link></h2>`)

					d := gomail.NewDialer("smtp.gmail.com", 587, "idee.verify", "droptablegoogle")

					if errDial := d.DialAndSend(m); errDial != nil {
						this.Ctx.Output.SetStatus(400)
						this.Data["json"] = errDial.Error()
					} else {
						this.Ctx.Output.SetStatus(200) //can be ommited
						this.Data["json"] = user
					}

				}
			}
		}
	}
	beego.Debug(this.Data["json"])
	this.ServeJSON()
}

func (this *UserController) GetAllUsers() {

	o := orm.NewOrm()

	var users []*models.User
	o.QueryTable("user").All(&users)

	this.Data["json"] = users
	this.ServeJSON()

}

func (this *UserController) Login() {
	req := struct {
		Username string
		Password string
	}{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &req)

	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Data["json"] = err.Error()
	} else {
		o := orm.NewOrm()
		var user models.User
		err := o.QueryTable("user").Filter("Username", req.Username).One(&user)
		if err != nil {
			this.Data["json"] = err
		} else {
			if checkPasswordHash(req.Password, user.Password) {
				if user.Verified {
					this.Ctx.Output.SetStatus(200)
					this.Data["json"] = struct {
						Token string `json:"token"`
					}{user.Token}
				} else {
					this.Ctx.Output.SetStatus(400)
					this.Data["json"] = "User not verified"
				}

			} else {
				this.Ctx.Output.SetStatus(400)
				this.Data["json"] = "Incorrect password"
			}

		}
	}

	this.ServeJSON()
}
