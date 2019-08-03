package pipeline

import (
	"context"
	"encoding/json"
	"github.com/betterde/ects/config"
	"github.com/betterde/ects/internal/discover"
	"github.com/betterde/ects/models"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"log"
	"time"
)

var Pipelines map[string]*models.Pipeline

func WatchPipelines(local string) {
	Pipelines = make(map[string]*models.Pipeline)
	var curRevision int64 = 0

	for {
		rangeResp, err := discover.Client.Get(context.TODO(), config.Conf.Etcd.Pipeline, clientv3.WithPrefix())

		if err != nil {
			continue
		}
		curRevision = rangeResp.Header.Revision + 1
		break
	}

	watchChan := discover.Client.Watch(context.TODO(), config.Conf.Etcd.Pipeline, clientv3.WithPrefix(), clientv3.WithRev(curRevision), clientv3.WithPrevKV())
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			var pipeline models.Pipeline
			switch event.Type {
			case mvccpb.PUT:
				log.Printf("%s", event.Kv.Value)
				if err := json.Unmarshal(event.Kv.Value, &pipeline); err != nil {
					log.Println(err)
				}

				for _, node := range pipeline.Nodes {
					if node == local {
						Pipelines[pipeline.Id] = &pipeline
					}
				}
			case mvccpb.DELETE:
				if err := json.Unmarshal(event.PrevKv.Value, &pipeline); err != nil {
					log.Println(err)
				}
				delete(Pipelines, pipeline.Id)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func WatchKiller() {
	var curRevision int64 = 0

	for {
		rangeResp, err := discover.Client.Get(context.TODO(), config.Conf.Etcd.Pipeline, clientv3.WithPrefix())

		if err != nil {
			continue
		}
		curRevision = rangeResp.Header.Revision + 1
		break
	}

	watchChan := discover.Client.Watch(context.TODO(), "", clientv3.WithPrefix(), clientv3.WithRev(curRevision))
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			var pipeline models.Pipeline
			if err := json.Unmarshal(event.Kv.Value, &pipeline); err != nil {
				log.Println(err)
			}

			switch event.Type {
			case mvccpb.PUT:
				// TODO 添加或修改本地 Pipeline 属性
				log.Printf("节点：%s 注册成功", pipeline.Id)
			case mvccpb.DELETE:
				// TODO 删除本地 Pipeline
				log.Printf("Pipeline：%s 离线", pipeline.Id)
			}
		}
	}
}