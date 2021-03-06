/**********************************************
** @Des: base controller
** @Author: haodaquan
** @Date:   2017-09-07 16:54:40
** @Last Modified by:   haodaquan
** @Last Modified time: 2017-09-18 10:28:01
***********************************************/
package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"streamConsole/libs"
	"streamConsole/utils"

	"github.com/astaxie/beego"
	"github.com/patrickmn/go-cache"
	"streamConsole/models"
)

const (
	MSG_OK  = 0
	MSG_ERR = -1
)

type BaseController struct {
	beego.Controller
	controllerName string
	actionName     string
	user           *models.Admin
	userId         int
	userName       string
	loginName      string
	pageSize       int
	allowUrl       string
}

//前期准备
func (self *BaseController) Prepare() {
	self.pageSize = 20
	controllerName, actionName := self.GetControllerAndAction()
	self.controllerName = strings.ToLower(controllerName[0 : len(controllerName)-10])
	self.actionName = strings.ToLower(actionName)
	self.Data["appName"] = beego.AppConfig.String("appname")
	self.Data["version"] = beego.AppConfig.String("version")
	self.Data["siteName"] = beego.AppConfig.String("site.name")
	self.Data["curRoute"] = self.controllerName + "." + self.actionName
	self.Data["curController"] = self.controllerName
	self.Data["curAction"] = self.actionName
	// noAuth := "ajaxsave/ajaxdel/table/loginin/loginout/getnodes/start"
	// isNoAuth := strings.Contains(noAuth, self.actionName)

	if models.SerialValid {
		//已注册
		if (strings.Compare(self.controllerName, "apidoc")) != 0 {
			self.auth()
		}
	} else if strings.Compare(self.controllerName, "serial") != 0 {
		fmt.Println(self.controllerName)
		self.redirect(beego.URLFor("SerialController.Detail"))
	}

	self.Data["loginUserId"] = self.userId
	self.Data["loginUserName"] = self.userName
}

//登录权限验证
func (self *BaseController) auth() {
	arr := strings.Split(self.Ctx.GetCookie("auth"), "|")
	self.userId = 0
	if len(arr) == 2 {
		idstr, password := arr[0], arr[1]
		userId, _ := strconv.Atoi(idstr)
		if userId > 0 {
			var err error

			cheUser, found := utils.Che.Get("uid" + strconv.Itoa(userId))
			user := &models.Admin{}
			if found && cheUser != nil { //从缓存取用户
				user = cheUser.(*models.Admin)
			} else {
				user, err = models.AdminGetById(userId)
				utils.Che.Set("uid"+strconv.Itoa(userId), user, cache.DefaultExpiration)
			}
			if err == nil && password == libs.Md5([]byte(self.getClientIp()+"|"+user.Password+user.Salt)) {
				self.userId = user.Id

				self.loginName = user.LoginName
				self.userName = user.RealName
				self.user = user
				self.AdminAuth()
			}

			isHasAuth := strings.Contains(self.allowUrl, self.controllerName+"/"+self.actionName)
			//不需要权限检查
			noAuth := "ajaxsave/ajaxdel/table/loginin/loginout/getnodes/start/show/ajaxapisave/index/group/public/env/code/apidetail"
			isNoAuth := strings.Contains(noAuth, self.actionName)
			if isHasAuth == false && isNoAuth == false {
				self.Ctx.WriteString("没有权限")
				self.ajaxMsg("没有权限", MSG_ERR)
				return
			}
		}
	}

	if self.userId == 0 && (self.controllerName != "login" && self.actionName != "loginin") {
		self.redirect(beego.URLFor("LoginController.LoginIn"))
	}
}

