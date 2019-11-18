// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package apiapp

import (
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	"github.com/cisordeng/bee/cmd/commands"
	"github.com/cisordeng/bee/cmd/commands/version"
	"github.com/cisordeng/bee/generate"
	"github.com/cisordeng/bee/logger"
	"github.com/cisordeng/bee/utils"
)

var CmdApiapp = &commands.Command{
	// CustomFlags: true,
	UsageLine: "api [appname]",
	Short:     "Creates a Beego API application",
	Long: `
  The command 'api' creates a Beego API application.

  {{"Example:"|bold}}
      $ bee api [appname] [-tables=""] [-driver=mysql] [-conn="root:@tcp(127.0.0.1:3306)/test"]

  If 'conn' argument is empty, the command will generate an example API application. Otherwise the command
  will connect to your database and generate models based on the existing tables.

  The command 'api' creates a folder named [appname] with the following structure:

	    ├── main.go
		├── .gitignore
	    ├── {{"conf"|foldername}}
	    │     └── app.conf
	    ├── {{"rest"|foldername}}
	    │     └── init.go
	    │     └── account
	    │           └── user.go
	    │           └── login_user.go
	    ├── {{"model"|foldername}}
	    │     └── init.go
	    │     └── account
	    │           └── user.go
	    └── {{"business"|foldername}}
	          └── init.go
	          └── account
	                └── user.go
	                └── user_repository.go
	                └── encode_user.go
	                └── auth_user_service.go
`,
	PreRun: func(cmd *commands.Command, args []string) { version.ShowShortVersionBanner() },
	Run:    createAPI,
}
var gitIgnore = `.idea/
*.tmp
{{.Appname}}
`
var apiConf = `appname = {{.Appname}}
httpport = 8080
runmode = dev
autorender = false
copyrequestbody = true
EnableDocs = true

[db]
DB_HOST = localhost
DB_PORT = 3306
DB_NAME = {{.Appname}}
DB_USER = {{.Appname}}
DB_PASSWORD = s:66668888
DB_CHARSET = utf8

[api]
apiUrl = http://localhost
enableSign = true
signSecret = 7d736a2822f8c005a8f034b477b23f27
signEffectiveSeconds = 15
aesCommonKey = 7d736a2822f8c005a8f034b477b23f27
`
var apiMain = `package main

import (
	"github.com/cisordeng/beego/xenon"

	_ "{{.Appname}}/model"
	_ "{{.Appname}}/rest"
)

func main() {
	xenon.Run()
}
`

var apiRest = `package account

import (
	"github.com/cisordeng/beego/xenon"

	bUser "{{.Appname}}/business/account"
)

type User struct {
	xenon.RestResource
}

func init () {
	xenon.RegisterResource(new(User))
}

func (this *User) Resource() string {
	return "account.user"
}

func (this *User) Params() map[string][]string {
	return map[string][]string{
		"PUT": []string{
			"name",
			"password",
			"avatar",
		},
	}
}

func (this *User) Put() {
	name := this.GetString("name", "")
	password := this.GetString("password", "")
	avatar := this.GetString("avatar", "")

	user := bUser.NewUser(name, password, avatar)
	data := bUser.EncodeUser(user)
	this.ReturnJSON(data)
}
`

var apiRestLogin = `package account

import (
	"github.com/cisordeng/beego/xenon"

	bUser "{{.Appname}}/business/account"
)

type LoginUser struct {
	xenon.RestResource
}

func init () {
	xenon.RegisterResource(new(LoginUser))
}

func (this *LoginUser) Resource() string {
	return "account.login_user"
}

func (this *LoginUser) Params() map[string][]string {
	return map[string][]string{
		"PUT": []string{
			"name",
			"password",
		},
	}
}

func (this *LoginUser) Put() {
	name := this.GetString("name", "")
	password := this.GetString("password", "")
	sid := bUser.AuthUser(name, password)
	if sid != "" {
		user := bUser.GetUserByName(name)
		data := bUser.EncodeUser(user)
		data["sid"] = sid
		this.ReturnJSON(data)
	} else {
		xenon.RaiseException("rest:name or password is wrong", "用户名或密码错误")
	}
}
`

var apiRestInit = `package rest

import (
	_ "{{.Appname}}/rest/account"
)

func init() {
}
`

var apiModel = `package account

import (
	"time"

	"github.com/cisordeng/beego/orm"
)

type User struct {
	Id int
	Name string
	Password string
	Avatar string
	CreatedAt time.Time `+"`orm:\"auto_now_add;type(datetime)\"`"+`
}

func (o *User) TableName() string {
	return "account_user"
}

func init() {
	orm.RegisterModel(new(User))
}
`

var apiModelInit = `package model

import (
	_ "{{.Appname}}/model/account"
)

func init() {
}
`

var apiBusiness = `package account

import (
	"time"

	"github.com/cisordeng/beego/orm"
	"github.com/cisordeng/beego/xenon"

	mUser "{{.Appname}}/model/account"

)

type User struct {
	Id int
	Name string
	Password string
	Avatar string
	CreatedAt time.Time
}

func init() {
}

func InitUserFromModel(model *mUser.User) *User {
	instance := new(User)
	instance.Id = model.Id
	instance.Name = model.Name
	instance.Password = model.Password
	instance.Avatar = model.Avatar
	instance.CreatedAt = model.CreatedAt

	return instance
}

func NewUser(name string, password string, avatar string) (user *User) {
	model := mUser.User{
		Name: name,
		Password: xenon.EncodeMD5(password),
		Avatar: avatar,
	}
	_, err := orm.NewOrm().Insert(&model)
	xenon.PanicNotNilError(err)
	return InitUserFromModel(&model)
}
`

