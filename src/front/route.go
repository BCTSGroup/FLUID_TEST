package front

import (
	"DAG-Exp/src/front/function"
	"net/http"
)
// 路由结构
type Route struct {
	routeName   string
	Method      string
	HandlerFunc http.HandlerFunc
}
// 路由表结构
type Routes []Route
var ALLRoutes = Routes{
	// 同步消息测试
	Route{
		routeName:   "/sync",
		Method:      "POST",
		HandlerFunc: function.HandleSync,
	},
	Route{
		"/GetLocalAccounts",
		"GET",
		function.HandleGetLocalAccounts,
	},

	// 创建 DAG
	Route{
		"/CreateDAG",
		"POST",
		function.HandleCreateDAG,
	},
	Route{
	"/GetDag",
	"GET",
	function.HandleGetDag,
	},
	// 请求数据
	Route{
		"/RequestData",
		"POST",
		function.HandleRequestData,
	},
	Route{
		routeName:   "/GetLatency",
		Method:      "GET",
		HandlerFunc: function.HandleLatency,
	},
	Route{
		routeName: "/Submit",
		Method: "POST",
		HandlerFunc: function.HandleSubmit,
	},
	Route{
		routeName:   "/EpochLatency",
		Method:      "GET",
		HandlerFunc: function.HandleEpochLatency,
	},
	Route{
		routeName:   "/GetTips",
		Method:      "GET",
		HandlerFunc: function.HandleGetTips,
	},
}