func (self *BaseController) AdminAuth() {
	cheMen, found := utils.Che.Get("menu" + strconv.Itoa(self.user.Id))
	if found && cheMen != nil { //从缓存取菜单
		menu := cheMen.(*CheMenu)
		//fmt.Println("调用显示菜单")
		self.Data["SideMenu1"] = menu.List1 //一级菜单
		self.Data["SideMenu2"] = menu.List2 //二级菜单
		self.allowUrl = menu.AllowUrl
	} else {
		// 左侧导航栏
		filters := make([]interface{}, 0)
		filters = append(filters, "status", 1)
		if self.userId != 1 {
			//普通管理员
			adminAuthIds, _ := models.RoleAuthGetByIds(self.user.RoleIds)
			adminAuthIdArr := strings.Split(adminAuthIds, ",")
			filters = append(filters, "id__in", adminAuthIdArr)
		}
		result, _ := models.AuthGetList(1, 1000, filters...)
		list := make([]map[string]interface{}, len(result))
		list2 := make([]map[string]interface{}, len(result))
		allow_url := ""
		i, j := 0, 0
		for _, v := range result {
			if v.AuthUrl != " " || v.AuthUrl != "/" {
				allow_url += v.AuthUrl
			}
			row := make(map[string]interface{})
			if v.Pid == 1 && v.IsShow == 1 {
				row["Id"] = int(v.Id)
				row["Sort"] = v.Sort
				row["AuthName"] = v.AuthName
				row["AuthUrl"] = v.AuthUrl
				row["Icon"] = v.Icon
				row["Pid"] = int(v.Pid)
				list[i] = row
				i++
			}
			if v.Pid != 1 && v.IsShow == 1 {
				row["Id"] = int(v.Id)
				row["Sort"] = v.Sort
				row["AuthName"] = v.AuthName
				row["AuthUrl"] = v.AuthUrl
				row["Icon"] = v.Icon
				row["Pid"] = int(v.Pid)
				list2[j] = row
				j++
			}
		}
		self.Data["SideMenu1"] = list[:i]  //一级菜单
		self.Data["SideMenu2"] = list2[:j] //二级菜单

		self.allowUrl = allow_url + "/home/index"
		cheM := &CheMenu{}
		cheM.AllowUrl = self.allowUrl
		cheM.List1 = self.Data["SideMenu1"].([]map[string]interface{})
		cheM.List2 = self.Data["SideMenu2"].([]map[string]interface{})
		utils.Che.Set("menu"+strconv.Itoa(self.user.Id), cheM, cache.DefaultExpiration)
	}

}

type CheMenu struct {
	List1    []map[string]interface{}
	List2    []map[string]interface{}
	AllowUrl string
}

// 是否POST提交
func (self *BaseController) isPost() bool {
	return self.Ctx.Request.Method == "POST"
}

//获取用户IP地址
func (self *BaseController) getClientIp() string {
	s := self.Ctx.Request.RemoteAddr
	l := strings.LastIndex(s, ":")
	return s[0:l]
}

// 重定向
func (self *BaseController) redirect(url string) {
	self.Redirect(url, 302)
	self.StopRun()
}

//加载模板
func (self *BaseController) display(tpl ...string) {
	var tplname string
	if len(tpl) > 0 {
		tplname = strings.Join([]string{tpl[0], "html"}, ".")
	} else {
		tplname = self.controllerName + "/" + self.actionName + ".html"
	}
	self.Layout = "public/layout.html"
	self.TplName = tplname
}

//ajax返回
func (self *BaseController) ajaxMsg(msg interface{}, msgno int) {
	out := make(map[string]interface{})
	out["status"] = msgno
	out["message"] = msg
	self.Data["json"] = out
	self.ServeJSON()
	self.StopRun()
}

//ajax返回 列表
func (self *BaseController) ajaxList(msg interface{}, msgno int, count int64, data interface{}) {
	out := make(map[string]interface{})
	out["code"] = msgno
	out["msg"] = msg
	out["count"] = count
	out["data"] = data
	self.Data["json"] = out
	self.ServeJSON()
	self.StopRun()
}
