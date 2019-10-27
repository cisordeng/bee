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

package generate

import (
	"fmt"
	"os"
	"path"
	"strings"

	beeLogger "github.com/cisordeng/bee/logger"
	"github.com/cisordeng/bee/logger/colors"
	"github.com/cisordeng/bee/utils"
)

var rest = `package {{.package_name}}

import (
	"github.com/cisordeng/beego/xenon"

	b{{.PackageName}} "{{.app_name}}/business/{{.package_name}}"
)

type {{.ResourceName}} struct {
	xenon.RestResource
}

func init () {
	xenon.RegisterResource(new({{.ResourceName}}))
}

func (this *{{.ResourceName}}) Resource() string {
	return "{{.package_name}}.{{.resource_name}}"
}

func (this *{{.ResourceName}}) Params() map[string][]string {
	return map[string][]string{
		"GET":  []string{
			"id",
		},
	}
}

func (this *{{.ResourceName}}) Get() {
	id, _ := this.GetInt("id", 0)

	{{.resourceName}} := b{{.PackageName}}.Get{{.ResourceName}}ById(id)
	data := b{{.PackageName}}.Encode{{.ResourceName}}({{.resourceName}})
	this.ReturnJSON(data)
}
`

var businessEntity = `package {{.package_name}}

import (
	"time"

	"github.com/cisordeng/beego/orm"
	"github.com/cisordeng/beego/xenon"

	m{{.PackageName}} "{{.app_name}}/model/{{.package_name}}"
)

type {{.ResourceName}} struct {
	Id int
	CreatedAt time.Time
}

func init() {
}

func Init{{.ResourceName}}FromModel(model *m{{.ResourceName}}.{{.ResourceName}}) *{{.ResourceName}} {
	instance := new({{.ResourceName}})
	instance.Id = model.Id
	
	instance.CreatedAt = model.CreatedAt

	return instance
}

func New{{.ResourceName}}() ({{.resourceName}} *{{.ResourceName}}) {
	model := m{{.ResourceName}}.{{.ResourceName}}{
		
	}
	_, err := orm.NewOrm().Insert(&model)
	xenon.PanicNotNilError(err)
	return Init{{.ResourceName}}FromModel(&model)
}
`

var businessRepository = `package {{.package_name}}

import (
	"github.com/cisordeng/beego/orm"
	"github.com/cisordeng/beego/xenon"

	m{{.PackageName}} "{{.app_name}}/model/{{.package_name}}"
)

func GetOne{{.ResourceName}}(filters xenon.Map) *{{.ResourceName}} {
	o := orm.NewOrm()
	qs := o.QueryTable(&m{{.PackageName}}.{{.ResourceName}}{})

	var model m{{.PackageName}}.{{.ResourceName}}
	if len(filters) > 0 {
		qs = qs.Filter(filters)
	}

	err := qs.One(&model)
	xenon.PanicNotNilError(err, "raise:{{.resource_name}}:not_exits", "{{.resource_name}}不存在")
	return Init{{.ResourceName}}FromModel(&model)
}

func Get{{.ResourceName}}s(filters xenon.Map, orderExprs ...string ) []*{{.ResourceName}} {
	o := orm.NewOrm()
	qs := o.QueryTable(&m{{.PackageName}}.{{.ResourceName}}{})

	var models []*m{{.PackageName}}.{{.ResourceName}}
	if len(filters) > 0 {
		qs = qs.Filter(filters)
	}
	if len(orderExprs) > 0 {
		qs = qs.OrderBy(orderExprs...)
	}

	_, err := qs.All(&models)
	xenon.PanicNotNilError(err)


	{{.resourceName}}s := make([]*{{.ResourceName}}, 0)
	for _, model := range models {
		{{.resourceName}}s = append({{.resourceName}}s, Init{{.ResourceName}}FromModel(model))
	}
	return {{.resourceName}}s
}

func Get{{.ResourceName}}ById(id int) *{{.ResourceName}} {
	return GetOne{{.ResourceName}}(xenon.Map{
		"id": id,
	})
}
`

var businessEncode = `package {{.package_name}}
import (
	"github.com/cisordeng/beego/xenon"
)

func Encode{{.ResourceName}}({{.resourceName}} *{{.ResourceName}}) xenon.Map {
	if {{.resourceName}} == nil {
		return nil
	}

	map{{.ResourceName}} := xenon.Map{
		"id": {{.resourceName}}.Id,
		
		"created_at": {{.resourceName}}.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	return map{{.ResourceName}}
}


func EncodeMany{{.ResourceName}}({{.resourceName}}s []*{{.ResourceName}}) []xenon.Map {
	map{{.ResourceName}}s := make([]xenon.Map, 0)
	for _, {{.resourceName}} := range {{.resourceName}}s {
		map{{.ResourceName}}s = append(map{{.ResourceName}}s, Encode{{.ResourceName}}({{.resourceName}}))
	}
	return map{{.ResourceName}}s
}
`

