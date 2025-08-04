package main

//
//import (
//	"context"
//	"fmt"
//	"github.com/coreos/etcd/clientv3"
//	"time"
//)
//
//var cli *clientv3.Client
//
//func init() {
//	var err error
//	cli, err = clientv3.New(clientv3.Config{
//		Endpoints: []string{"127.0.0.1:2379", "127.0.0.1:3379", "127.0.0.1:4379"},
//		DialTimeout: 5 * time.Second,
//	})
//	if err != nil {
//		fmt.Printf("Connect to etcd server failed, err:%v", err)
//	}
//	fmt.Println("Connect to etcd success")
//}
//
//func GetContextWithTimeout(second int64) context.Context {
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(second) * time.Second)
//	return ctx
//}
//
//func SetEtcdValue(key string, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error){
//	return cli.Put(GetContextWithTimeout(5), key, value, opts...)
//}
//
//func GetEtcdValue(key string, opts ...clientv3.OpOption)(*clientv3.GetResponse, error){
//	return cli.Get(GetContextWithTimeout(5), key, opts...)
//}
//
//func GrantLease(expireTime int64) (*clientv3.LeaseGrantResponse, error) {
//	return cli.Grant(context.TODO(), expireTime) // 租期
//}
//
//func KeepAlive(id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
//	return cli.KeepAlive(context.TODO(), id)
//}
//
//func DeleteEtcdKey(key string) (*clientv3.DeleteResponse,error){
//	return cli.Delete(context.TODO(), key)
//}
//
//func TestGrantLease() {
//	// 创建3秒的租约
//	g, err := GrantLease(3)
//	if err != nil {
//		fmt.Printf("GrantLease failed, err:%v", err)
//		return
//	}
//	//put,3s后key shane会被删除
//	_, err = SetEtcdValue("dong", "xiang", clientv3.WithLease(g.ID))
//	if err != nil {
//		fmt.Printf("Put to etcd failed, err:%v", err)
//		return
//	}
//
//	//get
//	res1, err := GetEtcdValue("dong")
//	if err != nil {
//		fmt.Printf("Get value from etcd failed, err:%v", err)
//		return
//	}
//	fmt.Println("kv1:", res1.Kvs)
//	for _, v := range res1.Kvs {
//		fmt.Printf("key=%s,value=%s\n", v.Key, v.Value)
//	}
//
//	time.Sleep(4 * time.Second)
//	res2, err := GetEtcdValue("dong")
//	if err != nil {
//		fmt.Printf("Get value from etcd failed, err:%v", err)
//		return
//	}
//	fmt.Println("kv2:", res2.Kvs)
//	for _, v := range res2.Kvs {
//		fmt.Printf("key=%s,value=%s\n", v.Key, v.Value)
//	}
//
//	//Connect to etcd success
//	//kv1: [key:"dong" create_revision:7 mod_revision:7 version:1 value:"xiang" lease:8488306318722885387 ]
//	//key=dong,value=xiang
//	//kv2: []
//}
//
//
//func TestKeepAlive() {
//	// 创建3秒的租约
//	g, err := GrantLease(3)
//	if err != nil {
//		fmt.Printf("GrantLease failed, err:%v", err)
//		return
//	}
//	//put,3s后key shane会被删除
//	_, err = SetEtcdValue("xiang", "xiang", clientv3.WithLease(g.ID))
//	if err != nil {
//		fmt.Printf("Put to etcd failed, err:%v", err)
//		return
//	}
//
//	ch, kerr := KeepAlive(g.ID)
//	if kerr != nil {
//		fmt.Printf("Set keepAlived failed. err:%v", err)
//	}
//	for {
//		ka := <-ch
//		fmt.Println("ttl", ka.TTL)
//	}
//}
//
//func TestDelete() {
//	res1, err := GetEtcdValue("shane")
//	if err != nil {
//		fmt.Printf("Get value from etcd failed, err:%v", err)
//		return
//	}
//	fmt.Println("kv1:", res1.Kvs)
//	for _, v := range res1.Kvs {
//		fmt.Printf("key=%s,value=%s\n", v.Key, v.Value)
//	}
//	DeleteEtcdKey("shane")
//
//	res2, err := GetEtcdValue("shane")
//	if err != nil {
//		fmt.Printf("Get value from etcd failed, err:%v", err)
//		return
//	}
//	fmt.Println("kv1:", res2.Kvs)
//
//	//Connect to etcd success
//	//kv1: [key:"shane" create_revision:2 mod_revision:4 version:3 value:"dong" ]
//	//key=shane,value=dong
//	//kv1: []
//}
//
//func main() {
//	//res2, err := GetEtcdValue("shane")
//	//if err != nil {
//	//	fmt.Printf("Get value from etcd failed, err:%v", err)
//	//	return
//	//}
//	//fmt.Println("kv2:", res2.Kvs)
//	//for _, v := range res2.Kvs {
//	//	fmt.Printf("key=%s,value=%s\n", v.Key, v.Value)
//	//}
//	//TestKeepAlive()
//
//	TestDelete()
//}
//
//
