package hprose

import (
	"os"

	"fmt"
	"path"
	"strings"

	"github.com/cisordeng/bee/cmd/commands"
	"github.com/cisordeng/bee/cmd/commands/version"
	"github.com/cisordeng/bee/generate"
	"github.com/cisordeng/bee/logger"
	"github.com/cisordeng/bee/utils"
)

var CmdHproseapp = &commands.Command{
	// CustomFlags: true,
	UsageLine: "hprose [appname]",
	Short:     "Creates an RPC application based on Hprose and Beego frameworks",
	Long: `
  The command 'hprose' creates an RPC application based on both Beego and Hprose (http://hprose.com/).

  {{"To scaffold out your application, use:"|bold}}

      $ bee hprose [appname] [-tables=""] [-driver=mysql] [-conn="root:@tcp(127.0.0.1:3306)/test"]

  If 'conn' is empty, the command will generate a sample application. Otherwise the command
  will connect to your database and generate models based on the existing tables.

  The command 'hprose' creates a folder named [appname] with the following structure:

	    ├── main.go
	    ├── {{"conf"|foldername}}
	    │     └── app.conf
	    └── {{"models"|foldername}}
	          └── object.go
	          └── user.go
`,
	PreRun: func(cmd *commands.Command, args []string) { version.ShowShortVersionBanner() },
	Run:    createhprose,
}

func init() {
	CmdHproseapp.Flag.Var(&generate.Tables, "tables", "List of table names separated by a comma.")
	CmdHproseapp.Flag.Var(&generate.SQLDriver, "driver", "Database driver. Either mysql, postgres or sqlite.")
	CmdHproseapp.Flag.Var(&generate.SQLConn, "conn", "Connection string used by the driver to connect to a database instance.")
	commands.AvailableCommands = append(commands.AvailableCommands, CmdHproseapp)
}

var APIModels = `package models
import (
	"errors"
	"strconv"
	"time"
)
var (
	Objects map[string]*Object
)
type Object struct {
	ObjectId   string
	Score      int64
	PlayerName string
}
func init() {
	Objects = make(map[string]*Object)
	Objects["hjkhsbnmn123"] = &Object{"hjkhsbnmn123", 100, "cisordeng"}
	Objects["mjjkxsxsaa23"] = &Object{"mjjkxsxsaa23", 101, "someone"}
}
func AddOne(object Object) (ObjectId string) {
	object.ObjectId = "cisordeng" + strconv.FormatInt(time.Now().UnixNano(), 10)
	Objects[object.ObjectId] = &object
	return object.ObjectId
}
func GetOne(ObjectId string) (object *Object, err error) {
	if v, ok := Objects[ObjectId]; ok {
		return v, nil
	}
	return nil, errors.New("ObjectId Not Exist")
}
func GetAll() map[string]*Object {
	return Objects
}
func Update(ObjectId string, Score int64) (err error) {
	if v, ok := Objects[ObjectId]; ok {
		v.Score = Score
		return nil
	}
	return errors.New("ObjectId Not Exist")
}
func Delete(ObjectId string) {
	delete(Objects, ObjectId)
}
`

var APIModels2 = `package models
import (
	"errors"
	"strconv"
	"time"
)
var (
	UserList map[string]*User
)
func init() {
	UserList = make(map[string]*User)
	u := User{"user_11111", "cisordeng", "11111", Profile{"male", 20, "Singapore", "cisordeng@gmail.com"}}
	UserList["user_11111"] = &u
}
type User struct {
	Id       string
	Username string
	Password string
	Profile  Profile
}
type Profile struct {
	Gender  string
	Age     int
	Address string
	Email   string
}
func AddUser(u User) string {
	u.Id = "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	UserList[u.Id] = &u
	return u.Id
}
func GetUser(uid string) (u *User, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}
func GetAllUsers() map[string]*User {
	return UserList
}
func UpdateUser(uid string, uu *User) (a *User, err error) {
	if u, ok := UserList[uid]; ok {
		if uu.Username != "" {
			u.Username = uu.Username
		}
		if uu.Password != "" {
			u.Password = uu.Password
		}
		if uu.Profile.Age != 0 {
			u.Profile.Age = uu.Profile.Age
		}
		if uu.Profile.Address != "" {
			u.Profile.Address = uu.Profile.Address
		}
		if uu.Profile.Gender != "" {
			u.Profile.Gender = uu.Profile.Gender
		}
		if uu.Profile.Email != "" {
			u.Profile.Email = uu.Profile.Email
		}
		return u, nil
	}
	return nil, errors.New("User Not Exist")
}
func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}
func DeleteUser(uid string) {
	delete(UserList, uid)
}
`

func createhprose(cmd *commands.Command, args []string) int {
	output := cmd.Out()

	if len(args) != 1 {
		beeLogger.Log.Fatal("Argument [appname] is missing")
	}

	curpath, _ := os.Getwd()
	if len(args) > 1 {
		cmd.Flag.Parse(args[1:])
	}
	apppath, packpath, err := utils.CheckEnv(args[0])
	if err != nil {
		beeLogger.Log.Fatalf("%s", err)
	}
	if generate.SQLDriver == "" {
		generate.SQLDriver = "mysql"
	}
	beeLogger.Log.Info("Creating Hprose application...")

	os.MkdirAll(apppath, 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", apppath, "\x1b[0m")
	os.Mkdir(path.Join(apppath, "conf"), 0755)
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "conf"), "\x1b[0m")
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "conf", "app.conf"), "\x1b[0m")
	utils.WriteToFile(path.Join(apppath, "conf", "app.conf"),
		strings.Replace(generate.Hproseconf, "{{.Appname}}", args[0], -1))

	if generate.SQLConn != "" {
		beeLogger.Log.Infof("Using '%s' as 'driver'", generate.SQLDriver)
		beeLogger.Log.Infof("Using '%s' as 'conn'", generate.SQLConn)
		beeLogger.Log.Infof("Using '%s' as 'tables'", generate.Tables)
		generate.GenerateHproseAppcode(string(generate.SQLDriver), string(generate.SQLConn), "1", string(generate.Tables), path.Join(curpath, args[0]))

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "main.go"), "\x1b[0m")
		maingoContent := strings.Replace(generate.HproseMainconngo, "{{.Appname}}", packpath, -1)
		maingoContent = strings.Replace(maingoContent, "{{.DriverName}}", string(generate.SQLDriver), -1)
		maingoContent = strings.Replace(maingoContent, "{{HproseFunctionList}}", strings.Join(generate.HproseAddFunctions, ""), -1)
		if generate.SQLDriver == "mysql" {
			maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/go-sql-driver/mysql"`, -1)
		} else if generate.SQLDriver == "postgres" {
			maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/lib/pq"`, -1)
		}
		utils.WriteToFile(path.Join(apppath, "main.go"),
			strings.Replace(
				maingoContent,
				"{{.conn}}",
				generate.SQLConn.String(),
				-1,
			),
		)
	} else {
		os.Mkdir(path.Join(apppath, "models"), 0755)
		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "models"), "\x1b[0m")

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "models", "object.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(apppath, "models", "object.go"), APIModels)

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "models", "user.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(apppath, "models", "user.go"), APIModels2)

		fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(apppath, "main.go"), "\x1b[0m")
		utils.WriteToFile(path.Join(apppath, "main.go"),
			strings.Replace(generate.HproseMaingo, "{{.Appname}}", packpath, -1))
	}
	beeLogger.Log.Success("New Hprose application successfully created!")
	return 0
}
