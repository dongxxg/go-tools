package main

//
//func main() {
//	ctx := kctx.NewDefaultContext()
//	getter := conf.NewKitConf(viper.GetViper(), conf.WithNacosSync(true))
//	os.Setenv("UNIBASE_CONFIG_PATH", "../../config")
//	cli, err := client.EsSendRepoInit(ctx, getter, "Pub.ES",
//		func(ctx kctx.Ctx, infos []client.EsSendCap, esErr error, res map[string]interface{}) {
//		ctx.Info(esErr)
//	})
//	if err != nil {
//		ctx.Fatal(err)
//	}
//	if err := cli.Ping(); err != nil {
//		ctx.Fatal(err)
//	}
//
//	startup.Bootstrap(&startup.BootConfig{
//		ServiceName: "demo",
//		APIs:        nil,
//	})
//
//}