var model = `package {{.package_name}}

import (
	"time"
	
	"github.com/cisordeng/beego/orm"
)

type {{.ResourceName}} struct {
	Id int
	
	CreatedAt time.Time ` + "`orm:\"auto_now_add;type(datetime)\"`" + `
}

func (o *{{.ResourceName}}) TableName() string {
	return "{{.package_name}}_{{.resource_name}}"
}

func init() {
	orm.RegisterModel(new({{.ResourceName}}))
}
`

func GenerateResource(cname, currpath string) {
	w := colors.NewColorWriter(os.Stdout)

	inGoPath := ""
	for _, goPath := range utils.GetGOPATHs() {
		if strings.Contains(currpath, goPath) {
			inGoPath = goPath
			break
		}
	}
	if inGoPath == "" {
		beeLogger.Log.Fatal("Wrong generate resource command, current path is not $GOPATH")
	}
	appName := strings.Split(currpath[len(inGoPath) + 1:], "/")[1]
	cname = strings.Replace(cname, ".", "/", -1)
	p, f := path.Split(cname)
	resourceName := strings.Title(f)
	packageName := ""

	if p != "" {
		i := strings.LastIndex(p[:len(p)-1], "/")
		packageName = p[i+1 : len(p)-1]
	} else {
		beeLogger.Log.Fatal("Wrong generate resource command, it should like [bee generate resource package/resource]")
	}

	beeLogger.Log.Infof("Using '%s' as resource name", utils.CamelString(resourceName))
	beeLogger.Log.Infof("Using '%s' as package name", packageName)

	// rest
	restPath := path.Join(currpath, "rest", packageName)
	os.MkdirAll(restPath, 0755)
	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(restPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))), "\x1b[0m")
	utils.WriteToFile(path.Join(restPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))),
		replaceTpl(rest, appName, packageName, strings.ToLower(resourceName)))

	// business
	businessPath := path.Join(currpath, "business", packageName)
	os.MkdirAll(businessPath, 0755)
	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(businessPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))), "\x1b[0m")
	utils.WriteToFile(path.Join(businessPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))),
		replaceTpl(businessEntity, appName, packageName, strings.ToLower(resourceName)))

	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(businessPath, fmt.Sprintf("%s_repository.go", strings.ToLower(resourceName))), "\x1b[0m")
	utils.WriteToFile(path.Join(businessPath, fmt.Sprintf("%s_repository.go", strings.ToLower(resourceName))),
		replaceTpl(businessRepository, appName, packageName, strings.ToLower(resourceName)))

	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(businessPath, fmt.Sprintf("encode_%s.go", strings.ToLower(resourceName))), "\x1b[0m")
	utils.WriteToFile(path.Join(businessPath, fmt.Sprintf("encode_%s.go", strings.ToLower(resourceName))),
		replaceTpl(businessEncode, appName, packageName, strings.ToLower(resourceName)))

	// model
	modelPath := path.Join(currpath, "model", packageName)
	os.MkdirAll(modelPath, 0755)
	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(modelPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))), "\x1b[0m")
	utils.WriteToFile(path.Join(modelPath, fmt.Sprintf("%s.go", strings.ToLower(resourceName))),
		replaceTpl(model, appName, packageName, strings.ToLower(resourceName)))
}

func replaceTpl(tpl string, app string, package_name string, resource_name string) string {
	PackageName := utils.CamelCase(package_name)
	packageName := string(package_name[0]) + PackageName[1:]

	ResourceName := utils.CamelCase(resource_name)
	resourceName := string(resource_name[0]) + ResourceName[1:]

	a := strings.Replace(tpl, "{{.app_name}}", app, -1)
	p := strings.Replace(strings.Replace(strings.Replace(a, "{{.package_name}}", package_name, -1), "{{.packageName}}", packageName, -1), "{{.PackageName}}", PackageName, -1)
	return strings.Replace(strings.Replace(strings.Replace(p, "{{.resource_name}}", resource_name, -1), "{{.resourceName}}", resourceName, -1), "{{.ResourceName}}", ResourceName, -1)
}
