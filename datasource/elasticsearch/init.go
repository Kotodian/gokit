package elasticsearch

import (
	"context"
	"reflect"
	"time"

	"github.com/olivere/elastic/v7"
	"go.uber.org/zap/zapcore"
)

var client *elastic.Client
var mapping = `
	{
		"settings":{
			"number_of_shards":2,
			"number_of_replicas":1,
			"max_result_window": 100000000
		},
		"mappings": {
			"properties": {
				"message": {
					"type": "text"
				}
			}
		}
	}
`

func Init() {
	var err error
	client, err = elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true),
		elastic.SetHealthcheckTimeout(5*time.Second),
		elastic.SetURL("http://elasticsearch-master:9200"),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewSimpleBackoff(200))),
	)
	if err != nil {
		panic(err)
	}
}

type ElasticsearchHook struct {
	index string
}

func NewElasticHook(index string) *ElasticsearchHook {
	return &ElasticsearchHook{index: index}
}

func ZapEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "severity",
		TimeKey:      "@timestamp",
		EncodeTime:   ZapTimeEncoder,
		CallerKey:    "logger",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
}

const RFC3339Mili = "2006-01-02T15:04:05.999Z07:00"

func ZapTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(RFC3339Mili))
}

func (r *ElasticsearchHook) Write(data []byte) (n int, err error) {
	ctx := context.Background()
	index := r.index + "-" + time.Now().Format("2006-01-02")
	indexService := client.Index().Index(index).BodyString(string(data))
	_, err = indexService.Do(ctx)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func IndexCreate(ctx context.Context, index string) error {
	_, err := client.CreateIndex(index).Body(mapping).Do(ctx)
	if err != nil {
		return err
	}
	return err
}

func IndexExists(ctx context.Context, index string) (bool, error) {
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func Insert(ctx context.Context, id, index string, body interface{}) error {
	indexService := client.Index().Index(index).BodyJson(body)
	if id != "" {
		indexService.Id(id)
	}
	_, err := indexService.Do(ctx)
	if err != nil {
		return err
	}
	_, err = client.Refresh(index).Do(ctx)
	return err
}

func InsertBulk(ctx context.Context, index string, body []interface{}) error {
	request := client.Bulk()
	for _, v := range body {
		tmp := v
		req := elastic.NewBulkIndexRequest().Index(index).Doc(tmp)
		request = request.Add(req)
	}
	_, err := request.Do(ctx)
	return err
}

func Update(ctx context.Context, index string, id string, fields map[string]interface{}, defaultValue ...map[string]interface{}) error {
	up := client.Update().Index(index).Id(id).Doc(fields)
	if len(defaultValue) > 0 {
		up.Upsert(defaultValue[0])
	}
	if _, err := up.Do(ctx); err != nil {
		return err
	}
	_, err := client.Refresh(index).Do(ctx)
	if err != nil {
		return err
	}
	return err
}

func UpdateBulk(ctx context.Context, index string, body []map[string]string) error {
	bulkRequest := client.Bulk()
	for _, v := range body {
		tmp := v
		tmps := copyMap(tmp)
		delete(tmps, "id")
		req := elastic.NewBulkUpdateRequest().Index(index).Id(tmp["id"]).Doc(tmps)
		bulkRequest = bulkRequest.Add(req)
	}
	_, err := bulkRequest.Do(ctx)
	return err
}

func Delete(ctx context.Context, index, id string) error {
	_, err := client.Delete().Index(index).Id(id).Do(ctx)
	if err != nil {
		return err
	}
	_, err = client.Refresh(index).Do(ctx)
	return err
}
func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string)
	for is, vs := range m {
		cp[is] = vs
	}
	return cp
}

func UpdateByQuery(ctx context.Context, index string, params, query map[string]interface{}) error {
	script := buildScript(params)
	updateByQueryService := client.UpdateByQuery(index).Script(elastic.NewScript(script).Params(params))
	for k, v := range query {
		if reflect.TypeOf(v).Kind() == reflect.String {
			k += ".keyword"
		}
		updateByQueryService.Query(elastic.NewTermQuery(k, v))
	}
	_, err := updateByQueryService.Do(ctx)
	if err != nil {
		return err
	}
	_, err = client.Refresh(index).Do(ctx)
	return err
}

func DocExist(ctx context.Context, index string, id string) (bool, error) {
	get, err := client.Get().Index(index).Id(id).Do(ctx)
	if err != nil {
		return false, err
	}
	return get.Found, nil
}

const scriptPrefix = "ctx._source."
const paramPrefix = "params."

func buildScript(params map[string]interface{}) string {
	var script string
	for k, _ := range params {
		script += scriptPrefix + k + "=" + paramPrefix + k + ";"
	}
	return script
}
