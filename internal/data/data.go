// Package data 提供数据访问层，持有 DB/Redis 连接，并统一管理资源生命周期。
//
// 本包各文件职责：
//   - data.go         — Data 容器结构体，New() 构造与 cleanup
//   - user_repo.go    — UserRepo（用户表）+ TokenBlacklistRepo（Redis 黑名单）
//   - message_repo.go — MessageRepo（留言表）
//   - comment_repo.go — CommentRepo（评论表）
package data

import (
	"github.com/jiujuan/go-ants/pkg/database"
	"github.com/jiujuan/go-ants/pkg/log"
	pkgredis "github.com/jiujuan/go-ants/pkg/redis"
)

// Data 数据层容器，持有 DB 和 Redis 客户端，被各 Repo 共享。
type Data struct {
	db    *database.DB
	redis *pkgredis.Client
	log   *log.Logger
}

// New 创建 Data 实例，同时返回资源释放函数。
// 调用方应在程序退出时调用 cleanup 关闭连接。
func New(db *database.DB, redisClient *pkgredis.Client) (*Data, func(), error) {
	d := &Data{
		db:    db,
		redis: redisClient,
		log:   log.DefaultLogger(),
	}
	cleanup := func() {
		if d.db != nil {
			if err := d.db.Close(); err != nil {
				d.log.Errorw("close db failed", "err", err)
			}
		}
		if d.redis != nil {
			if err := d.redis.Close(); err != nil {
				d.log.Errorw("close redis failed", "err", err)
			}
		}
	}
	return d, cleanup, nil
}
