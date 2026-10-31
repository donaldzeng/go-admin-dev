package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/v2/logger"
	"github.com/go-admin-team/go-admin-core/v2/sdk"
)

// TenantDb 由宿主应用（mes-iot-api）在启动时注入，接管数据库中间件的
// 租户解析（token claim > X-Tenant 头 > Host，实现见 mes/middleware 包）。
// 本框架只保留这个接缝，保持与上游最小差异；为 nil 时行为与上游一致——
// 按请求 Host 查库（通配 "*" 由 runtime 短路）。
var TenantDb func(c *gin.Context)

func WithContextDb(c *gin.Context) {
	if TenantDb != nil {
		TenantDb(c)
		return
	}
	db := sdk.Runtime.GetDbByTenant(c.Request.Host)
	if db == nil {
		// 数据库是按 host 分租户注册的。取不到只有两种可能：
		//   1. databases.<host> 里的 key 与实际访问的 Host（含端口）不一致；
		//   2. 连接根本没建立（database 段缺失或初始化失败）。
		// 单库写法（database: 段）由框架自动补成通配 "*"，任何 Host 都命中，
		// 不会走到这里。
		//
		// 这里必须自己拦住：本中间件注册在 CustomError 之前，panic 会逃出
		// CustomError 的 recover，被 net/http 兜住后直接关连接，客户端只看到
		// "Empty reply from server"，日志里除了一段 goroutine 栈什么都没有。
		log.Errorf("no database for host %q: check the database/databases section of the config", c.Request.Host)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "no database configured for host " + c.Request.Host,
		})
		return
	}
	c.Set("db", db.WithContext(c.Request.Context()))
	c.Next()
}
