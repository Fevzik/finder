package discovery

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	ctxTimeout = 1 * time.Second
)

type Service struct {
	Name string
	IP   string
	Port string
}

func getClient(hosts []string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		log.Fatalln(err.Error())
		return nil, err
	}
	return cli, nil
}

func RegisterService(hosts []string, s *Service) error {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return err
	}
	kv := clientv3.NewKV(cli)
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	_, err = kv.Put(ctx, "/services/"+s.Name+"/"+s.IP+":"+s.Port, strconv.FormatInt(time.Now().Unix(), 10))
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func UnregisterService(hosts []string, s *Service) error {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	_, err = cli.Delete(ctx, "/services/"+s.Name+"/"+s.IP+":"+s.Port)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func GetService(hosts []string, sName string) (*Service, error) {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	rangeResp, err := cli.Get(ctx, "/services/"+sName+"/", clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	var servs []Service
	for _, v := range rangeResp.Kvs {
		u := strings.Split(string(v.Key), "/")
		conn := strings.Split(u[len(u)-1], ":")
		s := Service{
			Name: sName,
			IP:   conn[0],
			Port: conn[1],
		}
		servs = append(servs, s)
	}
	if len(servs) > 1 {
		return &servs[rand.Intn(len(servs)-1)], nil
	}
	if len(servs) == 1 {
		return &servs[0], nil
	}
	return nil, nil
}

func GetRegisteredModules(hosts []string) (DiscoveryModulesList, error) {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	rangeResp, err := cli.Get(ctx, "/modules")
	cancel()
	if err != nil {
		return nil, err
	}
	mods := DiscoveryModulesList{}
	if len(rangeResp.Kvs) > 0 {
		err = json.Unmarshal(rangeResp.Kvs[0].Value, &mods)
		if err != nil {
			return mods, err
		}
	}
	return mods, nil
}

func RegisterModule(hosts []string, mod *DiscoveryModule) error {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return err
	}
	mods, err := GetRegisteredModules(hosts)
	if err != nil {
		return err
	}
	exists := false
	for _, v := range mods {
		if v.Code == mod.Code {
			exists = true
		}
	}
	if !exists {
		mods = append(mods, *mod)
		dataBytes, err := json.Marshal(mods)
		if err != nil {
			return err
		}
		kv := clientv3.NewKV(cli)
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		_, err = kv.Put(ctx, "/modules", string(dataBytes))
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func UnRegisterModule(hosts []string, mod *DiscoveryModule) error {
	cli, err := getClient(hosts)
	defer cli.Close()
	if err != nil {
		return err
	}
	mods, err := GetRegisteredModules(hosts)
	if err != nil {
		return err
	}
	for k, v := range mods {
		if v.Code == mod.Code {
			mods = append(mods[:k], mods[k+1:]...)
			dataBytes, err := json.Marshal(mods)
			if err != nil {
				return err
			}
			kv := clientv3.NewKV(cli)
			ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
			_, err = kv.Put(ctx, "/modules", string(dataBytes))
			cancel()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
