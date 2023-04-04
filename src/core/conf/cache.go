package conf

func init() {
	// TODO ...
	//redisStore := redis_store.NewRedis(redis.NewClient(&redis.Options{
	//	Addr: "127.0.0.1:6379",
	//}))
	//
	//cacheManager := cache.New[string](redisStore)
	//
	//err := cacheManager.Set(ctx, "my-key", "my-value", store.WithExpiration(15*time.Second))
	//if err != nil {
	//	panic(err)
	//}
	//
	//value, err := cacheManager.Get(ctx, "my-key")
	//switch err {
	//case nil:
	//	fmt.Printf("Get the key '%s' from the redis cache. Result: %s", "my-key", value)
	//case redis.Nil:
	//	fmt.Printf("Failed to find the key '%s' from the redis cache.", "my-key")
	//default:
	//	fmt.Printf("Failed to get the value from the redis cache with key '%s': %v", "my-key", err)
	//}
}
