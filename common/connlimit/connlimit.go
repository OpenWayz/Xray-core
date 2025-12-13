package connlimit

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	userConnCount sync.Map // key: email, value: *atomic.Int32
	maxConn int32 = 500    // 固定上限
)

func init() {
	log.Println("[Trojan] connlimit.go initialized")
	go cleanupInactiveUsers()
}

// incConn: 增加计数并判断是否超过限制
func IncConncustom(email string) bool {
	v, _ := userConnCount.LoadOrStore(email, new(atomic.Int32))
	counter := v.(*atomic.Int32)
	n := counter.Add(1)
	if n > maxConn {
		counter.Add(-1)
		return false
	}
	return true
}

// decConn: 连接结束时递减计数
func DecConncustom(email string) {
	if v, ok := userConnCount.Load(email); ok {
		c := v.(*atomic.Int32)
		if c.Add(-1) <= 0 {
			// 延迟清理交给后台协程
		}
	}
}

// cleanupInactiveUsers: 定期清理空用户记录
func cleanupInactiveUsers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		userConnCount.Range(func(key, value any) bool {
			if value.(*atomic.Int32).Load() <= 0 {
				userConnCount.Delete(key)
			}
			return true
		})
	}
}
