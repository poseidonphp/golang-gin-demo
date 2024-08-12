package initializers

import (
	"github.com/gin-contrib/sessions/redis"
	"log"
	"os"
)

// We will be using Redis for our Memory DB

var KvStore redis.Store

func ConnectMemoryDb() {

	redisUrl := os.Getenv("REDIS_URL")
	KvStore, _ = redis.NewStore(10, "tcp", redisUrl, "", []byte("secret"))
	log.Println("Connected to Redis")
}
