package front

import (
	"bfc/src/front/function"
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
	// 接入请求
	Route{
		routeName:   "/JoinRequest",
		Method:      "POST",
		HandlerFunc: function.HandleJoinRequest,
	},
	// 查询共享规则
	Route{
		routeName:   "/JoinRequest/{From}",
		Method:      "GET",
		HandlerFunc: function.HandleSearchRule,
	},
	// 提交共享数据
	Route{
		routeName:   "/SubmitSharingData",
		Method:      "POST",
		HandlerFunc: function.HandleSubmitRequest,
	},
	// 请求数据共享
	Route{
		routeName:   "/GetSharingData",
		Method:      "POST",
		HandlerFunc: function.HandleGetSharingData,
	},
	// 测试专用 同步区块头
	Route{
		routeName:   "/SyncBlockHead/{LocalHeight}",
		Method:      "GET",
		HandlerFunc: function.HandleSyncBlockHead,
	},
	Route{
		"/GetBlockInfo",
		"GET",
		function.HandleGetBlockInfo,
	},
	Route{
		"/GetLocalAccounts",
		"GET",
		function.HandleGetLocalAccounts,
	},
	Route{
		"/GetNewBlock",
		"GET",
		function.HandleGetNewBlock,
	},
	Route{
		routeName:   "/TestTps",
		Method:      "POST",
		HandlerFunc: function.HandleTestTps,
	},
	Route{
		routeName:   "/GetTxPool",
		Method:      "GET",
		HandlerFunc: function.HandleGetTxPool,
	},
	Route{
		routeName:   "/RequestData",
		Method:      "POST",
		HandlerFunc: function.HandleRequestData,
	},
	Route{
		"/GetLatency",
		"GET",
		function.HandleGetLatency,
	},
	//Route{
	//	"/contract/{contractID}",
	//	"GET",
	//	function.HandleContractQuery,
	//},
	//Route{
	//	"/VoteProducer",
	//	"Post",
	//	function.HandleVoteTransaction,
	//},
}


