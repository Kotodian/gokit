package elasticsearch

import (
	"context"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/silenceper/pool"
)

var p pool.Pool

func init() {
	//factory 创建连接的方法
	factory := func() (interface{}, error) {
		return elastic.NewClient(elastic.SetSniff(false), elastic.SetHealthcheck(false), elastic.SetURL("http://elasticsearch-master:9200"))
	}

	//close 关闭链接的方法
	_close := func(v interface{}) error {
		return nil
	}

	//创建一个连接池： 初始化5，最大链接30
	poolConfig := &pool.Config{
		InitialCap: 5,
		MaxIdle:    10,
		MaxCap:     65535,
		Factory:    factory,
		Close:      _close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
	}
	// ch = make(chan struct{}, 1)

	// go func() {
	var err error
	p, err = pool.NewChannelPool(poolConfig)
	if err != nil {
		panic(err)
	}
}

func Conn() (*elastic.Client, error) {
	c, err := p.Get()
	if err != nil {
		return nil, err
	}
	return c.(*elastic.Client), nil
}

func Close(c interface{}) error {
	return p.Put(c)
}

func Update(ctx context.Context, index string, _type string, id string, fields map[string]interface{}, defaultValue ...map[string]interface{}) error {
	es, err := Conn()
	if err != nil {
		return err
	}
	defer Close(es)

	//
	//scriptArr := []string{}
	//for k, v := range fields {
	//	if _, ok := v.(string); ok {
	//		scriptArr = append(scriptArr, fmt.Sprintf("ctx._source.%s='%v'", k, v))
	//	} else {
	//		scriptArr = append(scriptArr, fmt.Sprintf("ctx._source.%s=%v", k, v))
	//	}
	//}
	//
	//fmt.Println(scriptArr)
	//_, err = es.Update().Index(index).Type(_type).Id(id).Script(
	//	elastic.NewScriptInline(strings.Join(scriptArr, ";")),
	//).Upsert(defaultValue).Do(ctx)
	//
	up := es.Update().Index(index).Type(_type).Id(id).Doc(fields)
	if len(defaultValue) > 0 {
		up.Upsert(defaultValue[0])
	}
	if _, err = up.Do(ctx); err != nil {
		return err
	}
	_, err = es.Refresh(index).Do(ctx)
	if err != nil {
		return err
	}
	return err
}