var apiBusinessRepository = `package account

import (
	"github.com/cisordeng/beego/orm"
	"github.com/cisordeng/beego/xenon"

	mUser "{{.Appname}}/model/account"
)

func GetUserByName(name string) (user *User)  {
	model := mUser.User{}
	err := orm.NewOrm().QueryTable(&mUser.User{}).Filter(xenon.Map{
		"name": name,
	}).One(&model)
	xenon.PanicNotNilError(err, "raise:account:not_exits", "用户不存在")
	user = InitUserFromModel(&model)
	return user
}
`

var apiBusinessEncode = `package account

import (
	"github.com/cisordeng/beego/xenon"
)

func EncodeUser(user *User) xenon.Map {
	mapUser := xenon.Map{
		"id": user.Id,
		"name": user.Name,
		"avatar": user.Avatar,
		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	return mapUser
}
`

var apiBusinessAuth = `package account

import (
	"encoding/json"

	"github.com/cisordeng/beego"
	"github.com/cisordeng/beego/xenon"
)

func AuthUser(name string, password string) string {
	user := GetUserByName(name)
	userMap := EncodeUser(user)
	if user.Password == xenon.EncodeMD5(password) {
		decodedByteToken, err := json.Marshal(userMap)
		xenon.PanicNotNilError(err)
		decodedToken := string(decodedByteToken)

		commonKey := beego.AppConfig.String("api::aesCommonKey")
		sid, err := xenon.EncodeAesWithCommonKey(decodedToken, commonKey)
		xenon.PanicNotNilError(err)
		return sid
	}
	return ""
}
`

var apiBusinessInit = `package business

func init() {
}
`

func init() {
	CmdApiapp.Flag.Var(&generate.Tables, "tables", "List of table names separated by a comma.")
	CmdApiapp.Flag.Var(&generate.SQLDriver, "driver", "Database driver. Either mysql, postgres or sqlite.")
	CmdApiapp.Flag.Var(&generate.SQLConn, "conn", "Connection string used by the driver to connect to a database instance.")
	commands.AvailableCommands = append(commands.AvailableCommands, CmdApiapp)
}

func createAPI(cmd *commands.Command, args []string) int {
	output := cmd.Out()

	if len(args) < 1 {
		beeLogger.Log.Fatal("Argument [appname] is missing")
	}

	if len(args) > 1 {
		err := cmd.Flag.Parse(args[1:])
		if err != nil {
			beeLogger.Log.Error(err.Error())
		}
	}

	appPath, _, err := utils.CheckEnv(args[0])
	appName := path.Base(args[0])
	if err != nil {
		beeLogger.Log.Fatalf("%s", err)
	}
	if generate.SQLDriver == "" {
		generate.SQLDriver = "mysql"
	}

	beeLogger.Log.Info("Creating API...")

	os.MkdirAll(appPath, 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", appPath, "\x1b[0m")
	os.Mkdir(path.Join(appPath, "conf"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf"), "\x1b[0m")
	os.Mkdir(path.Join(appPath, "rest"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "rest"), "\x1b[0m")
	os.Mkdir(path.Join(appPath, "model"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "model"), "\x1b[0m")
	os.Mkdir(path.Join(appPath, "business"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business"), "\x1b[0m")


	// config
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "conf", "app.conf"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "conf", "app.conf"),
		strings.Replace(apiConf, "{{.Appname}}", appName, -1))

	// rest
	os.Mkdir(path.Join(appPath, "rest", "account"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "rest", "account"), "\x1b[0m")
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "rest", "account", "user.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "rest", "account", "user.go"),
		strings.Replace(apiRest, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "rest", "account", "login_user.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "rest", "account", "login_user.go"),
		strings.Replace(apiRestLogin, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "rest", "init.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "rest", "init.go"),
		strings.Replace(apiRestInit, "{{.Appname}}", appName, -1))

	// business
	os.Mkdir(path.Join(appPath, "business", "account"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "account"), "\x1b[0m")
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "account", "user.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "business", "account", "user.go"),
		strings.Replace(apiBusiness, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "account", "user_repository.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "business", "account", "user_repository.go"),
		strings.Replace(apiBusinessRepository, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "account", "encode_user.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "business", "account", "encode_user.go"),
		strings.Replace(apiBusinessEncode, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "account", "auth_user_service.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "business", "account", "auth_user_service.go"),
		strings.Replace(apiBusinessAuth, "{{.Appname}}", appName, -1))
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "business", "init.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "business", "init.go"),
		strings.Replace(apiBusinessInit, "{{.Appname}}", appName, -1))

	// model
	os.Mkdir(path.Join(appPath, "model", "account"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "model", "account"), "\x1b[0m")
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "model", "account", "user.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "model", "account", "user.go"), apiModel)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "model", "init.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "model", "init.go"),
		strings.Replace(apiModelInit, "{{.Appname}}", appName, -1))

	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, "main.go"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, "main.go"),
		strings.Replace(apiMain, "{{.Appname}}", appName, -1))

	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(appPath, ".gitignore"), "\x1b[0m")
	utils.WriteToFile(path.Join(appPath, ".gitignore"),
		strings.Replace(gitIgnore, "{{.Appname}}", appName, -1))

	beeLogger.Log.Success("New API successfully created!")
	return 0
}
